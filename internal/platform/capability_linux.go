//go:build linux

package platform

import (
	"runtime"

	"github.com/godbus/dbus/v5"
)

type defaultReporter struct{}

func NewDefaultReporter() CapabilityReporter {
	return &defaultReporter{}
}

func (d *defaultReporter) BluetoothAvailable() (bool, error) {
	if runtime.GOOS == "linux" {
		conn, err := dbus.ConnectSystemBus()
		if err != nil {
			return false, &BluetoothError{
				Category: BluetoothErrorUnavailable,
				Message:  "No Bluetooth adapter found or bluetoothd service is not running. Ensure bluetoothd is active, install libbluetooth-dev, and add your user to the bluetooth group: sudo usermod -aG bluetooth $USER",
			}
		}
		conn.Close()
	}
	return true, nil
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
