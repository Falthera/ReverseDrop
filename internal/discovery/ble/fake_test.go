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
	"testing"
	"time"
)

func TestFakeScannerScan(t *testing.T) {
	events := []Advertisement{
		{Address: "AA:BB:CC:DD:EE:FF", RSSI: -50, LocalName: "Device1"},
		{Address: "11:22:33:44:55:66", RSSI: -60, LocalName: "Device2"},
	}
	scanner := NewFakeScanner(events, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	var count int
	for range ch {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 advertisements, got %d", count)
	}
}

func TestFakeScannerStop(t *testing.T) {
	scanner := NewFakeScanner(nil, time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	ch, err := scanner.Scan(ctx)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	_ = scanner.Stop()
	for range ch {
	}
}

func TestFakeScannerDoubleScan(t *testing.T) {
	scanner := NewFakeScanner(nil, time.Hour)
	ctx := context.Background()

	_, _ = scanner.Scan(ctx)
	_, err := scanner.Scan(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled on double scan, got %v", err)
	}
}
