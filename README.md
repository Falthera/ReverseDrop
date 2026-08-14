# ReverseDrop

Cross-platform peer-to-peer communication application inspired by AirDrop.

## What is ReverseDrop?

ReverseDrop is an open-source, cross-platform peer-to-peer communication application. It uses BLE (Bluetooth Low Energy) and local network discovery to find nearby devices and establish connections.

## Supported Platforms

| Platform | Build | GUI | BLE | Network Discovery | Status |
|---|---:|---:|---:|---:|---|
| Windows | ✓ | ✓ | ✓ | ✓ | Tier 1 |
| macOS | ✓ | ✓ | ✓ | ✓ | Tier 1 |
| Linux | ✓ | ✓ | ✓ | ✓ | Tier 1 |
| FreeBSD | ✓ | ✓ | ✓ | ✓ | Tier 2 |
| OpenBSD | ✓ | ✓ | ✓ | ✓ | Tier 2 |
| NetBSD | ✓ | ✓ | ✓ | ✓ | Tier 2 |

## Building

```bash
go build -o reversedrop ./cmd/reversedrop
```

## Running

```bash
./reversedrop scan
```

## GUI

```bash
go build -o reversedrop-gui ./gui
./reversedrop-gui
```

## Architecture

- `internal/protocol` - Peer model, state machine, registry
- `internal/discovery` - BLE and network discovery abstractions
- `internal/platform` - OS-specific capabilities
- `internal/app` - Application services and event model
- `gui` - Desktop GUI using Fyne
- `cmd/reversedrop` - CLI entry point

## License

GNU General Public License v3.0. See [LICENSE](LICENSE) for details.
