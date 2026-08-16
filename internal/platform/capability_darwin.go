//go:build darwin

package platform

import (
	"os/exec"
	"runtime"
	"strings"
)

type defaultReporter struct{}

func NewDefaultReporter() CapabilityReporter {
	return &defaultReporter{}
}

func (d *defaultReporter) BluetoothAvailable() (bool, error) {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("sysctl", "hw.bluetooth").Output()
		if err != nil || !strings.Contains(string(out), "1") {
			return false, &BluetoothError{
				Category: BluetoothErrorUnavailable,
				Message:  "No Bluetooth adapter found. Check System Settings > Bluetooth to ensure Bluetooth is turned on, and System Settings > Privacy & Security > Bluetooth to grant permissions.",
			}
		}
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
