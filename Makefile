.PHONY: all build stub clean test install help

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

## build: Build the adipo binary (builds stub first)
build: stub
	@echo "Building adipo..."
	go build $(MAINFLAGS) -o $(MAIN_BIN) ./cmd/adipo

## stub: Build the self-extracting stub binary
stub:
	@echo "Building stub binary..."
	go build $(STUBFLAGS) -o $(STUB_BIN) ./stub

## clean: Remove built binaries
clean:
	@echo "Cleaning..."
	rm -f $(MAIN_BIN) $(STUB_BIN)

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
install: build
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

## build-all-arch: Build for multiple architectures
build-all-arch: stub-all-arch
	@echo "Building for multiple architectures..."
	GOOS=linux GOARCH=amd64 go build $(MAINFLAGS) -o $(MAIN_BIN)-linux-amd64 ./cmd/adipo
	GOOS=linux GOARCH=arm64 go build $(MAINFLAGS) -o $(MAIN_BIN)-linux-arm64 ./cmd/adipo
	@echo "Built: $(MAIN_BIN)-linux-amd64, $(MAIN_BIN)-linux-arm64"

## stub-all-arch: Build stub for multiple architectures
stub-all-arch:
	@echo "Building stub for multiple architectures..."
	@mkdir -p build/stub
	GOOS=linux GOARCH=amd64 go build $(STUBFLAGS) -o build/stub/stub-linux-amd64 ./stub
	GOOS=linux GOARCH=arm64 go build $(STUBFLAGS) -o build/stub/stub-linux-arm64 ./stub
	@echo "Built: build/stub/stub-linux-amd64, build/stub/stub-linux-arm64"

## integration-test: Build and run integration test
integration-test: build
	@echo "Running integration test..."
	@echo "Building test binaries (cross-compiling to Linux/ELF)..."
	@mkdir -p test/bin
	@echo 'package main\nimport "fmt"\nfunc main() { fmt.Println("Hello from test binary!") }' > test/bin/hello.go
	GOOS=linux GOARCH=amd64 go build -o test/bin/hello-v1 test/bin/hello.go
	GOOS=linux GOARCH=amd64 go build -o test/bin/hello-v2 test/bin/hello.go
	@echo "Creating fat binary..."
	./$(MAIN_BIN) create -o test/bin/hello.fat \
		--binary test/bin/hello-v1:x86-64-v1 \
		--binary test/bin/hello-v2:x86-64-v2
	@echo "Inspecting fat binary..."
	./$(MAIN_BIN) inspect test/bin/hello.fat
	@echo "Extracting binary..."
	./$(MAIN_BIN) extract -t 0 -o test/bin/hello-extracted test/bin/hello.fat
	@echo "Integration test passed!"
	@echo "Note: Cannot test execution on non-Linux platforms (fat binaries are Linux/ELF only)"
	@rm -rf test/bin

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test
	@echo "All checks passed!"
