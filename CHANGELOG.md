# Changelog

All notable changes to ReverseDrop will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Real Apple AirDrop BLE advertisement parser (company ID `0x004C`, sub-type `0x05`)
- Real AirDrop mDNS service discovery (`_airdrop._tcp.local.` on port 8770)
- AirDrop-compatible HTTP/1.1 API with `/Discover`, `/Ask`, and `/Upload` endpoints
- Binary Property List (bplist) encoding/decoding for all AirDrop messages
- CPIO newc archive creation and extraction
- DVZip (adaptive chunked compression) for file transfers
- gzip fallback for receivers that do not support DVZip
- Proper TLS 1.2/1.3 transport with AirDrop-style ALPN (`airdrop`)
- Windows MSI installer with proper WiX UI sequence
- macOS DMG and PKG installers
- Linux .deb and .rpm packages
- Comprehensive user documentation (INSTALL.md, USAGE.md, TROUBLESHOOTING.md)
- FAQ.md and KNOWN_LIMITATIONS.md
- Updated README.md to reflect genuine reverse-engineered AirDrop implementation
- Updated architecture docs with real protocol stack

### Changed
- Default transfer port changed from 9999 to 8770 (AirDrop standard)
- Transfer protocol completely rewritten from custom JSON to AirDrop HTTP/bplist/CPIO
- BLE parser updated from fake ServiceData records to real 18-byte AirDrop payload format
- mDNS service type changed from `_reversedrop._tcp.local.` to `_airdrop._tcp.local.`
- TLS ALPN changed from `reversedrop` to `airdrop`
- Certificate Common Name format changed to match AirDrop style (`com.apple.idms.appleid.prd.<UUID>`)

### Fixed
- WiX XML generation fixed to produce valid XML for candle/light compilation
- BLE parser now correctly extracts truncated SHA-256 hashes from real AirDrop advertisements

## [0.1.0] - 2025-01-15

### Added
- Initial release
- Basic BLE scanning with fake scanner
- Basic TLS file transfer (custom JSON protocol)
- CLI and GUI interfaces
- Cross-platform support (Windows, Mac, Linux)
