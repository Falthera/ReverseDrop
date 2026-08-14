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

import "testing"

func TestTransitionValid(t *testing.T) {
	p := &Peer{Address: "AA:BB:CC:DD:EE:FF", State: StateDiscovered}
	if err := Transition(p, StateConnecting); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.State != StateConnecting {
		t.Fatalf("expected StateConnecting, got %s", p.State)
	}
}

func TestTransitionInvalid(t *testing.T) {
	p := &Peer{Address: "AA:BB:CC:DD:EE:FF", State: StateDiscovered}
	if err := Transition(p, StateTransferring); err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestTransitionNilPeer(t *testing.T) {
	if err := Transition(nil, StateDiscovered); err == nil {
		t.Fatal("expected error for nil peer")
	}
}
