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
	mu                sync.RWMutex
	scanner           ble.Scanner
	capabilities      *app.CapabilitySet
	capReporter       platform.CapabilityReporter
	registry          *app.PeerRegistryAdapter
	running           bool
	ctx               context.Context
	cancel            context.CancelFunc
	fastScanInterval  time.Duration
	slowScanInterval  time.Duration
	backgroundInterval time.Duration
	background        bool
	lastDiscoveryTime time.Time
}

type ManagerOption func(*Manager)

func WithCapabilityReporter(cr platform.CapabilityReporter) ManagerOption {
	return func(m *Manager) {
		m.capReporter = cr
	}
}

func WithScanIntervals(fast, slow, background time.Duration) ManagerOption {
	return func(m *Manager) {
		if fast > 0 {
			m.fastScanInterval = fast
		}
		if slow > 0 {
			m.slowScanInterval = slow
		}
		if background > 0 {
			m.backgroundInterval = background
		}
	}
}

func NewManager(scanner ble.Scanner, registry *app.PeerRegistryAdapter, caps *app.CapabilitySet, opts ...ManagerOption) *Manager {
	m := &Manager{
		scanner:           scanner,
		capabilities:      caps,
		registry:          registry,
		fastScanInterval:  time.Second,
		slowScanInterval:  5 * time.Second,
		backgroundInterval: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) SetBackground(background bool) {
	m.mu.Lock()
	m.background = background
	m.mu.Unlock()
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
		avail, err := m.capReporter.BluetoothAvailable()
		status := app.CapabilityAvailable
		detail := "Bluetooth is available"
		if err != nil {
			detail = err.Error()
			if bluetoothErr, ok := err.(*platform.BluetoothError); ok {
				switch bluetoothErr.Category {
				case platform.BluetoothErrorPermissionDenied:
					status = app.CapabilityPermissionDenied
				case platform.BluetoothErrorUnavailable:
					status = app.CapabilityUnavailable
				case platform.BluetoothErrorDisabled:
					status = app.CapabilityUnavailable
				default:
					status = app.CapabilityUnavailable
				}
			} else {
				status = app.CapabilityUnavailable
			}
		}
		if !avail && status == app.CapabilityAvailable {
			status = app.CapabilityUnavailable
		}
		m.capabilities.Set(app.CapabilityInfo{
			Name:   app.CapabilityBluetooth,
			Status: status,
			Detail: detail,
		})
	}

	go m.runScanWithBackoff()
	return nil
}

func (m *Manager) getScanInterval() time.Duration {
	m.mu.RLock()
	background := m.background
	lastDiscovery := m.lastDiscoveryTime
	fastInterval := m.fastScanInterval
	slowInterval := m.slowScanInterval
	bgInterval := m.backgroundInterval
	m.mu.RUnlock()

	if background {
		return bgInterval
	}
	if time.Since(lastDiscovery) < 10*time.Second {
		return fastInterval
	}
	return slowInterval
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

		interval := m.getScanInterval()

		scanCtx, cancel := context.WithTimeout(m.ctx, interval)
		ch, err := m.scanner.Scan(scanCtx)
		if err != nil {
			cancel()
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
		m.drainScan(ch, scanCtx.Done())
		cancel()
		_ = m.scanner.Stop()
	}
}

func (m *Manager) drainScan(ch <-chan ble.Advertisement, done <-chan struct{}) {
	for {
		select {
		case <-m.ctx.Done():
			for {
				select {
				case adv, ok := <-ch:
					if !ok {
						return
					}
					m.handleAdv(adv)
				default:
					return
				}
			}
	case <-done:
		for {
			select {
			case adv, ok := <-ch:
				if !ok {
					return
				}
				m.handleAdv(adv)
			default:
				return
			}
		}
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
	m.mu.Lock()
	m.lastDiscoveryTime = time.Now()
	m.mu.Unlock()
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
	disc         *network.Discovery
	port         int
	scanInterval time.Duration
}

func (a *networkScannerAdapter) Scan(ctx context.Context) (<-chan ble.Advertisement, error) {
	if err := a.disc.Start(a.port); err != nil {
		return nil, fmt.Errorf("failed to start network discovery: %w", err)
	}
	out := make(chan ble.Advertisement, 32)
	go func() {
		defer close(out)
		interval := a.scanInterval
		if interval <= 0 {
			interval = 500 * time.Millisecond
		}
		ticker := time.NewTicker(interval)
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

func (a *networkScannerAdapter) SetScanInterval(d time.Duration) {
	a.scanInterval = d
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

func (a *ScannerAdapter) SetScanInterval(d time.Duration) {
	a.ble.SetScanInterval(d)
	a.netScanner.SetScanInterval(d)
}
