# adipo - Architecture-Aware Fat Binaries

`adipo` creates and runs fat binaries containing multiple versions of the same executable, each optimized for different CPU micro-architectures. Unlike Apple's `lipo` which targets different architectures (x86-64 vs ARM64), `adipo` targets **micro-architecture versions** within the same architecture family.

## Why adipo?

Modern CPUs within the same architecture family have significantly different capabilities:
- **x86-64**: v1 (baseline) → v2 (SSE4.2) → v3 (AVX2) → v4 (AVX-512)
- **ARM64**: v8.0 (baseline) → v8.1 (atomics) → v8.2 (SVE) → v9.0 (SVE2)

A binary compiled for x86-64-v3 with AVX2 can be 20-40% faster than v1, but won't run on older CPUs. `adipo` solves this by packaging all versions together and selecting the best one at runtime.

## Use Cases

### Multi-Cloud & Multi-Instance Type Deployments

When deploying across multiple cloud providers or instance types, you face a dilemma:

**Problem**: Different instance types have different CPU capabilities
- AWS c5 instances: x86-64-v2 (Skylake)
- AWS c7a instances: x86-64-v3 (Zen 3)
- AWS c7i instances: x86-64-v4 (Sapphire Rapids with AVX-512)
- ARM Graviton 2: ARM64 v8.2
- ARM Graviton 3: ARM64 v8.4 with SVE

**Traditional solutions**:
1. **Single binary** (x86-64-v1): Works everywhere but leaves performance on the table
2. **Multiple Docker images**: `myapp:amd64-v2`, `myapp:amd64-v3`, `myapp:arm64-v8`
   - Complex deployment logic to choose the right image
   - More images to build, store, and manage
   - Need orchestration layer to select correct image per node

**adipo solution**:
```bash
# Build once with all optimizations
adipo create -o myapp.fat \
  --binary myapp-v1:x86-64-v1 \
  --binary myapp-v2:x86-64-v2 \
  --binary myapp-v3:x86-64-v3 \
  --binary myapp-v4:x86-64-v4 \
  --binary myapp-arm64:aarch64-v8.0

# Deploy the same binary everywhere
./myapp.fat  # Automatically selects the best version
```

**Benefits**:
- ✅ **Single artifact**: One binary for all instance types
- ✅ **Automatic selection**: No orchestration logic needed
- ✅ **Maximum performance**: Each instance runs the best available version
- ✅ **Simple deployments**: No need to match image tags to instance capabilities

### When to Use Docker Images vs adipo

**Use separate Docker images** (one for amd64, one for arm64) when:
- You need different **architectures** (x86-64 vs ARM64)
- You want container registry's multi-arch manifest support
- You're running on **very** old kernels (pre-3.17) without `memfd_create`

**Use adipo fat binaries** when:
- You want different **micro-architecture versions** (v1 vs v2 vs v3 vs v4)
- You're deploying across heterogeneous instance types
- You want to simplify deployment without orchestration logic
- You want a single artifact for simplified CI/CD

**Best of both worlds**: Use both!
```dockerfile
FROM alpine
COPY myapp.fat /usr/local/bin/myapp
ENTRYPOINT ["/usr/local/bin/myapp"]
```
Build one Docker image for amd64 and one for arm64, but each contains a fat binary with all micro-architecture versions. This gives you:
- Docker's multi-arch manifest for architecture selection
- adipo's runtime selection for micro-architecture optimization

## Installation

### Prerequisites

- Go 1.23 or later (required for ARM64 v8.1+ support via GOARM64)
- Linux or macOS

### Install from Pre-built Releases (Recommended)

