# Known Limitations

ReverseDrop implements the Apple AirDrop protocol as reverse-engineered from public research. Some limitations exist compared to the official Apple AirDrop implementation.

## Protocol Limitations

- **No Apple ID Validation**: ReverseDrop does not implement Apple's PKCS#7 signed Validation Record (VR) chain. Contact matching uses truncated SHA-256 hashes only.
- **Simplified Contact Matching**: Real AirDrop compares hashes against the user's address book. ReverseDrop uses the hashes from its own configuration.
- **No Auto-Accept**: Real AirDrop can auto-accept files from trusted contacts. ReverseDrop always requires manual confirmation.
- **DVZip Partially Implemented**: The DVZip adaptive compression format is implemented, but may not handle all edge cases that Apple's `sharingd` handles.
- **No AWDL**: On macOS/iOS, AirDrop uses AWDL (Apple Wireless Direct Link). ReverseDrop uses standard Wi-Fi Direct or mDNS over the local network.
- **No Handoff/NameDrop**: Only basic file transfer is implemented. NameDrop (contact sharing) and other AirDrop features are not supported.

## Platform Limitations

- **Windows BLE**: WinRT BLE scanner is a stub in some cases. Real BLE scanning may require additional drivers.
- **Linux BLE**: BlueZ D-Bus interface is used. Some distributions may require `libbluetooth-dev` and `bluetoothd` to be running.
- **macOS Permissions**: First-run Bluetooth permission grant is required. The app may need to be added to System Settings > Privacy & Security > Bluetooth.

## Security Limitations

- **No Certificate Pinning**: Self-signed certificates are accepted without verification. Any device can impersonate another on the same network.
- **No Mutual Authentication**: TLS provides encryption only, not peer identity verification.
- **Truncated Hashes**: Using only the first 16 bits of SHA-256 means there is a higher chance of hash collisions compared to real AirDrop.

## Performance Limitations

- **Large Files**: Transfers of files larger than 100MB may be slow, especially on Windows where BLE scanning overhead is higher.
- **No Parallel Transfers**: Only one file transfer at a time is supported.
- **No Resumable Transfers**: If a transfer is interrupted, it must be restarted from the beginning.

## Compatibility Limitations

- **Apple Device Discovery**: ReverseDrop can discover Apple AirDrop devices via mDNS, but Apple devices may not discover ReverseDrop instances because they do not broadcast the same BLE advertisements (no Apple ID).
- **Firewall**: Some firewalls may block port 8770 or mDNS traffic.
- **Network Isolation**: mDNS may not work across VLANs or network segments without multicast routing.
