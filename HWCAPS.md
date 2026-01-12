# Hardware Capabilities (hwcaps) for Libraries

If you're distributing **shared libraries** instead of executables, you can achieve similar microarchitecture-specific optimization using **hwcaps** (hardware capabilities). This is a feature of the dynamic linker (`ld.so`) on Linux and some Unix systems.

## What are hwcaps?

Hardware capabilities allow the dynamic linker to automatically select optimized library variants based on CPU features at runtime. Unlike adipo (which packages multiple executables), hwcaps works by organizing library files in a specific directory structure.

## How it works

The dynamic linker searches for libraries in directories matching the CPU's capabilities:

```bash
# Standard library location
/usr/lib/libexample.so.1

# Optimized variants in hwcaps subdirectories
/usr/lib/glibc-hwcaps/x86-64-v2/libexample.so.1
/usr/lib/glibc-hwcaps/x86-64-v3/libexample.so.1
/usr/lib/glibc-hwcaps/x86-64-v4/libexample.so.1
```

At runtime, the dynamic linker automatically selects the best available version based on the CPU's capabilities, falling back to the baseline version if no optimized variant matches.

## Directory naming conventions

**x86-64 levels:**
- `x86-64-v2` - Requires SSE4.2, SSSE3, POPCNT
- `x86-64-v3` - Requires AVX, AVX2, BMI1, BMI2, F16C, FMA
- `x86-64-v4` - Requires AVX-512F, AVX-512BW, AVX-512CD, AVX-512DQ, AVX-512VL

**ARM64 (on supported systems):**
- Check your platform's documentation for ARM-specific hwcaps subdirectories
- Note: As of 2026, glibc hwcaps support for ARM64 is limited

## Example setup

```bash
# Build your library for different microarchitectures
GOAMD64=v1 go build -buildmode=c-shared -o libexample-v1.so
GOAMD64=v2 go build -buildmode=c-shared -o libexample-v2.so
GOAMD64=v3 go build -buildmode=c-shared -o libexample-v3.so

# Install to hwcaps directories
install -D libexample-v1.so /usr/lib/libexample.so.1
install -D libexample-v2.so /usr/lib/glibc-hwcaps/x86-64-v2/libexample.so.1
install -D libexample-v3.so /usr/lib/glibc-hwcaps/x86-64-v3/libexample.so.1

# Applications automatically use the best version
ldd myapp
#   libexample.so.1 => /usr/lib/glibc-hwcaps/x86-64-v3/libexample.so.1
```

## When to use hwcaps vs adipo

**Use hwcaps for:**
- Shared libraries (`.so`, `.dylib`)
- System-wide installations
- When you want the dynamic linker to handle selection
- Distribution packages (RPM, DEB)
- Libraries that are used by multiple applications

**Use adipo for:**
- Executables and standalone binaries
- Portable single-file deployments
- Application bundles
- When you don't have root/system access
- Cross-platform distribution
- Binaries that need self-contained deployment

## Combining hwcaps with adipo

You can use both approaches together:
- Use adipo for your executable (contains optimized versions)
- Use hwcaps for shared libraries (system-wide optimization)
- The combination provides optimization at both levels

For binaries that need specific library versions, see [LIBRARY_PATHS.md](LIBRARY_PATHS.md) for information about per-binary library path configuration.

## Learn more

- [glibc hardware capabilities documentation](https://www.gnu.org/software/libc/manual/html_node/Hardware-Capability-Tunables.html)
- [ld.so manual](https://man7.org/linux/man-pages/man8/ld.so.8.html) - See "Hardware capabilities" section
- [Optimizing with x86-64 microarchitecture levels](https://lwn.net/Articles/844831/)
