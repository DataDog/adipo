.PHONY: all build stub clean test install help integration-test-linux integration-test-macos

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'
STUB_LDFLAGS := -s -w

# Go build flags
GOFLAGS := -trimpath
STUBFLAGS := $(GOFLAGS) -ldflags="$(STUB_LDFLAGS)"
MAINFLAGS := $(GOFLAGS) -ldflags="$(LDFLAGS)"

# Output paths
STUB_BIN := internal/stub/stub.bin
MAIN_BIN := adipo

# Default target
all: build

## help: Display this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the adipo binary
build:
	@echo "Building adipo..."
	go build $(MAINFLAGS) -o $(MAIN_BIN) ./cmd/adipo

## stub: Build the self-extracting stub binary for current host
stub:
	@echo "Building stub binary for current host..."
	go build $(STUBFLAGS) -o $(STUB_BIN) ./cmd/adipo-stub

## clean: Remove built binaries
clean:
	@echo "Cleaning..."
	rm -f $(MAIN_BIN) $(STUB_BIN)
	rm -f adipo-stub adipo-stub-*
	rm -f adipo-darwin-* adipo-linux-*
	rm -rf build/

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## install: Install adipo to GOPATH/bin
install:
	@echo "Installing adipo..."
	go install $(MAINFLAGS) ./cmd/adipo

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
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## mod-tidy: Tidy go.mod
mod-tidy:
	@echo "Tidying go.mod..."
	go mod tidy

## build-all-arch: Build for multiple OS/arch combinations
build-all-arch:
	@echo "Building for multiple architectures..."
	@GOOS=linux GOARCH=amd64 go build $(MAINFLAGS) -o $(MAIN_BIN)-linux-amd64 ./cmd/adipo
	@GOOS=linux GOARCH=arm64 go build $(MAINFLAGS) -o $(MAIN_BIN)-linux-arm64 ./cmd/adipo
	@GOOS=darwin GOARCH=amd64 go build $(MAINFLAGS) -o $(MAIN_BIN)-darwin-amd64 ./cmd/adipo
	@GOOS=darwin GOARCH=arm64 go build $(MAINFLAGS) -o $(MAIN_BIN)-darwin-arm64 ./cmd/adipo
	@echo "Built: $(MAIN_BIN)-{linux,darwin}-{amd64,arm64}"

## stub-all-arch: Build stub for multiple architectures (for distribution)
stub-all-arch:
	@echo "Building stub for multiple architectures..."
	@mkdir -p build/stub
	@GOOS=linux GOARCH=amd64 go build $(STUBFLAGS) -o build/stub/adipo-stub-linux-amd64 ./cmd/adipo-stub
	@GOOS=linux GOARCH=arm64 go build $(STUBFLAGS) -o build/stub/adipo-stub-linux-arm64 ./cmd/adipo-stub
	@GOOS=darwin GOARCH=amd64 go build $(STUBFLAGS) -o build/stub/adipo-stub-darwin-amd64 ./cmd/adipo-stub
	@GOOS=darwin GOARCH=arm64 go build $(STUBFLAGS) -o build/stub/adipo-stub-darwin-arm64 ./cmd/adipo-stub
	@echo "Built stubs in build/stub/"

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
	@go build $(MAINFLAGS) -o $(MAIN_BIN) ./cmd/adipo
	@echo "Building stub for host platform..."
	@go build $(STUBFLAGS) -o adipo-stub ./cmd/adipo-stub
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
		./$(MAIN_BIN) create -o test/bin/hello.fat $$binary_args; \
		echo "Inspecting fat binary..."; \
		./$(MAIN_BIN) inspect test/bin/hello.fat; \
		echo "Running fat binary..."; \
		./$(MAIN_BIN) run test/bin/hello.fat; \
		./test/bin/hello.fat; \
	else \
		echo "Cross-platform test: creating fat binary without stub..."; \
		binary_args=""; \
		for spec in $(TEST_BINARIES); do \
			level=$${spec%%:*}; archspec=$${spec#*:}; \
			binary_args="$$binary_args --binary test/bin/hello-$$level:$$archspec"; \
		done; \
		./$(MAIN_BIN) create --no-stub -o test/bin/hello.fat $$binary_args; \
		echo "Inspecting fat binary..."; \
		./$(MAIN_BIN) inspect test/bin/hello.fat; \
		echo "Skipping execution tests (not on native platform)"; \
	fi
	@echo "Extracting binary..."
	@./$(MAIN_BIN) extract -t 0 -o test/bin/hello-extracted test/bin/hello.fat
	@echo "$(TEST_NAME) integration test passed!"
	@rm -rf test/bin adipo-stub

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test
	@echo "All checks passed!"
