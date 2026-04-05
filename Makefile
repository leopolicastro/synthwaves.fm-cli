BINARY_NAME=synthwaves
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/leo/synthwaves-cli/cmd.Version=$(VERSION)"

.PHONY: build install clean run tui

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

clean:
	rm -f $(BINARY_NAME)
	go clean

run:
	go run $(LDFLAGS) .

tui:
	go run $(LDFLAGS) . tui

lint:
	go vet ./...

test:
	go test ./...
