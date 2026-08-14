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
	"fmt"
	"sync"
	"time"
)

var validTransitions = map[PeerState]map[PeerState]struct{}{
	StateUnknown: {
		StateDiscovered: {},
		StateFailed:     {},
	},
	StateDiscovered: {
		StateConnecting: {},
		StateFailed:     {},
		StateDisconnected: {},
	},
	StateConnecting: {
		StateHandshaking: {},
		StateFailed:      {},
		StateDisconnected: {},
	},
	StateHandshaking: {
		StateTransferring: {},
		StateFailed:       {},
		StateDisconnected: {},
	},
	StateTransferring: {
		StateDisconnected: {},
		StateFailed:       {},
	},
	StateFailed: {
		StateDiscovered: {},
		StateConnecting: {},
	},
	StateDisconnected: {
		StateDiscovered: {},
		StateConnecting: {},
		StateFailed:     {},
	},
}

func IsValidTransition(from, to PeerState) bool {
	transitions, ok := validTransitions[from]
	if !ok {
		return false
	}
	_, ok = transitions[to]
	return ok
}

func Transition(p *Peer, newState PeerState) error {
	if p == nil {
		return fmt.Errorf("peer is nil")
	}
	if !IsValidTransition(p.State, newState) {
		return fmt.Errorf("invalid state transition: %s -> %s", p.State, newState)
	}
	p.State = newState
	return nil
}

type registry struct {
	peers map[string]Peer
	mu    sync.RWMutex
	ch    chan RegistryEvent
}

func NewRegistry() PeerRegistry {
	return &registry{
		peers: make(map[string]Peer),
		ch:    make(chan RegistryEvent, 256),
	}
}

func (r *registry) Upsert(p Peer) {
	if p.Address == "" {
		return
	}
	r.mu.Lock()
	prev, existed := r.peers[p.Address]
	r.peers[p.Address] = p
	r.mu.Unlock()

	evt := RegistryEvent{
		Type:      EventPeerUpserted,
		Peer:      p,
		Timestamp: time.Now(),
	}
	if existed {
		evt.Type = EventPeerUpdated
		evt.Previous = &prev
	}
	select {
	case r.ch <- evt:
	default:
	}
}

func (r *registry) Get(address string) (Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.peers[address]
	return p, ok
}

func (r *registry) List() []Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Peer, 0, len(r.peers))
	for _, p := range r.peers {
		out = append(out, p)
	}
	return out
}

func (r *registry) Remove(address string) {
	r.mu.Lock()
	prev, ok := r.peers[address]
	if ok {
		delete(r.peers, address)
	}
	r.mu.Unlock()
	if ok {
		select {
		case r.ch <- RegistryEvent{
			Type:      EventPeerRemoved,
			Peer:      prev,
			Timestamp: time.Now(),
		}:
		default:
		}
	}
}

func (r *registry) Update(address string, apply func(*Peer)) {
	r.mu.Lock()
	p, ok := r.peers[address]
	if !ok {
		r.mu.Unlock()
		return
	}
	prev := p
	apply(&p)
	r.peers[address] = p
	r.mu.Unlock()

	select {
	case r.ch <- RegistryEvent{
		Type:      EventPeerUpdated,
		Peer:      p,
		Previous:  &prev,
		Timestamp: time.Now(),
	}:
	default:
	}
}

func (r *registry) Transition(address string, newState PeerState) bool {
	r.mu.Lock()
	p, ok := r.peers[address]
	if !ok {
		r.mu.Unlock()
		return false
	}
	prev := p
	if err := Transition(&p, newState); err != nil {
		r.mu.Unlock()
		return false
	}
	r.peers[address] = p
	r.mu.Unlock()

	select {
	case r.ch <- RegistryEvent{
		Type:      EventPeerUpdated,
		Peer:      p,
		Previous:  &prev,
		Timestamp: time.Now(),
	}:
	default:
	}
	return true
}

func (r *registry) Events() <-chan RegistryEvent {
	return r.ch
}

func (r *registry) Close() {
	close(r.ch)
}
