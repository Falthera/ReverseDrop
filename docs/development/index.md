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
- `internal/app` - Application services, events, capabilities
- `internal/discovery` - BLE and network discovery
- `internal/protocol` - Peer model, state machine, registry
- `internal/platform` - OS-specific capabilities
- `gui` - Desktop GUI using Fyne
- `tests` - Integration tests
- `docs` - Documentation

## Contributing

See CONTRIBUTING.md
