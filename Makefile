BINARY     := bedrock-cli
MODULE     := github.com/kernelpanic09/bedrock-cli
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-s -w -X github.com/kernelpanic09/bedrock-cli/internal/cmd.Version=$(VERSION)"
BUILD_DIR  := dist

.PHONY: build test lint fmt install clean release help

## build: compile the binary for the current platform
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/$(BINARY)

## install: install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/$(BINARY)

## test: run all unit tests
test:
	go test -race -count=1 ./...

## test-verbose: run tests with verbose output
test-verbose:
	go test -v -race -count=1 ./...

## test-cover: run tests and show coverage
test-cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## fmt: run gofmt across the codebase
fmt:
	gofmt -w -s .

## lint: run golangci-lint (install: brew install golangci-lint)
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	go vet ./...

## tidy: update go.sum and remove unused dependencies
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

## release: build for all platforms via goreleaser (dry run)
release-dry:
	goreleaser release --snapshot --clean

## release: publish a real release (needs GITHUB_TOKEN)
release:
	goreleaser release --clean

## help: print this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

build: $(BUILD_DIR)
