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