Download the appropriate archive for your platform from [GitHub releases](https://github.com/DataDog/adipo/releases).
Each archive contains both `adipo` and the corresponding `adipo-stub` binary:

**Linux AMD64:**
```bash
curl -LO https://github.com/DataDog/adipo/releases/download/v0.3.0/adipo-v0.3.0-linux-amd64.tar.gz
tar xzf adipo-v0.3.0-linux-amd64.tar.gz
# Archive contains: adipo, adipo-stub-linux-amd64
sudo mv adipo /usr/local/bin/
sudo mv adipo-stub-linux-amd64 /usr/local/bin/
```

**macOS ARM64 (Apple Silicon):**
```bash
curl -LO https://github.com/DataDog/adipo/releases/download/v0.3.0/adipo-v0.3.0-darwin-arm64.tar.gz
tar xzf adipo-v0.3.0-darwin-arm64.tar.gz
sudo mv adipo /usr/local/bin/
sudo mv adipo-stub-darwin-arm64 /usr/local/bin/
```

The stub binary enables creating self-extracting fat binaries. Place it in the same directory as `adipo` for automatic discovery.

### Install with Go

```bash
# Install both adipo and adipo-stub
go install github.com/DataDog/adipo/cmd/adipo@latest
go install github.com/DataDog/adipo/cmd/adipo-stub@latest
```

Both will be installed to `$GOPATH/bin` (usually `~/go/bin`). Make sure this directory is in your `PATH`.

**Note:** Both binaries should be in the same directory for automatic stub discovery. When `adipo` creates a fat binary, it looks for `adipo-stub-{os}-{arch}` (e.g., `adipo-stub-linux-amd64`) or a generic `adipo-stub` next to the `adipo` binary.

If the stub is in a different location, use `--stub-path` to specify it explicitly:
```bash
adipo create --stub-path /path/to/adipo-stub -o app.fat app1 app2
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/DataDog/adipo
cd adipo

# Build adipo
make build

# Build stub for current platform
make stub

# Optionally install to /usr/local/bin
sudo mv adipo /usr/local/bin/
sudo mv internal/stub/stub.bin /usr/local/bin/adipo-stub-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')
```

### Alternative Build Systems

For Bazel users, see [BAZEL.md](BAZEL.md) **(experimental)** for instructions on building with Bazel.

### Homebrew (Coming Soon)

```bash
# Not yet available
brew install adipo
```

## Quick Start

### Create a Fat Binary

```bash
# Basic usage with automatic architecture detection
adipo create -o app.fat app-v1 app-v2 app-v3 app-v4

# With explicit architecture specifications
adipo create -o app.fat \
  --binary app-v1:x86-64-v1 \
  --binary app-v2:x86-64-v2 \
  --binary app-v3:x86-64-v3 \
  --binary app-v4:x86-64-v4

# ARM64 example
adipo create -o app.fat \
  --binary app-base:aarch64-v8.0 \
  --binary app-sve:aarch64-v8.2,sve \
  --binary app-sve2:aarch64-v9.0,sve2
```

### Run a Fat Binary

Fat binaries are self-executing:
```bash
./app.fat [args...]
```

Or use the CLI for debugging:
```bash
# Run with verbose output
adipo run --verbose app.fat

# Force specific version
adipo run --force x86-64-v2 app.fat
```

### Inspect a Fat Binary

```bash
# Table format (default)
adipo inspect app.fat

# JSON format
adipo inspect --format json app.fat
```

Example output:
```
Fat Binary: app.fat
Format Version: 1
Stub Size: 2.16 MB
Stub Architecture: aarch64-v8.0
Number of Binaries: 2

┌───────┬──────────────┬─────────┬────────────┬──────────┬────────────┬───────┐
│ Index │ Architecture │ Version │ Features   │ Original │ Compressed │ Ratio │
│ 0 *   │ aarch64      │ v8.0    │ (baseline) │ 2.28 MB  │ 1.36 MB    │ 59.6% │
│ 1     │ x86-64       │ v2      │ (baseline) │ 2.31 MB  │ 1.38 MB    │ 59.8% │
└───────┴──────────────┴─────────┴────────────┴──────────┴────────────┴───────┘

* Preferred binary for current CPU (aarch64 v8.0)
```

## Advanced Features

### Library Path Support

Configure library paths for binaries that require specific library versions. See [LIBRARY_PATHS.md](LIBRARY_PATHS.md) for details.

Quick example:
```bash
# Automatic glibc-hwcaps compatible paths
adipo create -o app.fat --enable-auto-lib \
  --binary app-v1:x86-64-v1 \
  --binary app-v3:x86-64-v3

# Custom template paths
adipo create -o app.fat \
  --auto-lib-path "/opt/glibc-{{.Version}}/lib" \
  --binary app-v1:x86-64-v1 \
  --binary app-v3:x86-64-v3
```

### Stub Discovery

When creating a fat binary, adipo looks for stub binaries in this order:

1. **Explicit path** (if `--stub-path` is provided): Uses the specified stub
2. **Platform-specific stub**: Looks for `adipo-stub-{os}-{arch}` next to adipo binary
3. **Generic stub**: Looks for `adipo-stub` next to adipo binary
4. **Error**: If no stub is found and `--no-stub` is not specified

### Architecture Specification Format

```
ARCH-VERSION[,FEATURE1,FEATURE2,...]
```

**Aliases**:
- `amd64` = `x86-64` (both are accepted)
- `arm64` = `aarch64` (both are accepted)

**Examples**:
```
x86-64-v1           # Baseline x86-64
amd64-v2            # Same as x86-64-v2
x86-64-v3,avx2      # v3 with AVX2 emphasized
x86-64-v4,avx512f   # v4 with specific AVX-512 features
aarch64-v8.0        # Baseline ARM64
arm64-v8.1,crc      # ARM v8.1 with CRC32
aarch64-v9.0,sve2   # ARM v9 with SVE2
```

### Environment Variables

```bash
ADIPO_VERBOSE=1      # Enable verbose output
ADIPO_DEBUG=1        # Enable debug output
ADIPO_FORCE=x86-64-v2  # Force specific version
ADIPO_PREFER_DISK=1  # Use disk instead of memory extraction
```

## Performance

Typical overhead:
- **Startup time**: ~10ms (CPU detection + decompression)
- **Memory**: ~2-3MB stub + decompressed binary size
- **Disk I/O**: None (memory extraction) or one temp file (fallback)

Space efficiency:
- Stub: ~2-3MB
- Compressed binaries: ~65% of original size (zstd)
- Example: 4 versions of a 10MB binary = 2MB stub + 26MB compressed = **28MB total** vs 40MB for separate files

## Limitations

- **Self-extracting stub is architecture-specific**: The embedded stub can only run on its target architecture
  - x86-64 stub works on x86-64 systems only
  - ARM64 stub works on ARM64 systems only
  - **Solution 1**: Use `--no-stub` and extract with `adipo run` or `adipo extract`
  - **Solution 2**: Build separate fat binaries for each main architecture (one with x86-64 stub, one with ARM64 stub)
  - **Solution 3**: Build cross-compiled stubs with `make stub-all-arch` (requires cross-compilation toolchain)

- **Memory extraction is Linux-only**: In-memory extraction uses `memfd_create` (Linux 3.17+)
  - Fallback to disk extraction works on macOS and older Linux kernels
  - No performance impact, just uses a temporary file

- **Supported binary formats**: ELF (Linux) and Mach-O (macOS)
  - PE (Windows) support planned

- **Micro-architecture mixing only**: Can mix x86-64 micro-architectures (v1/v2/v3/v4) and ARM64 versions (v8.x/v9.x)
  - But a single fat binary with both x86-64 AND ARM64 binaries needs architecture-specific stubs
  - Use Docker multi-arch manifests or separate fat binaries for different main architectures

## Documentation

- [LIBRARY_PATHS.md](LIBRARY_PATHS.md) - Per-binary library path configuration
- [TECHNICAL.md](TECHNICAL.md) - Technical implementation details
- [HWCAPS.md](HWCAPS.md) - Alternative approach using hardware capabilities for libraries
- [BAZEL.md](BAZEL.md) - Building with Bazel **(experimental)**
- [TODO.md](TODO.md) - Planned features and improvements

## Contributing

Contributions are welcome! Please open an issue or pull request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Created by Corentin Chary

Inspired by Apple's `lipo` but focused on micro-architecture levels rather than different architectures.

## Related Projects

- [lipo](https://ss64.com/osx/lipo.html) - Apple's tool for creating fat binaries (different architectures)
- [upx](https://upx.github.io/) - Executable packer (compression only, no multi-version support)
- [go-build](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) - Go's `-tags` build system (compile-time selection)
