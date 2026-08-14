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
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

type EventType string

const (
	EventTypePeerDiscovered EventType = "peer.discovered"
	EventTypePeerLost       EventType = "peer.lost"
	EventTypePeerUpdated    EventType = "peer.updated"
	EventTypeScanStarted    EventType = "scan.started"
	EventTypeScanStopped    EventType = "scan.stopped"
	EventTypeScanError      EventType = "scan.error"
	EventTypeCapabilityChange EventType = "capability.change"
)

type Event struct {
	Type      EventType
	Timestamp int64
	Peer      *peer.Peer
	Error     error
	Metadata  map[string]string
}

type Subscriber interface {
	OnEvent(Event)
}

type Publisher interface {
	Subscribe(Subscriber)
	Unsubscribe(Subscriber)
	Publish(Event)
}
