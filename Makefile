.PHONY: build test fmt lint vet tidy deps cyclo clean install help

BINARY := specsync
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

build:
	@echo "⇄ building"
	@go build $(LDFLAGS) -o $(BINARY) ./cmd/specsync

test:
	@echo "⇄ testing"
	@go test ./... -v

test-race:
	@echo "⇄ testing with race detector"
	@go test ./... -race

fmt:
	@go fmt ./...

lint:
	@echo "⇄ checking code quality"
	@go vet ./...
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "⇄ running golangci-lint"
	@which golangci-lint > /dev/null || (echo "install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run

vet:
	@go vet ./...

tidy:
	@go mod tidy

deps:
	@echo "⇄ downloading dependencies"
	@go mod download
	@go mod verify

cyclo:
	@echo "⇄ checking complexity"
	@which gocyclo > /dev/null || (echo "install: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest" && exit 1)
	@gocyclo -top 10 . | grep "^[2-9][0-9]" || echo "all clear"

clean:
	@rm -f $(BINARY)

install:
	@go install $(LDFLAGS) ./cmd/specsync

all: fmt vet test build

help:
	@echo "make build   - build binary"
	@echo "make test    - run tests"
	@echo "make lint    - vet, fmt check, golangci-lint"
	@echo "make deps    - download and verify modules"
	@echo "make cyclo   - check cyclomatic complexity"
	@echo "make clean   - remove binary"
	@echo "make all     - fmt, vet, test, build"
