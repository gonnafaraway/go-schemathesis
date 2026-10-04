PROJECT_NAME := $(shell go list -m)
BINARY := bin/schemathesis

# Version reported by `schemathesis --version`. Derived from git so a build from
# a tag reports the tag, and a build from a dirty tree says so.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# -s -w strips the symbol table and DWARF data; -X stamps the version.
LDFLAGS := -s -w -X $(PROJECT_NAME)/internal/cli.Version=$(VERSION)

.PHONY: build install test test-race lint tidy vendor run clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/schemathesis

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/schemathesis

test:
	go test ./...

# The runner drives requests through a worker pool, so the race detector matters.
test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

vendor:
	go mod vendor

run:
	go run ./cmd/schemathesis

clean:
	rm -rf bin
