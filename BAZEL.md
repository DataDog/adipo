# Building Adipo with Bazel

This document describes how to build adipo and create fat binaries using Bazel.

## Prerequisites

- Bazel 7.0+ with bzlmod support
- Go 1.24+ (for adipo and Go examples)
- GCC/Clang (for C/C++ examples)
- Linux environment (for best C/C++ example support)

## Quick Start

```bash
# Build adipo
bazel build //cmd/adipo

# Run adipo
bazel run //cmd/adipo -- --version

# Build all examples
bazel build //examples/...

# Build specific example
bazel build //examples/go_hello:hello_fat
```

## Project Structure

```
.
├── MODULE.bazel          # Bazel module configuration (bzlmod)
├── WORKSPACE             # Empty (using bzlmod)
├── .bazelrc              # Bazel configuration
├── BUILD.bazel           # Root build file
├── bazel/
│   ├── BUILD.bazel       # Bazel utilities
│   ├── adipo.bzl         # Custom rules for fat binaries
│   └── workspace_status.sh  # Version stamping script
├── cmd/adipo/            # Main adipo binary
│   └── BUILD.bazel
├── internal/             # Internal packages
│   ├── */BUILD.bazel     # Per-package build files
├── stub/                 # Self-extracting stub
│   └── BUILD.bazel
└── examples/             # Example projects
    ├── c_hello/          # C example with AVX2/SSE
    └── go_hello/         # Go example
```

## Building Adipo

The adipo binary is built in two stages:

1. **Stub binary**: A small self-extracting Go binary that gets embedded
2. **Main binary**: The full adipo CLI with the stub embedded

```bash
# Build everything
bazel build //cmd/adipo

# The binary is located at:
# bazel-bin/cmd/adipo/adipo_/adipo
```

### Version Information

Version information is stamped into the binary using workspace status:

```bash
# Check version
bazel run //cmd/adipo -- --version
```

The version is read from git and set in `bazel/workspace_status.sh`.

## Creating Fat Binaries

### Using the Bazel Rule

The `adipo_fat_binary` rule makes it easy to create fat binaries:

```python
load("//bazel:adipo.bzl", "adipo_fat_binary")

adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v1": "x86-64-v1",
        ":myapp_v2": "x86-64-v2",
        ":myapp_v3": "x86-64-v3",
    },
    compression = "zstd",  # or "lz4", "gzip", "none"
)
```

### Rule Parameters

- `binaries`: Dictionary mapping binary targets to architecture specifications
  - Keys: Bazel labels for binary targets (e.g., `:myapp_v1`, `//other:binary`)
  - Values: Architecture specs (e.g., `x86-64-v1`, `aarch64-v8.0`)
- `compression`: Compression algorithm (default: `zstd`)
  - Options: `zstd` (best), `lz4` (fast), `gzip` (standard), `none`
- `lib_path`: Default library path for all binaries (optional)
  - Absolute path prepended to `LD_LIBRARY_PATH` (Linux) or `DYLD_LIBRARY_PATH` (macOS)
- `binary_libs`: Per-binary library paths (optional)
  - Dictionary mapping binary basenames to their library paths
- `auto_lib_path`: Custom template for library path generation (optional)
  - Template variables: `{{.Arch}}`, `{{.Version}}`, `{{.ArchVersion}}`
- `enable_auto_lib`: Enable automatic two-path library generation (default: `False`)
  - Generates: `/opt/<arch>/lib:/usr/lib<width>/glibc-hwcaps/<arch-version>`

### Architecture Specifications

Format: `ARCH-VERSION[,FEATURE...]`

#### x86-64 Levels
- `x86-64-v1`: Baseline (all x86-64 CPUs)
- `x86-64-v2`: SSE4.2, POPCNT (2009+)
- `x86-64-v3`: AVX2, BMI2 (2013+)
- `x86-64-v4`: AVX-512 (2017+)

#### ARM64 Versions
- `aarch64-v8.0`: Baseline ARM64
- `aarch64-v8.1`: Adds atomics
- `aarch64-v8.2`: Adds SVE
- `aarch64-v9.0`: Adds SVE2

### Library Path Support

For binaries that require specific library versions, you can specify library paths that will be prepended to `LD_LIBRARY_PATH` (Linux) or `DYLD_LIBRARY_PATH` (macOS) before execution.

#### Automatic Library Paths (Recommended)

```python
adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v1": "x86-64-v1",
        ":myapp_v2": "x86-64-v2",
        ":myapp_v4": "x86-64-v4",
    },
    enable_auto_lib = True,
)

# Results in:
# myapp_v1 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v1
# myapp_v2 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v2
# myapp_v4 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v4
```

#### Custom Template Paths

```python
adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v1": "x86-64-v1",
        ":myapp_v3": "x86-64-v3",
    },
    auto_lib_path = "/opt/glibc-{{.Version}}/lib",
)

# Results in:
# myapp_v1 → /opt/glibc-v1/lib
# myapp_v3 → /opt/glibc-v3/lib
```

#### Per-Binary Library Paths

```python
adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v1": "x86-64-v1",
        ":myapp_v3": "x86-64-v3",
    },
    binary_libs = {
        "myapp_v1": "/custom/path/v1",
        "myapp_v3": "/custom/path/v3",
    },
)
```

#### Fixed Library Path

```python
adipo_fat_binary(
    name = "myapp_fat",
    binaries = {
        ":myapp_v1": "x86-64-v1",
        ":myapp_v2": "x86-64-v2",
    },
    lib_path = "/opt/myapp/lib",
)
```

## Examples

### Go Binary Example

