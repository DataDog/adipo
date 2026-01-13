.PHONY: all build build-all clean test install help lint fmt vet check release-snapshot release-local

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Default target
all: build

## help: Display this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build for current platform (fast, uses GoReleaser)
build:
	@echo "Building for current platform..."
	@goreleaser build --snapshot --clean --single-target --id adipo --id hwcaps-exec --id adipo-stub
	@echo ""
	@echo "Binaries built in dist/ directory:"
	@ls -lh dist/adipo_*/adipo dist/hwcaps-exec_*/hwcaps-exec dist/adipo-stub_*/adipo-stub-* 2>/dev/null || true

## build-all: Build for all platforms (uses GoReleaser)
build-all:
	@echo "Building for all platforms..."
	@goreleaser build --snapshot --clean
	@echo ""
	@echo "All binaries built in dist/ directory"

## release-snapshot: Test full release process locally (no publish)
release-snapshot:
	@echo "Testing release process..."
	@goreleaser release --snapshot --clean
	@$(MAKE) create-universal-stubs
	@echo ""
	@echo "Release artifacts in dist/ directory"
	@ls -lh dist/*.tar.gz dist/*.txt 2>/dev/null || true

## create-universal-stubs: Create universal stubs archive with all platforms
create-universal-stubs:
	@echo "Creating universal stubs archive..."
	@mkdir -p dist/all-stubs
	@cp dist/adipo-stub_*/adipo-stub-* dist/all-stubs/ 2>/dev/null || true
	@if [ -n "$$(ls -A dist/all-stubs 2>/dev/null)" ]; then \
		VERSION=$$(ls dist/*.tar.gz 2>/dev/null | head -1 | sed 's/.*adipo-//' | sed 's/-darwin.*//' | sed 's/-linux.*//'); \
		tar czf dist/adipo-stubs-$${VERSION}.tar.gz -C dist/all-stubs .; \
		rm -rf dist/all-stubs; \
		echo "Created dist/adipo-stubs-$${VERSION}.tar.gz"; \
	else \
		echo "No stub binaries found in dist/"; \
		rm -rf dist/all-stubs; \
	fi

## release-local: Create release locally with current version (for testing)
release-local:
	@echo "Creating local release (no git tag required)..."
	@goreleaser release --snapshot --clean --skip=publish
	@echo ""
	@echo "Release artifacts in dist/ directory"

## clean: Remove built binaries and dist directory
clean:
	@echo "Cleaning..."
	@rm -rf dist/
	@rm -f adipo hwcaps-exec adipo-stub adipo-stub-*
	@rm -f adipo-darwin-* adipo-linux-*
	@rm -f hwcaps-exec-darwin-* hwcaps-exec-linux-*

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## install: Install from source (dev build, not using GoReleaser)
install:
	@echo "Installing from source..."
	@go install -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'" ./cmd/adipo
	@go install -ldflags="-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'" ./cmd/hwcaps-exec
	@go install -ldflags="-s -w" ./cmd/adipo-stub
	@echo "Installed to $(shell go env GOPATH)/bin/"

## install-snapshot: Install from latest snapshot build
install-snapshot: build
	@echo "Installing from snapshot..."
	@cp dist/adipo_$(shell go env GOOS)_$(shell go env GOARCH)*/adipo $(shell go env GOPATH)/bin/
	@cp dist/hwcaps-exec_$(shell go env GOOS)_$(shell go env GOARCH)*/hwcaps-exec $(shell go env GOPATH)/bin/
	@cp dist/adipo-stub_$(shell go env GOOS)_$(shell go env GOARCH)*/adipo-stub-* $(shell go env GOPATH)/bin/
	@echo "Installed to $(shell go env GOPATH)/bin/"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
		exit 1; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test lint
	@echo "All checks passed!"

## mod-tidy: Tidy go.mod
mod-tidy:
	@echo "Tidying go.mod..."
	@go mod tidy

## goreleaser-check: Validate GoReleaser configuration
goreleaser-check:
	@echo "Checking GoReleaser configuration..."
	@goreleaser check

## integration-test-linux: Build and run integration test for Linux binaries
integration-test-linux:
	@$(MAKE) integration-test-impl \
		TEST_NAME="Linux" \
		TEST_GOOS=linux \
		TEST_GOARCH=amd64 \
		TEST_ARCH_VAR=GOAMD64 \
		TEST_BINARIES="v1:x86-64-v1 v2:x86-64-v2 v3:x86-64-v3 v4:x86-64-v4" \
		NATIVE_OS=Linux \
		NATIVE_ARCH=x86_64

## integration-test-macos: Build and run integration test for macOS binaries (requires Go 1.23+)
integration-test-macos:
	@$(MAKE) integration-test-impl \
		TEST_NAME="macOS" \
		TEST_GOOS=darwin \
		TEST_GOARCH=arm64 \
		TEST_ARCH_VAR=GOARM64 \
		TEST_BINARIES="v8.0:aarch64-v8.0 v8.1:aarch64-v8.1 v8.2:aarch64-v8.2 v9.0:aarch64-v9.0" \
		NATIVE_OS=Darwin \
		NATIVE_ARCH=arm64

# Internal target for integration tests (don't call directly)
integration-test-impl:
	@echo "Running $(TEST_NAME) integration test..."
	@echo "Building adipo for host platform..."
	@$(MAKE) build > /dev/null
	@cp dist/adipo_*/adipo ./adipo
	@cp dist/adipo-stub_*/adipo-stub-* ./adipo-stub 2>/dev/null || true
	@echo "Building test binaries..."
	@mkdir -p test/bin
	@echo 'package main\nimport "fmt"\nfunc main() { fmt.Println("Hello from test binary!") }' > test/bin/hello.go
	@for spec in $(TEST_BINARIES); do \
		level=$${spec%%:*}; \
		GOOS=$(TEST_GOOS) GOARCH=$(TEST_GOARCH) $(TEST_ARCH_VAR)=$$level go build -o test/bin/hello-$$level test/bin/hello.go; \
	done
	@if [ "$$(uname -s)" = "$(NATIVE_OS)" ] && [ "$$(uname -m)" = "$(NATIVE_ARCH)" ]; then \
		echo "Creating fat binary with stub (native platform)..."; \
		binary_args=""; \
		for spec in $(TEST_BINARIES); do \
			level=$${spec%%:*}; archspec=$${spec#*:}; \
			binary_args="$$binary_args --binary test/bin/hello-$$level:$$archspec"; \
		done; \
		./adipo create -o test/bin/hello.fat $$binary_args; \
		echo "Inspecting fat binary..."; \
		./adipo inspect test/bin/hello.fat; \
		echo "Running fat binary..."; \
		./adipo run test/bin/hello.fat; \
		./test/bin/hello.fat; \
	else \
		echo "Cross-platform test: creating fat binary without stub..."; \
		binary_args=""; \
		for spec in $(TEST_BINARIES); do \
			level=$${spec%%:*}; archspec=$${spec#*:}; \
			binary_args="$$binary_args --binary test/bin/hello-$$level:$$archspec"; \
		done; \
		./adipo create --no-stub -o test/bin/hello.fat $$binary_args; \
		echo "Inspecting fat binary..."; \
		./adipo inspect test/bin/hello.fat; \
		echo "Skipping execution tests (not on native platform)"; \
	fi
	@echo "Extracting binary..."
	@./adipo extract -t 0 -o test/bin/hello-extracted test/bin/hello.fat
	@echo "$(TEST_NAME) integration test passed!"
	@rm -rf test/bin adipo adipo-stub
