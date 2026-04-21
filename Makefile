# zvec-go Makefile
# Manages build, test, lint, and development tasks for the zvec Go SDK.

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

SHELL          := /bin/bash
GO             := go
GOFLAGS        ?=
CGO_ENABLED    := 1
BUILD_TAGS     := "source integration"
NPROC          := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2)

# zvec C-API paths
ZVEC_DIR       := zvec
ZVEC_BUILD_DIR := $(ZVEC_DIR)/build
ZVEC_LIB_DIR   := $(ZVEC_BUILD_DIR)/lib

# Tools
GOLANGCI_LINT  := $(shell command -v golangci-lint 2>/dev/null)
GOFUMPT        := $(shell command -v gofumpt 2>/dev/null)

# Go packages (excluding zvec submodule's third-party Go code)
GO_PACKAGES    = $(shell $(GO) list -tags $(BUILD_TAGS) ./... 2>/dev/null | grep -v '/zvec/')

export CGO_ENABLED

# ---------------------------------------------------------------------------
# Default target
# ---------------------------------------------------------------------------

.DEFAULT_GOAL := help

# ---------------------------------------------------------------------------
# Build targets
# ---------------------------------------------------------------------------

.PHONY: build-zvec
build-zvec: ## Build the zvec C-API library from submodule
	@echo "=== Building zvec C-API library ==="
	@git submodule update --init --recursive
	@mkdir -p $(ZVEC_BUILD_DIR)
	@cd $(ZVEC_BUILD_DIR) && \
		CMAKE_COMMON_ARGS="-DCMAKE_BUILD_TYPE=Release -DBUILD_C_BINDINGS=ON -DCMAKE_POLICY_VERSION_MINIMUM=3.5"; \
		if command -v ninja >/dev/null 2>&1 && ninja --version >/dev/null 2>&1; then \
			echo "Using Ninja generator"; \
			cmake .. $$CMAKE_COMMON_ARGS -G Ninja; \
		else \
			echo "Ninja not found, using Unix Makefiles"; \
			cmake .. $$CMAKE_COMMON_ARGS -G "Unix Makefiles"; \
		fi
	@cd $(ZVEC_BUILD_DIR) && cmake --build . -j$(NPROC) --target zvec_c_api
	@echo "✓ C-API library built at $(ZVEC_LIB_DIR)/"
	@ls -lh $(ZVEC_LIB_DIR)/libzvec_c_api.* 2>/dev/null || true

.PHONY: build
build: build-zvec ## Build C-API library and verify Go compilation
	@echo "=== Verifying Go compilation ==="
	@$(GO) build -tags $(BUILD_TAGS) $(GO_PACKAGES)
	@echo "✓ Go compilation successful"

# ---------------------------------------------------------------------------
# Pre-flight check: ensure the C-API library has been built
# ---------------------------------------------------------------------------

define check_zvec_lib
	@if [ ! -d "$(ZVEC_LIB_DIR)" ] || [ -z "$$(ls $(ZVEC_LIB_DIR)/libzvec_c_api.* 2>/dev/null)" ]; then \
		echo ""; \
		echo "✗ zvec C-API library not found at $(ZVEC_LIB_DIR)/"; \
		echo "  Please build it first:  make build-zvec"; \
		echo ""; \
		exit 1; \
	fi
endef

# ---------------------------------------------------------------------------
# Test targets
# ---------------------------------------------------------------------------

.PHONY: test
test: ## Run all Go tests
	$(check_zvec_lib)
	@echo "=== Running tests ==="
	$(GO) test -tags $(BUILD_TAGS) -count=1 -v $(GO_PACKAGES) 2>&1

.PHONY: test-short
test-short: ## Run tests in short mode (skip long-running tests)
	$(check_zvec_lib)
	@echo "=== Running tests (short mode) ==="
	$(GO) test -tags $(BUILD_TAGS) -short -count=1 -v $(GO_PACKAGES) 2>&1

.PHONY: test-race
test-race: ## Run tests with race detector
	$(check_zvec_lib)
	@echo "=== Running tests with race detector ==="
	$(GO) test -tags $(BUILD_TAGS) -race -count=1 -v $(GO_PACKAGES) 2>&1

.PHONY: test-cover
test-cover: ## Run tests with coverage report
	$(check_zvec_lib)
	@echo "=== Running tests with coverage ==="
	$(GO) test -tags $(BUILD_TAGS) -count=1 -coverprofile=coverage.out -covermode=atomic $(GO_PACKAGES) 2>&1
	$(GO) tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML report: go tool cover -html=coverage.out"

