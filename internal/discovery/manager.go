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
package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/discovery/network"
	"github.com/Falthera/ReverseDrop/internal/platform"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

type Manager struct {
	mu          sync.RWMutex
	scanner     ble.Scanner
	capabilities *app.CapabilitySet
	capReporter platform.CapabilityReporter
	registry    *app.PeerRegistryAdapter
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

type ManagerOption func(*Manager)

func WithCapabilityReporter(cr platform.CapabilityReporter) ManagerOption {
	return func(m *Manager) {
		m.capReporter = cr
	}
}

func NewManager(scanner ble.Scanner, registry *app.PeerRegistryAdapter, caps *app.CapabilitySet, opts ...ManagerOption) *Manager {
	return &Manager{
		scanner:     scanner,
		capabilities: caps,
		registry:    registry,
	}
}

func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	if m.ctx != nil {
		m.cancel()
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.running = true
	m.mu.Unlock()

	m.registry.Publish(app.Event{
		Type:      app.EventTypeScanStarted,
		Timestamp: time.Now().UnixNano(),
	})

	if m.capReporter != nil {
		if avail, detail := m.capReporter.BluetoothAvailable(); avail {
			m.capabilities.Set(app.CapabilityInfo{
				Name:   app.CapabilityBluetooth,
				Status: app.CapabilityAvailable,
				Detail: detail,
			})
		} else {
			m.capabilities.Set(app.CapabilityInfo{
				Name:   app.CapabilityBluetooth,
				Status: app.CapabilityUnavailable,
				Detail: detail,
			})
		}
	}

	go m.runScanWithBackoff()
	return nil
}

func (m *Manager) runScanWithBackoff() {
	backoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}
		ch, err := m.scanner.Scan(m.ctx)
		if err != nil {
			m.registry.Publish(app.Event{
				Type:  app.EventTypeScanError,
				Error: err,
			})
			select {
			case <-time.After(backoff):
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			case <-m.ctx.Done():
				return
			}
		}
		backoff = 100 * time.Millisecond
		m.drainScan(ch)
	}
}

func (m *Manager) drainScan(ch <-chan ble.Advertisement) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case adv, ok := <-ch:
			if !ok {
				return
			}
			m.handleAdv(adv)
		}
	}
}

func (m *Manager) handleAdv(adv ble.Advertisement) {
	p := peer.Peer{
		Address:    adv.Address,
		RSSI:       adv.RSSI,
		DeviceName: adv.LocalName,
		LastSeen:   time.Now(),
		State:      peer.StateDiscovered,
	}
	m.registry.Upsert(p)
	m.registry.Publish(app.Event{
		Type: app.EventTypePeerDiscovered,
		Peer: &p,
	})
}

func (m *Manager) Stop() {
	m.cancel()
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.registry.Publish(app.Event{
		Type:      app.EventTypeScanStopped,
		Timestamp: time.Now().UnixNano(),
	})
}

type networkScannerAdapter struct {
	disc *network.Discovery
	port int
}

func (a *networkScannerAdapter) Scan(ctx context.Context) (<-chan ble.Advertisement, error) {
	if err := a.disc.Start(a.port); err != nil {
		return nil, fmt.Errorf("failed to start network discovery: %w", err)
	}
	out := make(chan ble.Advertisement, 32)
	go func() {
		defer close(out)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, p := range a.disc.Peers() {
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

func (a *networkScannerAdapter) Stop() error {
	a.disc.Stop()
	return nil
}

type ScannerAdapter struct {
	mu         sync.RWMutex
	ble        ble.Scanner
	netScanner ble.Scanner
	cancel     context.CancelFunc
}

func NewScannerAdapter(bleScanner ble.Scanner, port int) *ScannerAdapter {
	_, cancel := context.WithCancel(context.Background())
	return &ScannerAdapter{
		ble:        bleScanner,
		netScanner: &networkScannerAdapter{disc: network.NewDiscovery(), port: port},
		cancel:     cancel,
	}
}

func (a *ScannerAdapter) Scan(ctx context.Context) (<-chan ble.Advertisement, error) {
	ch, err := a.ble.Scan(ctx)
	if err == nil {
		return ch, nil
	}
	return a.netScanner.Scan(ctx)
}

func (a *ScannerAdapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	_ = a.ble.Stop()
	_ = a.netScanner.Stop()
	return nil
}
