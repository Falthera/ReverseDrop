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

	"github.com/godbus/dbus/v5"
)

type bluezScanner struct {
	mu        sync.Mutex
	running   bool
	conn      *dbus.Conn
	ch        chan Advertisement
	cancel    context.CancelFunc
}

func NewBlueZScanner() Scanner {
	return &bluezScanner{ch: make(chan Advertisement, 64)}
}

func (b *bluezScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return b.ch, nil
	}
	b.mu.Unlock()

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}
	b.conn = conn

	scanCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
	); err != nil {
		conn.Close()
		cancel()
		return nil, err
	}

	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	go b.signalLoop(scanCtx)
	go b.startDiscovery(scanCtx)

	return b.ch, nil
}

func (b *bluezScanner) startDiscovery(ctx context.Context) {
	obj := b.conn.Object("org.bluez", "/")
	call := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0)
	if call.Err != nil {
		return
	}
	_ = call.Store
}

func (b *bluezScanner) signalLoop(ctx context.Context) {
	defer close(b.ch)
	signals := make(chan *dbus.Signal, 256)
	b.conn.Signal(signals)
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-signals:
			if !ok {
				return
			}
		}
	}
}

func (b *bluezScanner) Stop() error {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.conn != nil {
		b.conn.Close()
	}
	b.mu.Lock()
	b.running = false
	b.mu.Unlock()
	return nil
}
