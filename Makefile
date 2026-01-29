.PHONY: build test lint clean install release help

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go settings
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod
BINARY_NAME = tmux-package-status

# Build flags
LDFLAGS = -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)

# Platforms for cross-compilation
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	freebsd/amd64

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build for current platform
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)/

build-all: ## Build for all platforms
	@mkdir -p bin
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=bin/$(BINARY_NAME)_$${os}_$${arch}; \
		if [ "$$os" = "windows" ]; then output=$${output}.exe; fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) -ldflags "$(LDFLAGS)" -o $$output ./cmd/$(BINARY_NAME)/; \
	done

test: ## Run unit tests
	$(GOTEST) -v -race ./...

test-cover: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: build ## Run integration tests
	@if command -v scrut >/dev/null 2>&1; then \
		PATH="$(PWD):$$PATH" scrut test tests/*.md; \
	else \
		echo "Scrut not installed. Install with: pip install scrut"; \
		exit 1; \
	fi

lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

fmt: ## Format code
	$(GOCMD) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -rf bin/ dist/
	rm -f coverage.out coverage.html

install: build ## Install to ~/.local/bin
	@mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"

uninstall: ## Remove from ~/.local/bin
	rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Removed $(HOME)/.local/bin/$(BINARY_NAME)"

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

release: build-all ## Create release archives
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		binary=bin/$(BINARY_NAME)_$${os}_$${arch}; \
		if [ "$$os" = "windows" ]; then binary=$${binary}.exe; fi; \
		archive=dist/$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}; \
		if [ "$$os" = "windows" ]; then \
			zip $${archive}.zip -j $$binary README.md LICENSE; \
		else \
			tar -czvf $${archive}.tar.gz -C bin $(BINARY_NAME)_$${os}_$${arch} -C .. README.md LICENSE; \
		fi; \
	done
	@cd dist && sha256sum * > checksums.txt
	@echo "Release archives in dist/"

.DEFAULT_GOAL := help
