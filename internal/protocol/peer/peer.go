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
	"time"
)

const (
	StateUnknown      PeerState = "unknown"
	StateDiscovered   PeerState = "discovered"
	StateConnecting   PeerState = "connecting"
	StateHandshaking  PeerState = "handshaking"
	StateTransferring PeerState = "transferring"
	StateFailed       PeerState = "failed"
	StateDisconnected PeerState = "disconnected"
)

type PeerState string

type Peer struct {
	Address         string
	RSSI            int
	DeviceName      string
	RecordIDs       []string
	LastSeen        time.Time
	State           PeerState
	Capabilities    []string
	Platform        string
	ConnectionAddrs []string
	TrustState      string
}

type PeerEvent struct {
	Peer   Peer
	Reason string
}

type RegistryEventType string

const (
	EventPeerUpserted RegistryEventType = "upserted"
	EventPeerRemoved  RegistryEventType = "removed"
	EventPeerUpdated  RegistryEventType = "updated"
)

type RegistryEvent struct {
	Type      RegistryEventType
	Peer      Peer
	Previous  *Peer
	Timestamp time.Time
}

type PeerRegistry interface {
	Upsert(p Peer)
	Get(address string) (Peer, bool)
	List() []Peer
	Remove(address string)
	Update(address string, apply func(*Peer))
	Transition(address string, newState PeerState) bool
	Events() <-chan RegistryEvent
	Close()
}
