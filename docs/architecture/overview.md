# Architecture Overview

## Core Principles

- **Genuine AirDrop protocol**: not inspired by, but a direct reverse-engineering of Apple's AirDrop
- GUI and CLI share the same core services
- Platform-specific functionality is isolated
- All protocol details match the published AirDrop specifications

## Key Packages

- `internal/protocol` - Peer model, state machine, peer registry
- `internal/discovery` - BLE and mDNS/AWDL discovery using real AirDrop formats
- `internal/platform` - OS-specific capabilities
- `internal/app` - Application services, events, capabilities
- `internal/transfer` - AirDrop-compatible TLS + HTTP transfer with bplist and CPIO/DVZip
- `internal/discovery/parser` - Apple BLE advertisement parser (company ID 0x004C, sub-type 0x05)
- `gui` - Desktop GUI using Fyne

## AirDrop Protocol Stack

```
Layer 7: Application: HTTP/1.1 over TLS (/Discover, /Ask, /Upload)
Layer 6: Encoding: Apple Binary Property List (bplist00)
Layer 5: Compression: DVZip (adaptive) / gzip fallback
Layer 4: Archive: CPIO newc format (070701 magic, TRAILER!!!)
Layer 3: Security: TLS 1.2/1.3 (self-signed, no client auth)
Layer 2: Network: IPv6 link-local on AWDL / Wi-Fi Direct
Layer 1: Discovery: BLE (Apple 0x004C / 0x05) + mDNS (_airdrop._tcp.local.)
```

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

## AirDrop BLE Advertisement Format

```
Apple Company ID: 0x004C (little-endian: 4C 00)
Apple Sub-Type: 0x05

Payload (18 bytes):
  Offset 0-7:   8 zero bytes (padding)
  Offset 8:     1 byte version (0x01)
  Offset 9-10:  2 bytes SHA-256[0:2] of AppleID email
  Offset 11-12: 2 bytes SHA-256[0:2] of phone number
  Offset 13-14: 2 bytes SHA-256[0:2] of primary email
  Offset 15-16: 2 bytes SHA-256[0:2] of secondary email
  Offset 17:    1 byte zero (suffix)
```

## mDNS Discovery

- Service type: `_airdrop._tcp.local.`
- Port: 8770
- TXT records: `flags=3FB`, `cname`, `ehash`, `phash`

## Transfer Protocol

1. TCP connection to port 8770
2. TLS 1.2/1.3 handshake (self-signed certs, no hostname verification)
3. HTTP POST /Discover: bplist request/response
4. HTTP POST /Ask: bplist request/response (triggers UI consent)
5. HTTP POST /Upload: chunked CPIO+archive transfer

## Testing

- Unit tests for domain logic
- Fake BLE scanner for integration tests
- No physical hardware required for tests
