GO := /opt/homebrew/bin/go
BINARY := gitto
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build run test lint clean install

## Build the binary
build:
	$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

## Run gitto in the current directory
run: build
	./$(BINARY)

## Run all tests
test:
	$(GO) test ./... -v

## Run tests with coverage
test-cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

## Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

## Clean build artifacts
clean:
	rm -f $(BINARY) coverage.out coverage.html

## Install to GOPATH/bin
install:
	$(GO) install -ldflags "-X main.version=$(VERSION)" .
