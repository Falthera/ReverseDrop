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
	"testing"
)

type fakeReporter struct {
	available bool
	detail    string
}

func (f *fakeReporter) BluetoothAvailable() (bool, string)          { return f.available, f.detail }
func (f *fakeReporter) NetworkDiscoveryAvailable() (bool, string)   { return f.available, f.detail }
func (f *fakeReporter) NotificationsAvailable() (bool, string)      { return f.available, f.detail }

func TestDefaultReporterBluetooth(t *testing.T) {
	r := NewDefaultReporter()
	avail, _ := r.BluetoothAvailable()
	if !avail && runtime.GOOS != "js" {
		t.Log("Bluetooth reported unavailable on", runtime.GOOS)
	}
}

func TestDefaultReporterNetwork(t *testing.T) {
	r := NewDefaultReporter()
	avail, _ := r.NetworkDiscoveryAvailable()
	if !avail && runtime.GOOS != "js" {
		t.Log("Network discovery reported unavailable on", runtime.GOOS)
	}
}

func TestDefaultReporterNotifications(t *testing.T) {
	r := NewDefaultReporter()
	_, _ = r.NotificationsAvailable()
}

func TestOS(t *testing.T) {
	info := OS()
	if info.OS != runtime.GOOS {
		t.Errorf("OS() OS = %s, want %s", info.OS, runtime.GOOS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("OS() Arch = %s, want %s", info.Arch, runtime.GOARCH)
	}
	if info.Version != runtime.Version() {
		t.Errorf("OS() Version = %s, want %s", info.Version, runtime.Version())
	}
}