See [examples/go_hello](examples/go_hello/):

```bash
# Build the fat binary
bazel build //examples/go_hello:hello_fat

# Inspect it
bazel run //cmd/adipo -- inspect bazel-bin/examples/go_hello/hello_fat

# Run it (will auto-select best version for your CPU)
bazel run //examples/go_hello:hello_fat
```

The Go example demonstrates:
- Building the same Go source for different optimization levels
- Creating a fat binary with 3 versions (v1, v2, v3)
- Using the `adipo_fat_binary` rule

### C Binary Example

See [examples/c_hello](examples/c_hello/):

```bash
# Build the fat binary (requires Linux or cross-compilation)
bazel build //examples/c_hello:hello_fat

# Inspect it
bazel run //cmd/adipo -- inspect bazel-bin/examples/c_hello/hello_fat
```

The C example demonstrates:
- Conditional compilation with `#ifdef __AVX2__`
- Using SIMD instructions (AVX2 vs SSE)
- Platform-specific compilation flags
- Static vs dynamic linking per platform

**Note:** The C example requires Linux or a Linux cross-compilation toolchain. On macOS, the `-march=x86-64-v2/v3` flags may not be supported by the system compiler.

## Configuration

### .bazelrc

Key settings in `.bazelrc`:

```
# Use bzlmod
build --enable_bzlmod

# Go proxy (use public proxy)
build --action_env=GOPROXY=https://proxy.golang.org,direct

# Version stamping
build --workspace_status_command=$(pwd)/bazel/workspace_status.sh

# C/C++ cross-compilation support
build --incompatible_enable_cc_toolchain_resolution
```

### MODULE.bazel

Dependencies are managed via bzlmod:

```python
bazel_dep(name = "rules_go", version = "0.50.1")
bazel_dep(name = "gazelle", version = "0.39.1")
bazel_dep(name = "platforms", version = "0.0.10")

# Go dependencies are auto-imported from go.mod
go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")
go_deps.from_file(go_mod = "//:go.mod")
```

## Cross-Platform Builds

### Building for Different Architectures

For Go binaries, use `goarch` and `goos`:

```python
go_binary(
    name = "myapp_linux_amd64",
    goarch = "amd64",
    goos = "linux",
    # ...
)

go_binary(
    name = "myapp_linux_arm64",
    goarch = "arm64",
    goos = "linux",
    # ...
)
```

For C/C++ binaries, use `--platforms`:

```bash
# Build for Linux from macOS
bazel build --platforms=@rules_go//go/toolchain:linux_amd64 //examples/c_hello:hello_fat
```

### Docker-Based Builds

For the most reliable cross-platform builds (especially C/C++):

```bash
docker run --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  gcr.io/bazel-public/bazel:latest \
  bazel build //examples/...
```

## Testing

```bash
# Run all tests
bazel test //...

# Run specific test
bazel test //internal/format:format_test
```

## Cleaning

```bash
# Clean build outputs
bazel clean

# Deep clean (removes all caches)
bazel clean --expunge
```

## Troubleshooting

### "No such package" errors for Go dependencies

If you see errors about missing Go dependencies:

```bash
# Regenerate BUILD files
bazel run //:gazelle

# Update MODULE.bazel with correct use_repo declarations
bazel mod tidy
```

### C/C++ compilation errors on macOS

The `-march=x86-64-v2/v3` flags require newer GCC/Clang versions:
- Use Linux for native builds
- Or use Docker for cross-compilation
- Or install a newer LLVM toolchain on macOS

### Version not showing correctly

The version is stamped at build time:

```bash
# Check workspace status
bazel build --workspace_status_command=bazel/workspace_status.sh //cmd/adipo
```

## Performance

Bazel caches aggressively:

- **Incremental builds**: Only changed targets rebuild
- **Remote caching**: Can be configured for team builds
- **Parallel execution**: Builds targets in parallel

Typical build times:
- Cold build: ~30-60s (downloads dependencies)
- Incremental: ~2-5s (rebuilds only changed files)
- Fat binary creation: ~1-2s (runs adipo create)

## Best Practices

1. **Use `gazelle` to maintain BUILD files**: Run `bazel run //:gazelle` after changing Go code
2. **Keep MODULE.bazel in sync**: Run `bazel mod tidy` after changing dependencies
3. **Version your builds**: The workspace status script automatically stamps git info
4. **Test on target platform**: C/C++ examples work best on Linux
5. **Use specific architectures**: Specify exact arch specs for production binaries

## Integration with Existing Builds

### Using Adipo Rules in Your Project

Add adipo to your `MODULE.bazel`:

```python
# For a published version:
bazel_dep(name = "adipo", version = "1.0.0")

# Or for local development:
local_path_override(
    module_name = "adipo",
    path = "../adipo",
)
```

Then use the rules:

```python
load("@adipo//bazel:adipo.bzl", "adipo_fat_binary")

adipo_fat_binary(
    name = "myapp",
    binaries = {
        ":myapp_v2": "x86-64-v2",
        ":myapp_v3": "x86-64-v3",
    },
)
```

### Combining with Docker

You can use fat binaries inside Docker images:

```dockerfile
FROM alpine:latest
COPY bazel-bin/examples/go_hello/hello_fat /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
```

Build separate images for different architectures (amd64, arm64), each containing a fat binary with micro-architecture variants.

## Further Reading

- [Bazel Documentation](https://bazel.build/)
- [rules_go Documentation](https://github.com/bazelbuild/rules_go)
- [Gazelle Documentation](https://github.com/bazelbuild/bazel-gazelle)
- [Adipo README](README.md)
- [Example Walkthroughs](examples/README.md)
