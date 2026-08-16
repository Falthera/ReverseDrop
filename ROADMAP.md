# Roadmap

## Current Version: 1.0.0

### Completed
- [x] Genuine reverse-engineered Apple AirDrop protocol implementation
- [x] BLE discovery with Apple manufacturer ID `0x004C` and sub-type `0x05`
- [x] mDNS discovery with `_airdrop._tcp.local.` on port 8770
- [x] TLS 1.2/1.3 transport with self-signed certificates
- [x] HTTP API with `/Discover`, `/Ask`, and `/Upload` endpoints
- [x] Binary Property List (bplist) encoding/decoding
- [x] CPIO newc archive format implementation
- [x] DVZip adaptive compression format
- [x] Windows MSI installer with proper UI
- [x] macOS DMG and PKG installers
- [x] Linux `.deb` and `.rpm` packages
- [x] Cross-platform GUI using Fyne
- [x] CLI interface for advanced users
- [x] Comprehensive documentation
- [x] CI/CD pipeline with automated releases
- [x] GNU GPL v3 license inclusion in installers
- [x] Security policy and vulnerability reporting process
- [x] Code of Conduct
- [x] Issue templates and PR template
- [x] golangci-lint configuration
- [x] Dockerfile for reproducible builds

---

## In Progress / Next Release: 1.1.0

### Protocol Improvements
- [x] Improved contact matching with local address book integration
- [x] Apple ID Validation Record (VR) certificate chain parsing
- [x] Better error messages for Bluetooth permission issues
- [x] Auto-accept for trusted contacts
- [x] Support for AirDrop "Everyone" mode vs "Contacts Only" mode

### Platform Improvements
- [ ] Better Windows BLE reliability and driver compatibility
- [ ] macOS AWDL support for better performance
- [ ] Linux BlueZ D-Bus integration improvements
- [ ] FreeBSD/OpenBSD BLE stack improvements
- [ ] iOS support (via Swift wrapper or native implementation)
- [ ] Android support (via Termux or native)

### User Experience
- [x] Transfer progress indicators in GUI
- [ ] Parallel file transfers
- [ ] Resumable transfers
- [ ] Multiple file selection in GUI
- [ ] Dark mode support
- [ ] System tray integration
- [ ] Better notification system
- [ ] Offline mode indicator

### Performance
- [x] Improved DVZip compression ratios
- [ ] Better memory usage for large files
- [x] Optimized BLE scanning to reduce battery usage
- [ ] Faster mDNS discovery

---

## Planned: 1.2.0

### New Features
- [ ] NameDrop / contact sharing support
- [ ] Handoff integration
- [ ] Custom device names in BLE advertisements
- [ ] Multiple file selection in GUI
- [ ] Dark mode support
- [ ] Transfer history
- [ ] Favorite/trusted devices
- [ ] QR code pairing for easier discovery

### Protocol Enhancements
- [ ] Full Apple PKI certificate validation support
- [ ] Support for additional AirDrop record types
- [ ] Enhanced contact hash matching
- [ ] Support for AirDrop "Nearby" mode

---

## Ideas / Under Consideration

### Platform Expansion
- [ ] Browser-based interface for easier use
- [ ] WebRTC-based fallback when BLE is unavailable
- [ ] Mobile apps (iOS/Android) as standalone projects
- [ ] Raspberry Pi / embedded device support

### Protocol & Compatibility
- [ ] Network discovery without BLE (pure mDNS)
- [ ] Support for Samsung Quick Share protocol
- [ ] Support for Google Nearby Share protocol
- [ ] Universal file sharing protocol abstraction layer

### User Experience
- [ ] Integration with system file managers
- [ ] Context menu integration ("Send with ReverseDrop")
- [ ] Plugin system for custom handlers
- [ ] Theme system for GUI
- [ ] Accessibility improvements

### Developer Experience
- [ ] Public Go library for AirDrop protocol
- [ ] Python bindings for protocol implementation
- [ ] Protocol fuzzing and test suite
- [ ] Interoperability test suite with real AirDrop devices

### Infrastructure
- [ ] Automated GPG signing of releases
- [ ] Multi-platform release automation
- [ ] Package distribution via Homebrew, Chocolatey, Scoop
- [ ] Flatpak and Snap packages for Linux
- [ ] Automated security scanning in CI

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get involved. We welcome contributions of all kinds, from documentation improvements to protocol implementations.
