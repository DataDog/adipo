# C Hello World Fat Binary Example

This example demonstrates building a C program with multiple x86-64 micro-architecture levels and packaging them into a fat binary using adipo and Bazel.

## What It Does

The example:
1. Builds the same C source file with two different optimization levels:
   - `x86-64-v2`: Using SSE4.2 instructions (baseline for modern CPUs)
   - `x86-64-v3`: Using AVX2 instructions (faster on newer CPUs)
2. Uses conditional compilation (`#ifdef __AVX2__`) to use SIMD instructions when available
3. Packages both binaries into a single fat binary using adipo
4. At runtime, automatically selects the best version for the CPU

## Building

**Note:** This example requires Linux or a Linux-compatible build environment. The `-march=x86-64-v2` and `-march=x86-64-v3` flags are best supported on Linux with GCC/Clang 11+. On macOS, you may need to use alternative flags or cross-compile for Linux.

**Important:** These targets are tagged as `manual` and won't be built with `bazel test //...` or `bazel build //...` to avoid platform compatibility issues. Build them explicitly:

```bash
# Build individual binaries (on Linux)
bazel build //examples/c_hello:hello_v2
bazel build //examples/c_hello:hello_v3

# Build the fat binary (builds all versions automatically)
bazel build //examples/c_hello:hello_fat

# Run the fat binary (on Linux x86-64)
bazel run //examples/c_hello:hello_fat
```

### Cross-Compiling from macOS

To build for Linux from macOS, you would typically use a cross-compilation toolchain or build in a Linux container:

```bash
# Using Docker
docker run --rm -v $(pwd):/workspace -w /workspace gcr.io/bazel-public/bazel:latest \
  bazel build //examples/c_hello:hello_fat
```

## Inspecting

```bash
# Build adipo first
bazel build //cmd/adipo

# Inspect the fat binary
bazel-bin/cmd/adipo/adipo_/adipo inspect bazel-bin/examples/c_hello/hello_fat
```

## How It Works

The `BUILD.bazel` file defines:
- Two `cc_binary` targets with different `-march` flags
- An `adipo_fat_binary` rule that combines them

The adipo rule:
1. Takes the built binaries
2. Compresses them with zstd
3. Packages them with a self-extracting stub
4. Creates a single executable that auto-selects at runtime

## Key Points

- **Compilation flags**: The `-march=x86-64-v2` and `-march=x86-64-v3` flags tell GCC/Clang which instruction sets to use
- **Static linking**: `-static` ensures the binary has no external dependencies
- **Conditional compilation**: The code uses `#ifdef __AVX2__` to provide optimized implementations
- **Platform constraints**: `target_compatible_with` ensures we only build for Linux x86-64
