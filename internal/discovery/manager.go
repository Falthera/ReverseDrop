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
	"sync"
	"time"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
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
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		scanner:     scanner,
		capabilities: caps,
		registry:    registry,
		ctx:         ctx,
		cancel:      cancel,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

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

	go m.runScan()
	return nil
}

func (m *Manager) runScan() {
	ch, err := m.scanner.Scan(m.ctx)
	if err != nil {
		m.registry.Publish(app.Event{
			Type:  app.EventTypeScanError,
			Error: err,
		})
		return
	}
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
}
