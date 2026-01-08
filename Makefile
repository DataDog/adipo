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

## integration-test-linux: Build and run integration test on Linux
integration-test-linux: build
	@echo "Running Linux integration test..."
	@echo "Building test binaries with different GOAMD64 levels..."
	@mkdir -p test/bin
	@echo 'package main\nimport "fmt"\nfunc main() { fmt.Println("Hello from test binary!") }' > test/bin/hello.go
	GOAMD64=v1 go build -o test/bin/hello-v1 test/bin/hello.go
	GOAMD64=v2 go build -o test/bin/hello-v2 test/bin/hello.go
	GOAMD64=v3 go build -o test/bin/hello-v3 test/bin/hello.go
	GOAMD64=v4 go build -o test/bin/hello-v4 test/bin/hello.go
	@echo "Creating fat binary with 4 x86-64 variants..."
	./$(MAIN_BIN) create -o test/bin/hello.fat \
		--binary test/bin/hello-v1:x86-64-v1 \
		--binary test/bin/hello-v2:x86-64-v2 \
		--binary test/bin/hello-v3:x86-64-v3 \
		--binary test/bin/hello-v4:x86-64-v4
	@echo "Inspecting fat binary..."
	./$(MAIN_BIN) inspect test/bin/hello.fat
	@echo "Running fat binary via adipo run..."
	./$(MAIN_BIN) run test/bin/hello.fat
	@echo "Executing fat binary directly..."
	./test/bin/hello.fat
	@echo "Extracting binary..."
	./$(MAIN_BIN) extract -t 0 -o test/bin/hello-extracted test/bin/hello.fat
	@echo "Linux integration test passed!"
	@rm -rf test/bin

## integration-test-macos: Build and run integration test on macOS (requires Go 1.23+)
integration-test-macos: build
	@echo "Running macOS integration test..."
	@echo "Building test binaries with different GOARM64 levels (requires Go 1.23+)..."
	@mkdir -p test/bin
	@echo 'package main\nimport "fmt"\nfunc main() { fmt.Println("Hello from test binary!") }' > test/bin/hello.go
	GOARM64=v8.0 go build -o test/bin/hello-v80 test/bin/hello.go
	GOARM64=v8.1 go build -o test/bin/hello-v81 test/bin/hello.go
	GOARM64=v8.2 go build -o test/bin/hello-v82 test/bin/hello.go
	GOARM64=v9.0 go build -o test/bin/hello-v90 test/bin/hello.go
	@echo "Creating fat binary with 4 ARM64 variants..."
	./$(MAIN_BIN) create -o test/bin/hello.fat \
		--binary test/bin/hello-v80:aarch64-v8.0 \
		--binary test/bin/hello-v81:aarch64-v8.1 \
		--binary test/bin/hello-v82:aarch64-v8.2 \
		--binary test/bin/hello-v90:aarch64-v9.0
	@echo "Inspecting fat binary..."
	./$(MAIN_BIN) inspect test/bin/hello.fat
	@echo "Running fat binary via adipo run..."
	./$(MAIN_BIN) run test/bin/hello.fat
	@echo "Executing fat binary directly..."
	./test/bin/hello.fat
	@echo "Extracting binary..."
	./$(MAIN_BIN) extract -t 0 -o test/bin/hello-extracted test/bin/hello.fat
	@echo "macOS integration test passed!"
	@rm -rf test/bin

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test
	@echo "All checks passed!"
