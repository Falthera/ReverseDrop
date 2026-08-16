# Roadmap

## Current Version: 1.0.0

### Completed
- [x] Genuine reverse-engineered Apple AirDrop protocol implementation
- [x] BLE discovery with Apple manufacturer ID 0x004C and sub-type 0x05
- [x] mDNS discovery with _airdrop._tcp.local. on port 8770
- [x] TLS 1.2/1.3 transport with self-signed certificates
- [x] HTTP API with /Discover, /Ask, and /Upload endpoints
- [x] Binary Property List (bplist) encoding/decoding
- [x] CPIO newc archive format implementation
- [x] DVZip adaptive compression format
- [x] Windows MSI installer with proper UI
- [x] macOS DMG and PKG installers
- [x] Linux .deb and .rpm packages
- [x] Cross-platform GUI using Fyne
- [x] CLI interface for advanced users
- [x] Comprehensive documentation
- [x] CI/CD pipeline with automated releases

### In Progress
- [ ] Improved contact matching with local address book integration
- [ ] Better error messages for Bluetooth permission issues
- [ ] Auto-accept for trusted contacts

### Planned for 1.1.0
- [ ] Android support (via Termux or native)
- [ ] iOS support (via Swift wrapper)
- [ ] Improved DVZip compression ratios
- [ ] Parallel file transfers
- [ ] Resumable transfers
- [ ] Transfer progress indicators in GUI
- [ ] Better Windows BLE reliability
- [ ] macOS AWDL support for better performance

### Planned for 1.2.0
- [ ] NameDrop / contact sharing support
- [ ] Handoff integration
- [ ] Custom device names in BLE advertisements
- [ ] Multiple file selection in GUI
- [ ] Dark mode support

### Ideas / Under Consideration
- [ ] Browser-based interface for easier use
- [ ] Network discovery without BLE (pure mDNS)
- [ ] Support for AirDrop "Everyone" mode
- [ ] QR code pairing for easier discovery
- [ ] Integration with system file managers
- [ ] Plugin system for custom handlers
