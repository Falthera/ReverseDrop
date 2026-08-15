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

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

type Service struct {
	mu         sync.RWMutex
	subscribers map[Subscriber]struct{}
	registry   *PeerRegistryAdapter
	caps       *CapabilitySet
	scanner    ble.Scanner
}

type PeerRegistryAdapter struct {
	peer.PeerRegistry
}

func NewPeerRegistryAdapter(pr peer.PeerRegistry) *PeerRegistryAdapter {
	return &PeerRegistryAdapter{PeerRegistry: pr}
}

func (a *PeerRegistryAdapter) Publish(evt Event) {
	if a == nil || a.PeerRegistry == nil {
		return
	}
}

func NewService(scanner ble.Scanner) (*Service, error) {
	reg := peer.NewRegistry()
	caps := NewCapabilitySet()
	return &Service{
		subscribers: make(map[Subscriber]struct{}),
		registry:   NewPeerRegistryAdapter(reg),
		caps:       caps,
		scanner:    scanner,
	}, nil
}

func (s *Service) Subscribe(sub Subscriber) {
	if s == nil || sub == nil {
		return
	}
	s.mu.Lock()
	s.subscribers[sub] = struct{}{}
	s.mu.Unlock()
}

func (s *Service) Unsubscribe(sub Subscriber) {
	if s == nil || sub == nil {
		return
	}
	s.mu.Lock()
	delete(s.subscribers, sub)
	s.mu.Unlock()
}

func (s *Service) Publish(evt Event) {
	if s == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for sub := range s.subscribers {
		sub.OnEvent(evt)
	}
}

func (s *Service) Capabilities() *CapabilitySet {
	if s == nil {
		return NewCapabilitySet()
	}
	return s.caps
}

func (s *Service) Registry() *PeerRegistryAdapter {
	if s == nil {
		return nil
	}
	return s.registry
}
