.PHONY: all test vet lint fmt build cli gui clean docker

all: test vet lint fmt build

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run --timeout=5m

fmt:
	gofmt -w .
	@echo "Code formatted"

build: cli gui

cli:
	go build -o reversedrop ./cmd/reversedrop

gui:
	go build -o reversedrop-gui ./gui

clean:
	rm -f reversedrop reversedrop-gui reversedrop.exe reversedrop-gui.exe

docker:
	docker build -t reversedrop:latest .
