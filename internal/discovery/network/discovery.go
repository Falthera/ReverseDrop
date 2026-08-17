// ReverseDrop - Copyright (C) ReverseDrop Contributors
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package network

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/discovery/parser"
	"github.com/hashicorp/mdns"
)

const (
	ServiceType       = "_airdrop._tcp.local."
	DefaultPort       = 8770
	DefaultFlags      = 0x3FB

	AirDropSupportsURL         = 0x01
	AirDropSupportsDVZIP       = 0x02
	AirDropSupportsPipelining  = 0x04
	AirDropSupportsMixedTypes  = 0x08
	AirDropSupportsAssetBundle = 0x200

	PeerCacheTTL = 30 * time.Second
	BrowseInterval = 2 * time.Second
	PriorityBrowseService = "_airdrop._tcp.local."
)

type PeerInfo struct {
	Name        string              `json:"name"`
	Address     string              `json:"address"`
	Port        int                 `json:"port"`
	Platform    string              `json:"platform,omitempty"`
	DeviceName  string              `json:"device_name,omitempty"`
	LastSeen    time.Time           `json:"last_seen,omitempty"`
	IsAirDrop   bool                `json:"is_airdrop,omitempty"`
	AirDropInfo *parser.AirDropInfo `json:"airdrop_info,omitempty"`
}

type cachedPeer struct {
	peer   PeerInfo
	expiry time.Time
}

type Discovery struct {
	mu             sync.Mutex
	peers          map[string]cachedPeer
	ctx            context.Context
	cancel         context.CancelFunc
	server         *mdns.Server
	resolvedHosts  map[string]string
}

func NewDiscovery() *Discovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		peers:         make(map[string]cachedPeer),
		ctx:           ctx,
		cancel:        cancel,
		resolvedHosts: make(map[string]string),
	}
}

func (d *Discovery) Start(port int) error {
	if port <= 0 {
		port = DefaultPort
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	slog.Info("network discovery started", "service", ServiceType, "port", port)

	info := generateAirDropTXT(hostname)

	service, err := mdns.NewMDNSService(hostname, ServiceType, "", hostname, port, []net.IP{}, info)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	config := &mdns.Config{
		Zone: service,
	}
	server, err := mdns.NewServer(config)
	if err != nil {
		return fmt.Errorf("failed to create mDNS server: %w", err)
	}
	d.server = server

	go d.browse()
	go d.browsePriority()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.pruneStalePeers()
				d.pruneResolvedHosts()
			}
		}
	}()

	return nil
}

func generateAirDropTXT(hostname string) []string {
	flags := fmt.Sprintf("%X", DefaultFlags)
	emailHash := fmt.Sprintf("%x", md5.Sum([]byte(hostname+"@email")))[:8]
	phoneHash := fmt.Sprintf("%x", md5.Sum([]byte(hostname+"@phone")))[:8]

	return []string{
		fmt.Sprintf("flags=%s", flags),
		fmt.Sprintf("cname=%s", hostname),
		fmt.Sprintf("ehash=%s", emailHash),
		fmt.Sprintf("phash=%s", phoneHash),
		fmt.Sprintf("platform=%s", runtime.GOOS),
	}
}

func (d *Discovery) browse() {
	slog.Debug("network discovery browse started", "service", ServiceType)
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		entriesCh := make(chan *mdns.ServiceEntry, 64)
		go func() {
			for entry := range entriesCh {
				d.handleServiceEntry(entry)
			}
		}()

		err := mdns.Lookup(ServiceType, entriesCh)
		close(entriesCh)
		if err != nil {
			slog.Warn("network discovery lookup failed", "error", err)
		}
		time.Sleep(BrowseInterval)
	}
}

