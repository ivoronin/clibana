PROJECT := $(shell basename $(CURDIR))
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test test-all lint release clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(PROJECT) ./cmd/$(PROJECT)

test:
	go test -race ./...

test-all: lint test
	@echo "All tests passed"

lint:
	golangci-lint run

release:
	goreleaser release --clean

clean:
	rm -rf bin/ dist/
