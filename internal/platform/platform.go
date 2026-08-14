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
package platform

import (
	"runtime"
)

type CapabilityReporter interface {
	BluetoothAvailable() (bool, string)
	NetworkDiscoveryAvailable() (bool, string)
	NotificationsAvailable() (bool, string)
}

type OSInfo struct {
	OS       string
	Arch     string
	Version  string
	Hostname string
}

func OS() OSInfo {
	return OSInfo{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: runtime.Version(),
	}
}
