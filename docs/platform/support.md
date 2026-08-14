# Platform Support

## Tier 1 - Fully Tested

- Windows amd64/arm64
- macOS amd64/arm64
- Linux amd64/arm64

## Tier 2 - Builds and Core Features Work

- FreeBSD amd64/arm64
- OpenBSD amd64
- NetBSD amd64

## Platform Notes

### Windows
- Uses native BLE via WinRT
- GUI via Fyne
- High-DPI supported

### macOS
- Uses CoreBluetooth
- Privacy permissions required for BLE
- GUI via Fyne

### Linux
- Uses BlueZ via D-Bus
- GUI via Fyne
- Does not assume systemd

### BSD
- BLE support varies by system
- GUI via X11/Wayland
- May require additional packages

## Capability Reporting

The application detects and reports capabilities at runtime:
- Bluetooth available
- Network discovery available
- Notifications available
