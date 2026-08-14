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
package peer

import (
	"sync"
	"testing"
	"time"
)

func TestPeerRegistryUpsertAndGet(t *testing.T) {
	reg := NewRegistry()
	p := Peer{Address: "AA:BB:CC:DD:EE:FF", DeviceName: "Test", State: StateDiscovered}
	reg.Upsert(p)

	got, ok := reg.Get("AA:BB:CC:DD:EE:FF")
	if !ok {
		t.Fatal("expected peer to exist")
	}
	if got.DeviceName != "Test" {
		t.Fatalf("expected DeviceName Test, got %s", got.DeviceName)
	}
}

func TestPeerRegistryRemove(t *testing.T) {
	reg := NewRegistry()
	reg.Upsert(Peer{Address: "AA:BB:CC:DD:EE:FF", State: StateDiscovered})
	reg.Remove("AA:BB:CC:DD:EE:FF")

	if _, ok := reg.Get("AA:BB:CC:DD:EE:FF"); ok {
		t.Fatal("expected peer to be removed")
	}
}

func TestPeerRegistryUpdate(t *testing.T) {
	reg := NewRegistry()
	reg.Upsert(Peer{Address: "AA:BB:CC:DD:EE:FF", RSSI: -60, State: StateDiscovered})
	reg.Update("AA:BB:CC:DD:EE:FF", func(p *Peer) {
		p.RSSI = -40
	})

	got, _ := reg.Get("AA:BB:CC:DD:EE:FF")
	if got.RSSI != -40 {
		t.Fatalf("expected RSSI -40, got %d", got.RSSI)
	}
}

func TestPeerRegistryTransition(t *testing.T) {
	reg := NewRegistry()
	reg.Upsert(Peer{Address: "AA:BB:CC:DD:EE:FF", State: StateDiscovered})

	if !reg.Transition("AA:BB:CC:DD:EE:FF", StateConnecting) {
		t.Fatal("expected valid transition")
	}
	got, _ := reg.Get("AA:BB:CC:DD:EE:FF")
	if got.State != StateConnecting {
		t.Fatalf("expected StateConnecting, got %s", got.State)
	}

	if reg.Transition("AA:BB:CC:DD:EE:FF", StateUnknown) {
		t.Fatal("expected invalid transition to fail")
	}
}

func TestPeerRegistryConcurrent(t *testing.T) {
	reg := NewRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			addr := "AA:BB:CC:DD:EE:%02X"
			reg.Upsert(Peer{Address: addr, DeviceName: "Device", State: StateDiscovered})
			reg.Update(addr, func(p *Peer) { p.RSSI = -50 })
			reg.Get(addr)
		}(i)
	}
	wg.Wait()
}

func TestValidTransitions(t *testing.T) {
	cases := []struct {
		from, to PeerState
		ok      bool
	}{
		{StateUnknown, StateDiscovered, true},
		{StateDiscovered, StateConnecting, true},
		{StateConnecting, StateHandshaking, true},
		{StateHandshaking, StateTransferring, true},
		{StateTransferring, StateDisconnected, true},
		{StateDisconnected, StateDiscovered, true},
		{StateDiscovered, StateTransferring, false},
		{StateUnknown, StateTransferring, false},
	}
	for _, c := range cases {
		if got := IsValidTransition(c.from, c.to); got != c.ok {
			t.Errorf("IsValidTransition(%s, %s) = %v, want %v", c.from, c.to, got, c.ok)
		}
	}
}

func TestPeerRegistryEvents(t *testing.T) {
	reg := NewRegistry()
	defer reg.Close()

	var gotEvent RegistryEvent
	done := make(chan struct{})
	go func() {
		for evt := range reg.Events() {
			gotEvent = evt
			close(done)
			return
		}
	}()

	reg.Upsert(Peer{Address: "AA:BB:CC:DD:EE:FF", DeviceName: "Test", State: StateDiscovered})
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}

	if gotEvent.Type != EventPeerUpserted {
		t.Fatalf("expected EventPeerUpserted, got %s", gotEvent.Type)
	}
	if gotEvent.Peer.DeviceName != "Test" {
		t.Fatalf("expected device name Test, got %s", gotEvent.Peer.DeviceName)
	}
}
