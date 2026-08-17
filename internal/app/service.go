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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

type Service struct {
	mu                sync.RWMutex
	subscribers       map[Subscriber]struct{}
	registry          *PeerRegistryAdapter
	caps              *CapabilitySet
	scanner           ble.Scanner
	online            bool
	lastOnlineCheck   time.Time
	onlineCheckTicker *time.Ticker
	offlineCheckDone  chan struct{}
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
	s := &Service{
		subscribers: make(map[Subscriber]struct{}),
		registry:   NewPeerRegistryAdapter(reg),
		caps:       caps,
		scanner:    scanner,
		online:     true,
	}
	s.startOnlineMonitoring()
	return s, nil
}

func (s *Service) startOnlineMonitoring() {
	s.offlineCheckDone = make(chan struct{})
	s.onlineCheckTicker = time.NewTicker(5 * time.Second)
	go func() {
		defer s.onlineCheckTicker.Stop()
		for {
			select {
			case <-s.onlineCheckTicker.C:
				s.checkConnectivity()
			case <-s.offlineCheckDone:
				return
			}
		}
	}()
}

func (s *Service) checkConnectivity() {
	online := isOnline()
	s.mu.Lock()
	wasOnline := s.online
	s.online = online
	s.mu.Unlock()

	if wasOnline && !online {
		s.caps.Set(CapabilityInfo{Name: CapabilityNetworkDiscovery, Status: CapabilityOffline, Detail: "No network available"})
		s.Publish(Event{Type: EventTypeScanError, Error: fmt.Errorf("offline - no network available")})
	} else if !wasOnline && online {
		s.caps.Set(CapabilityInfo{Name: CapabilityNetworkDiscovery, Status: CapabilityAvailable, Detail: "Network available"})
	}
}

func isOnline() bool {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("HEAD", "https://www.google.com", nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (s *Service) IsOnline() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.online
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

func (s *Service) Stop() {
	s.mu.Lock()
	if s.onlineCheckTicker != nil {
		s.onlineCheckTicker.Stop()
	}
	if s.offlineCheckDone != nil {
		close(s.offlineCheckDone)
	}
	s.mu.Unlock()
	if s.scanner != nil {
		_ = s.scanner.Stop()
	}
}