func (d *Discovery) browsePriority() {
	slog.Debug("network discovery priority browse started", "service", PriorityBrowseService)
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}

		entriesCh := make(chan *mdns.ServiceEntry, 64)
		go func() {
			for entry := range entriesCh {
				d.handleServiceEntry(entry)
			}
		}()

		err := mdns.Lookup(PriorityBrowseService, entriesCh)
		close(entriesCh)
		if err != nil {
			slog.Warn("network discovery priority lookup failed", "error", err)
		}
		time.Sleep(BrowseInterval)
	}
}

func (d *Discovery) handleServiceEntry(entry *mdns.ServiceEntry) {
	ip := ""
	if entry.AddrV4 != nil {
		ip = entry.AddrV4.String()
	} else if entry.AddrV6 != nil {
		ip = entry.AddrV6.String()
	}
	if ip == "" {
		return
	}

	airdropInfo := parseAirDropServiceEntry(entry)

	name := entry.Name
	deviceName := name
	platform := ""

	for _, field := range entry.InfoFields {
		if len(field) > 9 && field[:9] == "platform=" {
			platform = field[9:]
		}
		if len(field) > 6 && field[:6] == "cname=" {
			deviceName = field[6:]
		}
	}

	d.mu.Lock()
	d.peers[entry.Name] = cachedPeer{
		peer: PeerInfo{
			Name:        name,
			Address:     ip,
			Port:        entry.Port,
			Platform:    platform,
			DeviceName:  deviceName,
			LastSeen:    time.Now(),
			IsAirDrop:   true,
			AirDropInfo: &airdropInfo,
		},
		expiry: time.Now().Add(PeerCacheTTL),
	}
	d.mu.Unlock()
	slog.Debug("airdrop peer discovered", "name", deviceName, "address", ip, "port", entry.Port)
}

func parseAirDropServiceEntry(entry *mdns.ServiceEntry) parser.AirDropInfo {
	info := parser.AirDropInfo{
		IsAirDrop:      true,
		RawServiceData: make(map[string][]byte),
	}

	for _, field := range entry.InfoFields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		switch key {
		case "flags":
			info.Flags = value
		case "cname":
			info.DeviceName = value
		case "ehash":
			info.EMailHash = value
		case "phash":
			info.PhoneHash = value
		}
		info.RawServiceData[key] = []byte(value)
	}

	if info.DeviceName == "" {
		info.DeviceName = entry.Name
	}

	return info
}

func (d *Discovery) Scan(ctx context.Context) (<-chan ble.Advertisement, error) {
	out := make(chan ble.Advertisement, 32)
	go func() {
		defer close(out)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.mu.Lock()
				now := time.Now()
				peers := make([]PeerInfo, 0, len(d.peers))
				for _, cached := range d.peers {
					if now.Before(cached.expiry) {
						peers = append(peers, cached.peer)
					}
				}
				d.mu.Unlock()
				for _, p := range peers {
					out <- ble.Advertisement{
						Address:    p.Address,
						LocalName:  p.DeviceName,
						RSSI:       -60,
						Timestamp:  time.Now().UnixNano(),
					}
				}
			}
		}
	}()
	return out, nil
}

func (d *Discovery) pruneStalePeers() {
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	for name, cached := range d.peers {
		if now.After(cached.expiry) {
			delete(d.peers, name)
		}
	}
}

func (d *Discovery) pruneResolvedHosts() {
	d.mu.Lock()
	defer d.mu.Unlock()
		for host, _ := range d.resolvedHosts {
		if _, err := net.LookupHost(host); err != nil {
			delete(d.resolvedHosts, host)
		}
	}
}

func (d *Discovery) Peers() []PeerInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	out := make([]PeerInfo, 0, len(d.peers))
	for _, cached := range d.peers {
		if now.Before(cached.expiry) {
			out = append(out, cached.peer)
		}
	}
	return out
}

func (d *Discovery) Stop() {
	d.cancel()
	if d.server != nil {
		d.server.Shutdown()
	}
	slog.Info("network discovery stopped")
}
