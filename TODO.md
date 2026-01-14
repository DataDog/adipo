# TODO - Future Improvements

## High Priority

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

### Process Name Visibility with memfd
- **Issue**: When using memfd execution, process monitoring tools (`top`, `ps`, APM) display "memfd:filename (deleted)" instead of proper process names
- **Impact**: Breaks process monitoring in production environments (Datadog APM, etc.)
- **Solutions**:
  - Add `prctl(PR_SET_NAME)` in stub to fix process comm name (90% fix, limited to 15 chars)
  - Document `ADIPO_PREFER_DISK=1` for monitored environments
  - Auto-detect Datadog agent and default to disk extraction
  - Consider making disk extraction the default, memfd opt-in
- **Trade-off**: Performance (memfd) vs observability (disk extraction)

## Cross-Platform Support

### macOS Support (Mach-O Format)
- Parse Mach-O headers instead of ELF
- Detect Apple Silicon (M1/M2/M3) vs Intel
- Handle universal binaries interaction
- Memory extraction: Use `mmap` instead of `memfd_create`

### Windows Support (PE Format)
- Parse PE headers
- Detect CPU features via CPUID on Windows
- Extraction: Use temporary files (no memory extraction equivalent)
- Handle .exe and .dll files

### BSD Support
- FreeBSD, OpenBSD, NetBSD
- Similar to Linux but with platform-specific syscalls

## Format Enhancements

### Binary Deduplication
- **Issue**: Same binary compiled with different arch flags often produces identical output (e.g., x86-64-v1/v2/v3 from same source)
- **Solution**: Detect duplicates via SHA-256 checksum during creation, store once and reference
- **Format**: Add `ReferenceIndex` field to BinaryMetadata (use 0xFFFFFFFF for non-references)
- **Benefits**: Massive space savings - 4 identical binaries → 1 stored + 3 references (only ~512 bytes each for metadata)
- **Implementation**:
  - Create: Build checksum map, detect duplicates, store reference index
  - Read: Check if metadata is reference, load from referenced index
  - Inspect: Show "→ references binary X" for duplicates

### Format Version 2.0
- **Shared Compression Dictionary**: Use a shared dictionary for better compression of similar binaries
- **Delta Compression**: Store deltas between similar binaries instead of full copies
- **Streaming Decompression**: Support partial decompression for faster startup
- **Digital Signatures**: Sign binaries for verification
- **Metadata Extensions**: Add build info, dependencies, etc.

### Multi-Container Support
- Support non-ELF containers in the same file format
- Example: ELF + Mach-O + PE all in one fat binary
- Automatic format detection

## Advanced Features

### Auto-Building
```bash
# Build all versions automatically from source
adipo auto-build --source myapp.c --output myapp.fat
```
- Detect optimal compiler flags for each version
- Automatically build v1, v2, v3, v4 variants
- Include ARM variants if cross-compilers available

### Progressive Download Support
- Network-aware fat binaries
- Download only the needed binary on first run
- Cache locally for subsequent runs
- Useful for container images over slow networks

### JIT Version Selection
- Support for JIT-compiled languages
- Select optimized LLVM IR or bytecode based on CPU
- Example: Multiple LLVM IR versions, compile on first run

### Docker Integration
```bash
# Create Docker image with fat binary
adipo docker-build --source Dockerfile --output myapp:latest

# Automatically creates multi-arch image with fat binaries
```

### Kubernetes Operator
- Automatically inject fat binaries into pods
- DaemonSet to detect node capabilities
- Mutating webhook to replace single binaries with fat binaries

## Performance Optimizations

### Stub Size Reduction
- Current: ~2-3 MB
- Goal: < 1 MB
- Techniques:
  - Strip unnecessary dependencies
  - Use TinyGo for stub
  - UPX compression of stub (trade startup time for size)

### Faster Decompression
- Use multiple cores for decompression
- Pre-decompress during idle time (speculative)
- Memory-mapped decompression

### Caching
- Cache decompressed binaries on disk
- Invalidate cache on binary change
- XDG cache directory support

## Developer Experience

### Better Error Messages
- Suggest missing features when no binary matches
- Explain why a binary was selected
- Show performance comparison of available binaries

### Profiling Support
- Built-in profiling mode to compare versions
- `adipo benchmark app.fat` - run all versions and compare
- Generate reports showing performance differences

### IDE Integration
- VSCode extension for inspecting fat binaries
- Show which binary would be selected on current machine
- Visualize binary selection scoring

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

### Video Tutorials
- Introduction to adipo
- Building fat binaries
- Deploying in Kubernetes
- Advanced use cases

## Ecosystem

### Package Manager Integration
- apt/yum repository with fat binaries
- Homebrew formula
- Snap package
- Flatpak package

### CI/CD Integration
- GitHub Actions for building fat binaries
- GitLab CI templates
- CircleCI orb
- Jenkins plugin

### Monitoring Integration
- Metrics export (Prometheus)
- Log which binary was selected
- Performance metrics per version
- Integration with APM tools (DataDog, New Relic)

## Nice to Have

### GUI Tool
- Graphical tool for inspecting fat binaries
- Drag-and-drop to create fat binaries
- Visual diff between binary versions

### Web Service
- Upload binaries, get fat binary back
- SaaS for building fat binaries
- No local tooling required

### Language Bindings
- Python library for creating/inspecting fat binaries
- Rust library
- Node.js library
- Go library (for programmatic use)

### Plugin System
- Custom binary selection algorithms
- Custom compression algorithms
- Custom extraction methods
- Hooks for pre/post execution

## Research Ideas

### Machine Learning Binary Selection
- Learn optimal binary selection from runtime metrics
- Predict best binary based on workload characteristics
- Adaptive selection based on observed performance

### Binary Specialization
- Generate specialized binaries on-the-fly based on actual usage
- Profile-guided optimization at runtime
- Hot code path detection and recompilation

### Zero-Copy Execution
- Execute directly from compressed binary
- Decompress pages on-demand (similar to mmap)
- Reduce memory footprint

---

## Contributing

Have an idea? Open an issue or PR! We'd love to hear your suggestions.

**Priority Guidelines**:
- 🔴 High: Blocking issues or widely requested features
- 🟡 Medium: Nice to have, improves UX
- 🟢 Low: Future enhancements, research ideas
