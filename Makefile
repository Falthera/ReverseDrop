.PHONY: all test vet build cli gui clean

all: test vet build

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build: cli gui

cli:
	go build -o reversedrop ./cmd/reversedrop

gui:
	go build -o reversedrop-gui ./gui

clean:
	del /Q reversedrop.exe reversedrop-gui.exe 2>nul || true
	del /Q reversedrop reversedrop-gui 2>nul || true
