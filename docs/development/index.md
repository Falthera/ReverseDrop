# Development

## Prerequisites

- Go 1.22 or later
- Git

## Building

```bash
make cli
make gui
```

## Running Tests

```bash
make test
make test-race
make vet
```

## Protocol Implementation

ReverseDrop is a genuine reverse-engineering of Apple's AirDrop protocol. Key protocol details:

- **BLE**: Apple company ID `0x004C`, sub-type `0x05`, 18-byte payload with truncated SHA-256 hashes
- **mDNS**: Service type `_airdrop._tcp.local.`, port 8770
- **TLS**: Self-signed certificates, TLS 1.2/1.3
- **HTTP**: `/Discover`, `/Ask`, `/Upload` with bplist bodies
- **Archive**: CPIO newc (`070701`) + DVZip/gzip

## CLI Usage

```bash
./reversedrop scan
./reversedrop scan --timeout 10s
./reversedrop scan --target AA:BB:CC:DD:EE:FF
./reversedrop scan --verbose
```

## GUI Usage

```bash
./reversedrop-gui
```

## Project Structure

- `cmd/reversedrop` - CLI entry point
- `gui` - Fyne-based desktop GUI
- `internal/app` - Application services, events, capabilities
- `internal/discovery` - BLE and mDNS/AWDL discovery
- `internal/discovery/parser` - Apple AirDrop BLE advertisement parser
- `internal/discovery/network` - mDNS `_airdrop._tcp.local.` discovery
- `internal/protocol/peer` - Peer model, state machine, registry
- `internal/transfer` - AirDrop-compatible TLS + HTTP transfer with bplist and CPIO/DVZip
- `internal/platform` - OS-specific capabilities
- `internal/notification` - Desktop notifications
- `internal/config` - JSON configuration
- `internal/trust` - Trust store
- `docs` - Documentation

## Contributing

See CONTRIBUTING.md
