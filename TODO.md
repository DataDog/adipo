# TODO - Future Improvements

## High Priority

### bazel integration

- rules to compile a go binary with multiple GOAMD and then bundle it with adipo
- tests for the rules inside the repo

### Config file
- To simplify calls (in particular when running) add /etc/adipo.conf and ~/.local/config/adipo.conf support

### Library Path Support for Sub-Architecture Optimizations
- **Issue**: When using optimized shared libraries (e.g., `libfoo.so.v3` built with AVX2), the selected binary needs to know where to find them
- **Solution**: Adjust `LD_LIBRARY_PATH` based on selected CPU version
- **Example**:
  ```
  /opt/myapp/lib/x86-64-v1/  # Baseline libraries
  /opt/myapp/lib/x86-64-v3/  # AVX2-optimized libraries
  ```
- **Implementation**: Add `--lib-path-template` flag to create command
  ```bash
  adipo create -o app.fat \
    --binary app-v1:x86-64-v1 \
    --binary app-v3:x86-64-v3 \
    --lib-path-template "/opt/myapp/lib/{arch}-{version}"
  ```

### Checksum Verification
- **Issue**: Currently checksums are calculated but not verified at runtime
- **Solution**: Add verification in stub before execution
- **Options**:
  - Fast mode: Verify header checksum only
  - Full mode: Verify each binary checksum
  - Flag: `--skip-verify` to disable for faster startup

## Cross-Platform Support

### macOS Support (Mach-O Format)
- Parse Mach-O headers instead of ELF
- Detect Apple Silicon (M1/M2/M3) vs Intel
- Handle universal binaries interaction
- Memory extraction: Use `mmap` instead of `memfd_create`
- Think about lipo integration

### Windows Support (PE Format)
- Parse PE headers
- Detect CPU features via CPUID on Windows
- Extraction: Use temporary files (no memory extraction equivalent)
- Handle .exe and .dll files

### BSD Support
- FreeBSD, OpenBSD, NetBSD
- Similar to Linux but with platform-specific syscalls

## Format Enhancements

### Format Version 2.0
- **Shared Compression Dictionary**: Use a shared dictionary for better compression of similar binaries
- **Streaming Decompression**: Support partial decompression for faster startup
- **Metadata Extensions**: Add build info, dependencies, etc.

### Multi-Container Support
- Support non-ELF containers in the same file format
- Example: ELF + Mach-O + PE all in one fat binary
- Automatic format detection

## Performance Optimizations

### Faster Decompression
- Use multiple cores for decompression
- Pre-decompress during idle time (speculative)
- Memory-mapped decompression

## Developer Experience

### Better Error Messages
- Suggest missing features when no binary matches
- Explain why a binary was selected
- Show performance comparison of available binaries

## Testing & Quality

### Comprehensive Test Suite
- Unit tests for all packages (currently minimal)
- Integration tests with real binaries
- Cross-platform tests (Linux, macOS, Windows)
- Performance regression tests

### Fuzzing
- Fuzz the format parser
- Fuzz the CPU detection
- Fuzz the compression/decompression

### Benchmarks
- Startup time benchmarks
- Decompression speed benchmarks
- Binary selection speed benchmarks
- Memory usage benchmarks

## Documentation

### Architecture Guide
- Detailed format specification
- CPU detection internals
- Binary selection algorithm explanation

### Best Practices Guide
- When to use fat binaries vs Docker images
- How to build optimized binaries
- Deployment patterns
- Monitoring and observability

## Ecosystem

### Package Manager Integration
- Homebrew formula
- Snap package
- Flatpak package

### Monitoring Integration
- Log which binary was selected
