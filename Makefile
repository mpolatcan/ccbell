# ccbell Makefile
# Cross-compilation for multiple platforms

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build settings
BINARY_NAME := ccbell
GO_MODULE   := github.com/mpolatcan/ccbell
CMD_PATH    := ./cmd/ccbell
BUILD_DIR   := ./bin
DIST_DIR    := ./dist

# Go settings
GO          := go
GOFLAGS     := -trimpath
LDFLAGS     := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(DATE)

# Platforms for cross-compilation
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

# Colors for output
BLUE  := \033[0;34m
GREEN := \033[0;32m
RESET := \033[0m

.PHONY: all build clean test lint install uninstall dist release checksums help

# Default target
all: build

# Build for current platform
build:
	@echo "$(BLUE)Building $(BINARY_NAME) $(VERSION)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

# Run tests
test:
	@echo "$(BLUE)Running tests...$(RESET)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests passed$(RESET)"

# Run tests with coverage report
coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(RESET)"

# Lint code
lint:
	@echo "$(BLUE)Linting...$(RESET)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet instead"; \
		$(GO) vet ./...; \
	fi
	@echo "$(GREEN)✓ Lint passed$(RESET)"

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(RESET)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(RESET)"
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Cleaned$(RESET)"

# Install to plugin directory
install: build
	@echo "$(BLUE)Installing to plugin directory...$(RESET)"
	@mkdir -p $(HOME)/.claude/plugins/cache/cc-plugins/ccbell/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.claude/plugins/cache/cc-plugins/ccbell/bin/
	chmod +x $(HOME)/.claude/plugins/cache/cc-plugins/ccbell/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to ~/.claude/plugins/cache/cc-plugins/ccbell/bin/$(BINARY_NAME)$(RESET)"

# Uninstall from plugin directory
uninstall:
	@echo "$(BLUE)Uninstalling...$(RESET)"
	rm -f $(HOME)/.claude/plugins/cache/cc-plugins/ccbell/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled$(RESET)"

# Build for all platforms
dist: clean
	@echo "$(BLUE)Building for all platforms...$(RESET)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output_name=$(BINARY_NAME)-$${GOOS}-$${GOARCH}; \
		if [ "$${GOOS}" = "windows" ]; then \
			output_name=$${output_name}.exe; \
		fi; \
		echo "  Building $${output_name}..."; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GO) build $(GOFLAGS) \
			-ldflags "$(LDFLAGS)" \
			-o $(DIST_DIR)/$${output_name} $(CMD_PATH); \
	done
	@echo "$(GREEN)✓ Built all platforms$(RESET)"
	@ls -lh $(DIST_DIR)/

# Generate checksums for release
checksums:
	@echo "$(BLUE)Generating checksums...$(RESET)"
	@cd $(DIST_DIR) && shasum -a 256 * > checksums.txt
	@echo "$(GREEN)✓ Checksums generated$(RESET)"
	@cat $(DIST_DIR)/checksums.txt

# Create release archives
release: dist checksums
	@echo "$(BLUE)Creating release archives...$(RESET)"
	@cd $(DIST_DIR) && for f in $(BINARY_NAME)-*; do \
		if [ -f "$$f" ] && [ "$$f" != "checksums.txt" ]; then \
			name=$${f%.*}; \
			ext=$${f##*.}; \
			if [ "$$ext" = "exe" ]; then \
				zip -q "$${name}.zip" "$$f"; \
				echo "  Created $${name}.zip"; \
			else \
				tar -czf "$${f}.tar.gz" "$$f"; \
				echo "  Created $${f}.tar.gz"; \
			fi; \
		fi; \
	done
	@echo "$(GREEN)✓ Release archives created$(RESET)"
	@ls -lh $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip 2>/dev/null || true

# Quick build and test
check: fmt lint test build
	@echo "$(GREEN)✓ All checks passed$(RESET)"

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

# Development build (faster, no optimizations)
dev:
	@echo "$(BLUE)Development build...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Dev build: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

# Run the binary
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Help
help:
	@echo "ccbell Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build for current platform (default)"
	@echo "  test       Run tests with race detection"
	@echo "  coverage   Run tests and generate coverage report"
	@echo "  lint       Run linter (golangci-lint or go vet)"
	@echo "  fmt        Format code"
	@echo "  clean      Remove build artifacts"
	@echo "  install    Install to ~/.claude/plugins/cache/cc-plugins/ccbell/bin/"
	@echo "  uninstall  Remove from plugin directory"
	@echo "  dist       Build for all platforms"
	@echo "  checksums  Generate SHA256 checksums"
	@echo "  release    Build, checksum, and create archives"
	@echo "  check      Run fmt, lint, test, and build"
	@echo "  dev        Quick development build"
	@echo "  version    Show version info"
	@echo "  help       Show this help"
	@echo ""
	@echo "Platforms: $(PLATFORMS)"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build for current platform"
	@echo "  make test               # Run tests"
	@echo "  make dist               # Cross-compile for all platforms"
	@echo "  make release            # Create release with archives"
	@echo "  make install            # Install to plugin directory"
	@echo "  make run ARGS=stop      # Build and run with arguments"
