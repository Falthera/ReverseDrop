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
package ble

import (
	"context"
	"sync"
	"time"
)

type coreBluetoothScanner struct {
	mu        sync.Mutex
	running   bool
	ch        chan Advertisement
	cancel    context.CancelFunc
	central   *CBPeripheralManager
}

func NewCoreBluetoothScanner() Scanner {
	return &coreBluetoothScanner{ch: make(chan Advertisement, 64)}
}

func (c *coreBluetoothScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return c.ch, nil
	}
	c.mu.Unlock()

	scanCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	go c.scanLoop(scanCtx)

	return c.ch, nil
}

func (c *coreBluetoothScanner) scanLoop(ctx context.Context) {
	defer close(c.ch)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			if !c.running {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

func (c *coreBluetoothScanner) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
	return nil
}

type CBPeripheralManager struct {
	Delegate CBPeripheralManagerDelegate
	Queue    chan struct{}
}

type CBPeripheralManagerDelegate interface {
	DidDiscoverPeripheral(manager *CBPeripheralManager, peripheral CBPeripheral, advertisementData map[string]interface{}, RSSI float64)
}

type CBPeripheral struct {
	Name      string
	UUID      string
	RSSI      float64
	State     string
}

func (c *coreBluetoothScanner) processPeripheral(peripheral CBPeripheral) {
	select {
	case c.ch <- Advertisement{
		Address:     peripheral.UUID,
		LocalName:   peripheral.Name,
		RSSI:        int(peripheral.RSSI),
		Timestamp:   time.Now().Unix(),
	}:
	default:
	}
}
