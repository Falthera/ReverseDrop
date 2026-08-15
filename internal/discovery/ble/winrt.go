//go:build windows

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
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

var (
	winrtDLL              = windows.NewLazyDLL("api-ms-win-core-winrt-l1-1-0.dll")
	winrtRoInitialize      = winrtDLL.NewProc("RoInitialize")
	winrtRoUninitialize    = winrtDLL.NewProc("RoUninitialize")
	winrtActivationFactory = winrtDLL.NewProc("RoGetActivationFactory")
)

const (
	LEAdvertisementPublisherStatusCreated = iota
	LEAdvertisementPublisherStatusWaiting
	LEAdvertisementPublisherStatusStarted
	LEAdvertisementPublisherStatusStopped
	LEAdvertisementPublisherStatusAborted
)

type winrtScanner struct {
	mu        sync.Mutex
	running   bool
	ch        chan Advertisement
	cancel    context.CancelFunc
	publisher uintptr
}

func NewWinRTScanner() Scanner {
	return &winrtScanner{ch: make(chan Advertisement, 64)}
}

func (w *winrtScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return w.ch, nil
	}
	w.mu.Unlock()

	if err := w.roInitialize(); err != nil {
		return nil, fmt.Errorf("RoInitialize failed: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	go w.scanLoop(scanCtx)

	return w.ch, nil
}

func (w *winrtScanner) scanLoop(ctx context.Context) {
	defer close(w.ch)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.Lock()
			if !w.running {
				w.mu.Unlock()
				return
			}
			w.mu.Unlock()
		}
	}
}

func (w *winrtScanner) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Lock()
	w.running = false
	w.mu.Unlock()
	return nil
}

func (w *winrtScanner) roInitialize() error {
	ret, _, _ := winrtRoInitialize.Call(uintptr(0))
	if ret != 0 {
		return fmt.Errorf("HRESULT 0x%08X", ret)
	}
	return nil
}

func (w *winrtScanner) roUninitialize() {
	winrtRoUninitialize.Call()
}

type bleAdvertisement struct {
	Address         string
	LocalName       string
	ManufacturerData map[uint16][]byte
}

func (w *winrtScanner) processAdvertisement(adv bleAdvertisement) {
	select {
	case w.ch <- Advertisement{
		Address:          adv.Address,
		LocalName:        adv.LocalName,
		ManufacturerData: adv.ManufacturerData,
		Timestamp:        time.Now().Unix(),
	}:
	default:
	}
}
