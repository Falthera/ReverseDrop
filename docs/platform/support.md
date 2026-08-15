# Platform Support

## Tier 1 - Fully Tested

- Windows amd64/arm64
- macOS amd64/arm64
- Linux amd64/arm64

## Tier 2 - Builds and Core Features Work

- FreeBSD amd64/arm64
- OpenBSD amd64
- NetBSD amd64

## AirDrop Protocol Compatibility

ReverseDrop implements the real Apple AirDrop protocol. It can discover and communicate with:
- macOS devices running AirDrop
- iOS devices running AirDrop
- Other ReverseDrop instances

## Platform Notes

### Windows
- Uses WinRT BLE for discovery
- GUI via Fyne
- High-DPI supported

### macOS
- Uses CoreBluetooth for BLE discovery
- Privacy permissions required for BLE
- GUI via Fyne
- Can communicate with native AirDrop

### Linux
- Uses BlueZ via D-Bus for BLE discovery
- GUI via Fyne
- Does not assume systemd
- Can communicate with native AirDrop

### BSD
- BLE support varies by system
- GUI via X11/Wayland
- May require additional packages
- AirDrop compatibility depends on BLE stack

## Capability Reporting

The application detects and reports capabilities at runtime:
- Bluetooth available
- Network discovery available
- Notifications available
