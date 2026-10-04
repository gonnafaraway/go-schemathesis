PROJECT_NAME := $(shell go list -m)
BINARY := bin/schemathesis

.PHONY: build test lint tidy vendor run

build:
	go build -o $(BINARY) ./cmd/schemathesis

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

vendor:
	go mod vendor

run:
	go run ./cmd/schemathesis
