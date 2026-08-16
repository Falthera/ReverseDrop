# ReverseDrop

![License: GPL-3.0](https://img.shields.io/badge/License-GPLv3-blue.svg)
![Go 1.22](https://img.shields.io/badge/Go-1.22-blue)
![Windows](https://img.shields.io/badge/Windows-10%2F11-success)
![macOS](https://img.shields.io/badge/macOS-12%2B-success)
![Linux](https://img.shields.io/badge/Linux-Ubuntu%2FFedora%2FDebian-success)
![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)

**A reverse-engineered, open-source implementation of Apple's AirDrop protocol.**

ReverseDrop replicates the real Apple AirDrop peer-to-peer file sharing protocol so that non-Apple devices can participate in the same AirDrop ecosystem. It is not merely "inspired by" AirDrop. It is AirDrop, reimplemented from public reverse-engineering research.

> **Note:** ReverseDrop is a genuine reverse-engineering effort based on published academic research. It is not affiliated with or endorsed by Apple Inc.

---

## What is ReverseDrop?

ReverseDrop is a genuine reverse-engineering of Apple's AirDrop. It implements the same protocols that real AirDrop uses:

- **BLE discovery** with Apple manufacturer ID `0x004C` and AirDrop sub-type `0x05`
- **mDNS/DNS-SD** service discovery for `_airdrop._tcp.local.`
- **TLS 1.2/1.3** transport with self-signed certificates
- **HTTP/1.1** API with `/Discover`, `/Ask`, and `/Upload` endpoints
- **Binary Property Lists** (bplist) for all message bodies
- **CPIO newc` archives compressed with **DVZip** (Apple's adaptive chunked format) or gzip fallback

This means ReverseDrop can discover and communicate with real Apple AirDrop devices.

---

## Features

- **Interoperable with Apple AirDrop**: uses the same BLE advertisements, mDNS records, and wire protocol
- **Cross-platform**: Windows, Mac, and Linux
- **No account needed**: no iCloud, no Apple ID required
- **Private by design**: files go directly between devices
- **Works offline**: no internet or cloud required

---

## How It Works

ReverseDrop implements the same two-phase discovery that Apple AirDrop uses:

1. **BLE Wake-Up**: Broadcasts truncated SHA-256 contact hashes in Apple BLE advertisements (company ID `0x004C`, sub-type `0x05`). Nearby AirDrop receivers activate their AWDL/Wi-Fi Direct interface when they see these advertisements.
2. **mDNS Discovery**: Uses DNS-SD (`_airdrop._tcp.local.`) over the local network to discover receivers and exchange capabilities.
3. **TLS Handshake**: Establishes a TLS 1.2/1.3 connection (self-signed, no hostname verification) to port 8770.
4. **HTTP API**: Exchanges binary property list messages:
   - `POST /Discover`: sender identity and capabilities
   - `POST /Ask`: file metadata, triggers receiver consent UI
   - `POST /Upload`: chunked CPIO+archive transfer
5. **Archive Delivery**: Files are packed into a CPIO newc archive (`070701` magic) compressed with DVZip or gzip, matching the exact format AirDrop uses.

---

## Supported Platforms

| Platform | Status | Notes |
|----------|--------|-------|
| Windows 10/11 | Fully supported | Uses WinRT BLE; MSI installer available |
| macOS 12+ | Fully supported | Uses CoreBluetooth; dmg/pkg installers available |
| Linux (Ubuntu, Fedora, etc.) | Fully supported | Uses BlueZ; .deb and .rpm packages available |
| FreeBSD / OpenBSD / NetBSD | Community tested | May require extra steps |

---

## Installation

See [INSTALL.md](INSTALL.md) for detailed instructions.

**Quick links:**
- [Windows Installer (MSI)](https://github.com/Falthera/ReverseDrop/releases)
- [macOS Installer (DMG / PKG)](https://github.com/Falthera/ReverseDrop/releases)
- [Linux Packages (.deb / .rpm)](https://github.com/Falthera/ReverseDrop/releases)

---

## Usage

See [USAGE.md](USAGE.md) for a complete guide.

**Quick start:**
- **Windows**: Launch "ReverseDrop" from the Start Menu
- **Mac**: Open "ReverseDrop" from Applications or Launchpad
- **Linux**: Run `reversedrop-gui` from your applications menu

The app starts scanning for nearby AirDrop devices automatically. When you see a device, click it, select a file, and confirm.

---

## Protocol Details

For a deep dive into the reverse-engineered protocol, see [docs/protocol/research.md](docs/protocol/research.md).

### Key Protocol Constants

| Item | Value |
|------|-------|
| BLE Apple Company ID | `0x004C` |
| BLE AirDrop Sub-Type | `0x05` |
| BLE PDU Type | `ADV_NONCONN_IND` |
| Hash truncation | First **16 bits** of SHA-256 |
| mDNS Service Type | `_airdrop._tcp.local.` |
| TCP Port | **8770** |
| TLS Version | TLS 1.2 / 1.3 |
| HTTP Endpoints | `/Discover`, `/Ask`, `/Upload`, `/Error` |
| Upload Format | DVZip (`application/x-dvzip`) or gzip + CPIO newc |
| Archive Magic | `070701` (CPIO newc) |
| Archive Terminator | `TRAILER!!!` |

---

## Privacy & Security

See [SECURITY.md](SECURITY.md) for full details.

> **Warning**: ReverseDrop does not implement Apple's full PKI certificate validation. Self-signed certificates are accepted without verification. Only use ReverseDrop on trusted networks.

- Files are transferred directly between devices. They never pass through a server.
- No accounts, no tracking, no telemetry.
- Open source: you can inspect the code yourself.
- TLS encrypts all transfers. No peer authentication beyond the application layer.

---

## Known Limitations

See [KNOWN_LIMITATIONS.md](KNOWN_LIMITATIONS.md) for current constraints.

> **Deprecation Notice**: Some features may change as the reverse-engineering research progresses. Check the changelog for breaking changes.

- Large file transfers may be slow on some systems.
- Some Linux distributions require additional Bluetooth packages.
- Contact hashing is simplified. No real Apple ID validation.
- DVZip adaptive compression is partially implemented.

---

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for solutions to common problems.

---

## Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Contributors

<a href="https://contrib.rocks/preview/raw?repo=Falthera/ReverseDrop">
  <img src="https://contrib.rocks/preview/raw?repo=Falthera/ReverseDrop" alt="Contributors" />
</a>

---

## License

ReverseDrop is released under the [GNU General Public License v3.0](LICENSE).

---

## Acknowledgments

Built by reverse-engineering Apple's AirDrop protocol based on public academic research:

- Stute et al., *"A Billion Open Interfaces for Eve and Mallory"*, USENIX Security '19
- Heinrich et al., *"Discontinued Privacy: Personal Data Leaks in Apple BLE Continuity Protocols"*, PoPETs 2020
- Ebrahim & Tippenhauer, *"Protocol Prying: AirDrop & Quick Share"*, arXiv:2606.26967, 2026

Inspired by [OpenDrop](https://github.com/seemoo-lab/opendrop) and [OWL](https://github.com/seemoo-lab/owl).
