# License Key Verify Tool Makefile

# Project info (falls back to 0.0.0 when no tag exists)
PROJECT_NAME = license-key-verify
VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
BUILD_TIME = $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
ifeq ($(strip $(VERSION)),)
VERSION := 0.0.0
endif

# Go commands
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod

# Build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Output directories
DIST_DIR = dist
BIN_DIR = bin

# Target platforms
PLATFORMS = darwin/amd64,darwin/arm64,linux/amd64,linux/arm64,windows/amd64

# Host executable suffix. Windows hosts (incl. Git Bash) get .exe, but WSL must be
# excluded: it inherits the OS=Windows_NT env var from Windows while uname says Linux.
UNAME_S := $(shell uname -s 2>/dev/null || echo Unknown)
ifneq ($(and $(filter Windows_NT,$(OS)),$(filter-out Linux,$(UNAME_S))),)
HOST_EXE := .exe
endif

.PHONY: all build build-all clean test test-e2e test-coverage deps fmt lint help release

# Default target (test-e2e depends on build, avoiding a double build)
all: clean deps test test-e2e

# Build the binary for the current platform
build:
	@echo "Building lkctl..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/lkctl$(HOST_EXE) ./cmd/lkctl
	@echo "Build complete: $(BIN_DIR)/lkctl$(HOST_EXE)"

# Build binaries for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(shell echo $(PLATFORMS) | tr ',' ' '); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building $$os/$$arch..."; \
		if [ "$$os" = "windows" ]; then \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkctl_$${os}_$${arch}.exe ./cmd/lkctl; \
		else \
			GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/lkctl_$${os}_$${arch} ./cmd/lkctl; \
		fi; \
	done
	@echo "All platforms built!"

# Remove build artifacts (never touches keys/ or *.lic -- those are user assets)
clean:
	@echo "Cleaning build artifacts..."
	-@$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf tests/test_run/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Run unit tests
test:
	@echo "Running unit tests..."
	@$(GOTEST) -v ./...

# Run end-to-end tests
test-e2e: build
	@echo "Running end-to-end tests..."
	@chmod +x tests/e2e_test.sh
	@./tests/e2e_test.sh

# Run tests and generate a coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Create release packages
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)/release
	@for platform in $(shell echo $(PLATFORMS) | tr ',' ' '); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		release_name="$(PROJECT_NAME)_$(VERSION)_$${os}_$${arch}"; \
		echo "Packaging $$release_name..."; \
		mkdir -p $(DIST_DIR)/release/$$release_name; \
		if [ "$$os" = "windows" ]; then \
			if [ -f "$(DIST_DIR)/lkctl_$${os}_$${arch}.exe" ]; then \
				cp $(DIST_DIR)/lkctl_$${os}_$${arch}.exe $(DIST_DIR)/release/$$release_name/lkctl.exe; \
			else \
				echo "Warning: $(DIST_DIR)/lkctl_$${os}_$${arch}.exe not found"; \
			fi; \
		else \
			if [ -f "$(DIST_DIR)/lkctl_$${os}_$${arch}" ]; then \
				cp $(DIST_DIR)/lkctl_$${os}_$${arch} $(DIST_DIR)/release/$$release_name/lkctl; \
				chmod +x $(DIST_DIR)/release/$$release_name/lkctl; \
			else \
				echo "Warning: $(DIST_DIR)/lkctl_$${os}_$${arch} not found"; \
			fi; \
		fi; \
		if [ -f "README.md" ]; then \
			cp README.md $(DIST_DIR)/release/$$release_name/; \
		fi; \
		(cd $(DIST_DIR)/release && zip -r $$release_name.zip $$release_name/ && rm -rf $$release_name); \
	done
	@echo "Release packages created in $(DIST_DIR)/release/"

# Show help
help:
	@echo "License Key Verify Tool - build targets"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - clean, deps, test, e2e (default)"
	@echo "  build         - build the binary for the current platform"
	@echo "  build-all     - build binaries for all platforms"
	@echo "  clean         - remove build artifacts"
	@echo "  test          - run unit tests"
	@echo "  test-e2e      - run end-to-end tests"
	@echo "  test-coverage - run tests and generate a coverage report"
	@echo "  deps          - download dependencies"
	@echo "  fmt           - format code"
	@echo "  lint          - lint code"
	@echo "  release       - create release packages"
	@echo "  help          - show this help"
	@echo ""
	@echo "Project info:"
	@echo "  project: $(PROJECT_NAME)"
	@echo "  version: $(VERSION)"
	@echo "  built:   $(BUILD_TIME)"
	@echo "  commit:  $(GIT_COMMIT)"
