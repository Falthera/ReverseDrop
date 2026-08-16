//go:build windows

package platform

import (
	"runtime"

	"golang.org/x/sys/windows"
)

var (
	winrtDLL           = windows.NewLazyDLL("api-ms-win-core-winrt-l1-1-0.dll")
	winrtRoInitialize   = winrtDLL.NewProc("RoInitialize")
	winrtRoUninitialize = winrtDLL.NewProc("RoUninitialize")
)

type defaultReporter struct{}

func NewDefaultReporter() CapabilityReporter {
	return &defaultReporter{}
}

func (d *defaultReporter) BluetoothAvailable() (bool, error) {
	if runtime.GOOS == "windows" {
		ret, _, _ := winrtRoInitialize.Call(uintptr(0))
		if ret != 0 {
			winrtRoUninitialize.Call()
			return false, &BluetoothError{
				Category: BluetoothErrorPermissionDenied,
				Message:  "Bluetooth adapter or WinRT permissions issue. Ensure you have a Bluetooth adapter with updated drivers and WinRT permissions enabled.",
			}
		}
		winrtRoUninitialize.Call()
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
