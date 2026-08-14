# Architecture Overview

## Core Principles

- Separate product layers: Platform -> Hardware/OS APIs -> Discovery -> Peer management -> Protocol -> Transfer -> Application services -> GUI/CLI
- GUI and CLI share the same core services
- Platform-specific functionality is isolated
- Unknown reverse-engineered behavior is clearly marked

## Key Packages

- `internal/protocol` - Peer model, state machine, peer registry
- `internal/discovery` - BLE and network discovery abstractions
- `internal/platform` - OS-specific capabilities
- `internal/app` - Application services, events, capabilities
- `internal/security` - Future cryptographic boundaries
- `gui` - Desktop GUI using Fyne
- `cmd/reversedrop` - CLI entry point

## State Machine

```
Unknown -> Discovered -> Connecting -> Handshaking -> Transferring
```

Invalid transitions are rejected.

## Capability System

Each capability reports:
- available
- unavailable
- unsupported
- permission-denied
- experimental

## Testing

- Unit tests for domain logic
- Fake BLE scanner for integration tests
- No physical hardware required for tests