.PHONY: bench
bench: ## Run benchmarks
	$(check_zvec_lib)
	@echo "=== Running benchmarks ==="
	$(GO) test -tags $(BUILD_TAGS) -bench=. -benchmem -benchtime=1s -count=1 -run=^$$ $(GO_PACKAGES) 2>&1

.PHONY: fuzz
fuzz: ## Run fuzz tests (default 30s per target)
	$(check_zvec_lib)
	@echo "=== Running fuzz tests ==="
	@FUZZ_TIME=$${FUZZ_TIME:-30s}; \
	for target in $$($(GO) test -tags $(BUILD_TAGS) -list '^Fuzz' $(GO_PACKAGES) 2>/dev/null | grep '^Fuzz'); do \
		echo "--- Fuzzing $$target ($$FUZZ_TIME) ---"; \
		$(GO) test -tags $(BUILD_TAGS) -fuzz=^$$target$$ -fuzztime=$$FUZZ_TIME . 2>&1 || exit 1; \
	done
	@echo "✓ All fuzz tests passed"

# ---------------------------------------------------------------------------
# Lint & format targets
# ---------------------------------------------------------------------------

.PHONY: lint
lint: vet ## Run all linters
ifdef GOLANGCI_LINT
	@echo "=== Running golangci-lint ==="
	$(GOLANGCI_LINT) run .
	@echo "✓ golangci-lint passed"
else
	@echo "⚠ golangci-lint not installed, skipping (install: https://golangci-lint.run/welcome/install/)"
endif

.PHONY: vet
vet: ## Run go vet
	@echo "=== Running go vet ==="
	$(GO) vet -tags $(BUILD_TAGS) $(GO_PACKAGES)
	@echo "✓ go vet passed"

.PHONY: fmt
fmt: ## Format Go source files
	@echo "=== Formatting Go files ==="
ifdef GOFUMPT
	find . -name '*.go' -not -path './zvec/*' | xargs $(GOFUMPT) -w
else
	$(GO) fmt ./...
endif
	@echo "✓ Formatting complete"

.PHONY: fmt-check
fmt-check: ## Check if Go files are formatted (CI-friendly)
	@echo "=== Checking format ==="
	@test -z "$$(find . -name '*.go' -not -path './zvec/*' | xargs gofmt -l 2>/dev/null)" || \
		{ echo "Files need formatting:"; find . -name '*.go' -not -path './zvec/*' | xargs gofmt -l; exit 1; }
	@echo "✓ All files formatted"

# ---------------------------------------------------------------------------
# Sync & update targets
# ---------------------------------------------------------------------------

.PHONY: sync-zvec
sync-zvec: ## Sync zvec submodule to latest main
	./scripts/sync-zvec.sh

.PHONY: sync-zvec-build
sync-zvec-build: ## Sync zvec submodule + rebuild + test
	./scripts/sync-zvec.sh --build

.PHONY: package-libs
package-libs: build-zvec ## Build and package vendor libs for current platform
	@echo "=== Packaging vendor libraries ==="
	./scripts/package-libs.sh
	@echo "✓ Vendor libraries packaged"

.PHONY: check-zvec
check-zvec: ## Check for upstream C-API changes (no update)
	./scripts/sync-zvec.sh --check-only

# ---------------------------------------------------------------------------
# Utility targets
# ---------------------------------------------------------------------------

.PHONY: clean
clean: ## Clean build artifacts
	@echo "=== Cleaning ==="
	@rm -rf $(ZVEC_BUILD_DIR)
	@rm -f coverage.out
	@rm -rf test_*_collection
	@echo "✓ Clean complete"

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "=== Downloading dependencies ==="
	$(GO) mod download
	GOFLAGS="-mod=mod" $(GO) mod tidy -e
	@echo "✓ Dependencies ready"

.PHONY: install-tools
install-tools: ## Install development tools (golangci-lint, gofumpt)
	@echo "=== Installing development tools ==="
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Installing gofumpt..."
	@go install mvdan.cc/gofumpt@latest
	@echo "✓ Tools installed"

.PHONY: all
all: build test lint ## Build, test, and lint (full CI check)
	@echo ""
	@echo "✓ All checks passed"

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------

.PHONY: help
help: ## Show this help message
	@echo "zvec-go Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build-zvec    # Build the C-API library"
	@echo "  make test          # Run all tests"
	@echo "  make lint          # Run linters"
	@echo "  make all           # Full CI check"
	@echo "  make sync-zvec     # Update zvec submodule"
