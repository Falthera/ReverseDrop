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

type fakeScanner struct {
	mu       sync.Mutex
	running  bool
	interval time.Duration
	events   []Advertisement
	idx      int
}

func NewFakeScanner(events []Advertisement, interval time.Duration) Scanner {
	return &fakeScanner{events: events, interval: interval}
}

func (f *fakeScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return nil, context.Canceled
	}
	f.running = true
	f.idx = 0
	ch := make(chan Advertisement, 16)
	f.mu.Unlock()

	go func() {
		defer close(ch)
		ticker := time.NewTicker(f.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				f.mu.Lock()
				if f.idx < len(f.events) {
					adv := f.events[f.idx]
					f.idx++
					f.mu.Unlock()
					select {
					case ch <- adv:
					case <-ctx.Done():
						return
					}
				} else {
					f.mu.Unlock()
				}
			}
		}
	}()
	return ch, nil
}

func (f *fakeScanner) Stop() error {
	f.mu.Lock()
	f.running = false
	f.mu.Unlock()
	return nil
}
