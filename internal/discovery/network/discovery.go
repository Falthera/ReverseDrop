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
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
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
	listener *net.UDPConn
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
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	d.listener = conn
	go d.readLoop()
	go d.announce()
	return nil
}

func (d *Discovery) readLoop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}
		n, _, err := d.listener.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		var info PeerInfo
		if err := json.Unmarshal(buf[:n], &info); err != nil {
			continue
		}
		if info.Name == "" {
			continue
		}
		d.mu.Lock()
		d.peers[info.Name] = info
		d.mu.Unlock()
	}
}

func (d *Discovery) announce() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.broadcast()
		}
	}
}

func (d *Discovery) broadcast() {
	info := PeerInfo{
		Name:    fmt.Sprintf("reversedrop-%s", getHostname()),
		Address: getLocalIP(),
		Port:    0,
	}
	data, _ := json.Marshal(info)
	addr, _ := net.ResolveUDPAddr("udp", "224.0.0.251:9999")
	if addr == nil {
		return
	}
	_, _ = d.listener.WriteTo(data, addr)
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
	if d.listener != nil {
		d.listener.Close()
	}
}

func getHostname() string {
	name, _ := os.Hostname()
	if name == "" {
		return "unknown"
	}
	return name
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			return ip.IP.String()
		}
	}
	return "127.0.0.1"
}
