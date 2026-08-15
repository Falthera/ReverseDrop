# Building

## Prerequisites

- Go 1.22 or later

## CLI Build

```bash
go build -o reversedrop ./cmd/reversedrop
```

## GUI Build

```bash
go build -o reversedrop-gui ./gui
```

## Running Tests

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Fyne Requirements

For BSD systems, you need X11 and Wayland libraries installed. See the Fyne documentation for platform-specific setup instructions.

## Protocol Implementation Notes

ReverseDrop implements the real Apple AirDrop protocol stack:

- **BLE**: Apple manufacturer ID `0x004C`, AirDrop sub-type `0x05`
- **mDNS**: Service type `_airdrop._tcp.local.` on port 8770
- **TLS**: Self-signed certificates, TLS 1.2/1.3
- **HTTP**: `/Discover`, `/Ask`, `/Upload` endpoints with bplist bodies
- **Archive**: CPIO newc + DVZip/gzip compression

The `internal/transfer` package uses `howett.net/plist` for bplist encoding/decoding.
The `internal/discovery/parser` package implements the real AirDrop BLE advertisement parser.
