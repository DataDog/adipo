# Adipo Bazel Examples

This directory contains examples showing how to use Bazel to build fat binaries with adipo.

## Examples

### C Hello World (`c_hello/`)

Demonstrates building a C program with multiple x86-64 micro-architecture optimization levels:
- Builds for x86-64-v2 (SSE4.2) and x86-64-v3 (AVX2)
- Uses conditional compilation to leverage SIMD instructions when available
- Shows the performance difference between optimization levels

**Key concepts:**
- Multi-architecture C compilation with GCC/Clang
- Using `-march` flags for different CPU generations
- Conditional compilation with `#ifdef`

### Go Hello World (`go_hello/`)

Demonstrates building a Go application with multiple x86-64 optimization levels:
- Builds for x86-64-v1, v2, and v3
- Shows how to structure Bazel rules for Go binaries

**Key concepts:**
- Building Go binaries for different architectures
- Using `GOAMD64` environment variable (conceptual)
- Static linking for self-contained binaries

## Getting Started

### Prerequisites

- Bazel 7.0+ (with bzlmod support)
- GCC or Clang (for C examples)
- Go 1.23+ (for Go examples)

### Building Examples

```bash
# Build everything
bazel build //examples/...

# Build specific example
bazel build //examples/c_hello:hello_fat
bazel build //examples/go_hello:hello_fat

# Run an example
bazel run //examples/c_hello:hello_fat
```

### Inspecting Fat Binaries

```bash
# Build the adipo tool
bazel build //cmd/adipo

# Inspect a fat binary
bazel-bin/cmd/adipo/adipo_/adipo inspect bazel-bin/examples/c_hello/hello_fat

# Extract binaries
bazel-bin/cmd/adipo/adipo_/adipo extract --all -o /tmp/extracted/ bazel-bin/examples/c_hello/hello_fat
```

## Using Adipo Rules in Your Project

To use the adipo Bazel rules in your own project:

1. Add adipo as a dependency in your `MODULE.bazel`:
```python
bazel_dep(name = "adipo", version = "1.0.0")
# Or use a local override during development:
# local_path_override(
#     module_name = "adipo",
#     path = "../adipo",
# )
```

2. Load the rules in your `BUILD.bazel`:
```python
load("@adipo//bazel:adipo.bzl", "adipo_fat_binary")
```

3. Create fat binaries:
```python
# For C/C++ binaries
cc_binary(
    name = "myapp_v2",
    srcs = ["main.c"],
    copts = ["-march=x86-64-v2"],
)

cc_binary(
    name = "myapp_v3",
    srcs = ["main.c"],
    copts = ["-march=x86-64-v3"],
)

adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v2": "x86-64-v2",
        ":myapp_v3": "x86-64-v3",
    },
)

# For Go binaries
go_binary(
    name = "myservice_v1",
    # ... config for v1
)

go_binary(
    name = "myservice_v2",
    # ... config for v2
)

adipo_fat_binary(
    name = "myservice_fat",
    binaries = {
        ":myservice_v1": "x86-64-v1",
        ":myservice_v2": "x86-64-v2",
    },
    compression = "zstd",  # or "lz4", "gzip", "none"
)
```

## Architecture Specifications

When creating fat binaries, you specify the CPU capabilities of each binary:

### x86-64 Levels
- `x86-64-v1`: Baseline (2003+, all x86-64 CPUs)
- `x86-64-v2`: SSE4.2, POPCNT (2009+, most CPUs)
- `x86-64-v3`: AVX2, BMI2 (2013+, newer CPUs)
- `x86-64-v4`: AVX-512 (2017+, high-end CPUs)

### ARM64 Versions
- `aarch64-v8.0`: Baseline ARM64
- `aarch64-v8.1`: Adds atomics, RDMA
- `aarch64-v8.2`: Adds SVE
- `aarch64-v9.0`: Adds SVE2

### Additional Features

You can also specify additional CPU features:
```python
"x86-64-v3,avx2,fma"
"aarch64-v8.2,sve,bf16"
```

## Benefits

Using adipo with Bazel provides:

1. **Single Build Artifact**: One fat binary instead of multiple images/binaries
2. **Automatic Optimization**: CPU-specific optimizations without deployment complexity
3. **Hermetic Builds**: Bazel's hermetic build system ensures reproducibility
4. **Incremental Builds**: Bazel's caching speeds up rebuilds
5. **Easy Testing**: Test all optimization levels in CI/CD

## Troubleshooting

### "No such file or directory: stub.bin"

The stub binary needs to be built first. Build it explicitly:
```bash
bazel build //:stub_bin
```

### Platform Compatibility Issues

The examples are configured for Linux x86-64. For other platforms:
- Modify `target_compatible_with` in BUILD files
- Adjust `goos` and `goarch` for Go binaries
- Use appropriate `-march` values for C/C++

### Cross-Compilation

To cross-compile:
```bash
# For Linux from macOS
bazel build --platforms=@rules_go//go/toolchain:linux_amd64 //examples/go_hello:hello_fat
```

## Learn More

- [Adipo README](../README.md) - Main project documentation
- [TODO](../TODO.md) - Planned features and roadmap
- [Bazel Rules](../bazel/adipo.bzl) - Rule implementation details
