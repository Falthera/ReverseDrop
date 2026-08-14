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

import "runtime"

type defaultReporter struct{}

func NewDefaultReporter() CapabilityReporter {
	return &defaultReporter{}
}

func (d *defaultReporter) BluetoothAvailable() (bool, string) {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		return true, "platform supports BLE"
	case "freebsd", "openbsd", "netbsd":
		return true, "platform may support BLE via HCI/BlueZ"
	default:
		return false, "unknown platform"
	}
}

func (d *defaultReporter) NetworkDiscoveryAvailable() (bool, string) {
	switch runtime.GOOS {
	case "darwin", "linux", "windows", "freebsd", "openbsd", "netbsd":
		return true, "platform supports local network"
	default:
		return false, "unknown platform"
	}
}

func (d *defaultReporter) NotificationsAvailable() (bool, string) {
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		return true, "platform supports notifications"
	default:
		return false, "platform notifications not implemented"
	}
}
