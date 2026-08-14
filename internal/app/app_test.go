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
package app

import (
	"sync"
	"testing"
)

type testSubscriber struct {
	mu     sync.Mutex
	events []Event
}

func (t *testSubscriber) OnEvent(e Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
}

func TestCapabilitySet(t *testing.T) {
	caps := NewCapabilitySet()
	if caps == nil {
		t.Fatal("expected non-nil CapabilitySet")
	}
	caps.Set(CapabilityInfo{Name: CapabilityBluetooth, Status: CapabilityAvailable, Detail: "ok"})
	info, ok := caps.Get(CapabilityBluetooth)
	if !ok {
		t.Fatal("expected capability to exist")
	}
	if info.Status != CapabilityAvailable {
		t.Fatalf("expected available, got %s", info.Status)
	}
	list := caps.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(list))
	}
}

func TestServiceSubscribePublish(t *testing.T) {
	sub := &testSubscriber{}
	svc := &Service{subscribers: map[Subscriber]struct{}{sub: {}}}
	svc.Publish(Event{Type: EventTypePeerDiscovered})
	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(sub.events))
	}
	if sub.events[0].Type != EventTypePeerDiscovered {
		t.Fatalf("unexpected event type: %s", sub.events[0].Type)
	}
}

func TestServiceUnsubscribe(t *testing.T) {
	sub := &testSubscriber{}
	svc := &Service{subscribers: map[Subscriber]struct{}{sub: {}}}
	svc.Unsubscribe(sub)
	svc.Publish(Event{Type: EventTypePeerDiscovered})
	sub.mu.Lock()
	defer sub.mu.Unlock()
	if len(sub.events) != 0 {
		t.Fatalf("expected 0 events after unsubscribe, got %d", len(sub.events))
	}
}

func TestPeerRegistryAdapterPublishNoop(t *testing.T) {
	a := NewPeerRegistryAdapter(nil)
	a.Publish(Event{Type: EventTypePeerDiscovered})
}
