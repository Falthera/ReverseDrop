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
	"testing"
	"time"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

type fakeCapReporter struct {
	available bool
	err       error
	detail    string
}

func (f *fakeCapReporter) BluetoothAvailable() (bool, error)       { return f.available, f.err }
func (f *fakeCapReporter) NetworkDiscoveryAvailable() (bool, string) { return true, "ok" }
func (f *fakeCapReporter) NotificationsAvailable() (bool, string)    { return true, "ok" }

func TestManagerStartStop(t *testing.T) {
	events := []ble.Advertisement{
		{Address: "AA:BB:CC:DD:EE:FF", RSSI: -50, LocalName: "Device", Timestamp: time.Now().Unix()},
	}
	scanner := ble.NewFakeScanner(events, 50*time.Millisecond)
	reg := peer.NewRegistry()
	regAdapter := app.NewPeerRegistryAdapter(reg)
	caps := app.NewCapabilitySet()

	mgr := NewManager(scanner, regAdapter, caps, WithCapabilityReporter(&fakeCapReporter{available: true, err: nil}), WithScanIntervals(50*time.Millisecond, 50*time.Millisecond, 50*time.Millisecond))
	if err := mgr.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	mgr.Stop()

	list := reg.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(list))
	}
	if list[0].Address != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("unexpected peer address: %s", list[0].Address)
	}
}

func TestManagerStartAlreadyStarted(t *testing.T) {
	scanner := ble.NewFakeScanner(nil, time.Hour)
	reg := peer.NewRegistry()
	regAdapter := app.NewPeerRegistryAdapter(reg)
	caps := app.NewCapabilitySet()
	mgr := NewManager(scanner, regAdapter, caps)
	_ = mgr.Start()
	_ = mgr.Start()
	mgr.Stop()
}

func TestManagerContextCancel(t *testing.T) {
	scanner := ble.NewFakeScanner(nil, time.Hour)
	reg := peer.NewRegistry()
	regAdapter := app.NewPeerRegistryAdapter(reg)
	caps := app.NewCapabilitySet()
	mgr := NewManager(scanner, regAdapter, caps)
	_ = mgr.Start()
	mgr.Stop()
}
