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
	err       error
	detail    string
}

func (f *fakeReporter) BluetoothAvailable() (bool, error)          { return f.available, f.err }
func (f *fakeReporter) NetworkDiscoveryAvailable() (bool, string)   { return f.available, f.detail }
func (f *fakeReporter) NotificationsAvailable() (bool, string)      { return f.available, f.detail }

func TestDefaultReporterBluetooth(t *testing.T) {
	r := NewDefaultReporter()
	avail, err := r.BluetoothAvailable()
	if err != nil {
		if bluetoothErr, ok := err.(*BluetoothError); ok {
			t.Logf("Bluetooth reported error on %s: category=%d, message=%s", runtime.GOOS, bluetoothErr.Category, bluetoothErr.Message)
		} else {
			t.Logf("Bluetooth reported error on %s: %v", runtime.GOOS, err)
		}
	}
	if !avail && runtime.GOOS != "js" {
		t.Log("Bluetooth reported unavailable on", runtime.GOOS)
	}
}

func TestBluetoothErrorCategories(t *testing.T) {
	tests := []struct {
		category BluetoothErrorCategory
		message  string
	}{
		{BluetoothErrorUnavailable, "no adapter"},
		{BluetoothErrorPermissionDenied, "permission denied"},
		{BluetoothErrorDisabled, "disabled"},
		{BluetoothErrorUnknown, "unknown"},
	}
	for _, tt := range tests {
		err := &BluetoothError{Category: tt.category, Message: tt.message}
		if err.Error() != tt.message {
			t.Errorf("BluetoothError.Error() = %q, want %q", err.Error(), tt.message)
		}
		if err.Category != tt.category {
			t.Errorf("BluetoothError.Category = %d, want %d", err.Category, tt.category)
		}
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
