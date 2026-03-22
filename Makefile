# rift Makefile

BINARY    := rift
MODULE    := github.com/Ommanimesh2/rift
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
  -X $(MODULE)/cmd.version=$(VERSION) \
  -X $(MODULE)/cmd.commitHash=$(COMMIT) \
  -X $(MODULE)/cmd.buildDate=$(BUILD_DATE)

.PHONY: build test lint cover install clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./... -race

lint:
	golangci-lint run ./...

cover:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY) coverage.out coverage.html
