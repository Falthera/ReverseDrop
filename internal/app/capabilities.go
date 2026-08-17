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

type Capability string

const (
	CapabilityBluetooth       Capability = "bluetooth"
	CapabilityNetworkDiscovery Capability = "network_discovery"
	CapabilityFileTransfer     Capability = "file_transfer"
	CapabilityGUI              Capability = "gui"
	CapabilityNotifications    Capability = "notifications"
)

type CapabilityStatus string

const (
	CapabilityAvailable      CapabilityStatus = "available"
	CapabilityUnavailable    CapabilityStatus = "unavailable"
	CapabilityUnsupported    CapabilityStatus = "unsupported"
	CapabilityPermissionDenied CapabilityStatus = "permission_denied"
	CapabilityExperimental   CapabilityStatus = "experimental"
	CapabilityOffline        CapabilityStatus = "offline"
)

type CapabilityInfo struct {
	Name     Capability
	Status   CapabilityStatus
	Detail   string
}

type CapabilitySet struct {
	items map[Capability]CapabilityInfo
}

func NewCapabilitySet() *CapabilitySet {
	return &CapabilitySet{items: make(map[Capability]CapabilityInfo)}
}

func (c *CapabilitySet) Set(info CapabilityInfo) {
	c.items[info.Name] = info
}

func (c *CapabilitySet) Get(name Capability) (CapabilityInfo, bool) {
	info, ok := c.items[name]
	return info, ok
}

func (c *CapabilitySet) List() []CapabilityInfo {
	out := make([]CapabilityInfo, 0, len(c.items))
	for _, info := range c.items {
		out = append(out, info)
	}
	return out
}
