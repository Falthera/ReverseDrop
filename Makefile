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
	rm -f reversedrop reversedrop-gui reversedrop.exe reversedrop-gui.exe
