BINARY   := proto2astro
MODULE   := github.com/sarathsp06/proto2astro
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: all build install test vet lint fmt clean snapshot release help

all: vet test build ## Run vet, test, and build

build: ## Build the binary
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/proto2astro

install: ## Install to $GOPATH/bin
	go install -ldflags "$(LDFLAGS)" ./cmd/proto2astro

test: ## Run tests
	go test -race -count=1 ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (must be installed)
	golangci-lint run ./...

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

clean: ## Remove build artifacts
	rm -rf bin/ dist/

snapshot: ## Build a local goreleaser snapshot (no publish)
	goreleaser release --snapshot --clean

release: ## Create a release (triggered by git tag, usually via CI)
	goreleaser release --clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
