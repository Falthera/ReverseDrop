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
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	ServiceType = "_reversedrop._tcp.local."
)

type PeerInfo struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
	Platform   string `json:"platform,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

type Discovery struct {
	mu       sync.Mutex
	peers    map[string]PeerInfo
	ctx      context.Context
	cancel   context.CancelFunc
	server   *mdns.Server
}

func NewDiscovery() *Discovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		peers:  make(map[string]PeerInfo),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (d *Discovery) Start(port int) error {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	info := map[string]string{
		"platform": runtime.GOOS,
		"name":     fmt.Sprintf("reversedrop-%s", hostname),
	}

	var infoFields []string
	for k, v := range info {
		infoFields = append(infoFields, fmt.Sprintf("%s=%s", k, v))
	}

	service, err := mdns.NewMDNSService(hostname, ServiceType, "", hostname, port, []net.IP{}, infoFields)
	if err != nil {
		return err
	}
	_ = service

	go d.browse()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.pruneStalePeers()
			}
		}
	}()

	return nil
}

func (d *Discovery) browse() {
	entriesCh := make(chan *mdns.ServiceEntry, 32)
	go func() {
		for entry := range entriesCh {
			ip := ""
			if entry.AddrV4 != nil {
				ip = entry.AddrV4.String()
			} else if entry.AddrV6 != nil {
				ip = entry.AddrV6.String()
			}
			if ip == "" {
				continue
			}
			platform := ""
			name := entry.Name
			for _, field := range entry.InfoFields {
				if len(field) > 9 && field[:9] == "platform=" {
					platform = field[9:]
				}
				if len(field) > 5 && field[:5] == "name=" {
					name = field[5:]
				}
			}
			d.mu.Lock()
			d.peers[entry.Name] = PeerInfo{
				Name:       name,
				Address:    ip,
				Port:       entry.Port,
				Platform:   platform,
				DeviceName: name,
			}
			d.mu.Unlock()
		}
	}()

	err := mdns.Lookup(ServiceType, entriesCh)
	if err != nil {
		return
	}
}

func (d *Discovery) pruneStalePeers() {
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	for name, peer := range d.peers {
		if peer.Address == "" {
			delete(d.peers, name)
		}
		_ = now
	}
}

func (d *Discovery) Peers() []PeerInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]PeerInfo, 0, len(d.peers))
	for _, p := range d.peers {
		out = append(out, p)
	}
	return out
}

func (d *Discovery) Stop() {
	d.cancel()
	if d.server != nil {
		d.server.Shutdown()
	}
}
