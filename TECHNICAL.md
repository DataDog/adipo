# Technical Details

This document describes the technical implementation details of adipo.

## File Format

```
┌─────────────────────────────────┐
│ Self-Extracting Stub (~2-3MB)  │ ← Optional: Pre-compiled Go binary
│                                 │   (Can be omitted with --no-stub)
├─────────────────────────────────┤
│ Magic Marker ("ADIPOFAT")      │
├─────────────────────────────────┤
│ Format Header (260 bytes)      │ ← Version, offsets, metadata,
│                                 │   stub architecture
├─────────────────────────────────┤
│ Binary Metadata Table          │ ← Arch, version, features, sizes
├─────────────────────────────────┤
│ Compressed Binary 0            │ ← zstd/lz4/gzip compressed
├─────────────────────────────────┤
│ Compressed Binary 1            │
├─────────────────────────────────┤
│ ...                            │
└─────────────────────────────────┘
```

## Stub Architecture

The self-extracting stub is a pre-compiled binary that is **architecture-specific**:

- **With stub**: The fat binary is executable on the stub's architecture only
  - x86-64 stub: Can run on x86-64 machines
  - ARM64 stub: Can run on ARM64 machines
  - The stub automatically extracts and executes the best embedded binary

- **Without stub (`--no-stub`)**: The fat binary is not directly executable
  - Saves ~2-3MB of space
  - Requires `adipo extract` or `adipo run` to use
  - Useful for distribution where size matters

### Building Stubs for Different Architectures

The stub is built for your current architecture by default. To build cross-platform fat binaries:

```bash
# Build stubs for all supported architectures
make stub-all-arch

# This creates:
# - build/stub/adipo-stub-linux-amd64
# - build/stub/adipo-stub-linux-arm64
# - build/stub/adipo-stub-darwin-amd64
# - build/stub/adipo-stub-darwin-arm64

# The appropriate stub is selected based on:
# 1. Target OS/arch when creating the fat binary
# 2. Or defaults to the host architecture
```

## Runtime Execution

1. **CPU Detection**: Detect x86-64 level (v1-v4) or ARM version (v8.x, v9.x) and available features
2. **Binary Selection**: Score each binary based on compatibility, version, features, and size
3. **Extraction**: Decompress and extract to memory (Linux `memfd_create`) or disk (fallback)
4. **Execution**: Execute via `fexecve` (memory) or `execve` (disk)

### Binary Selection Algorithm

The stub scores each binary based on:
- **Compatibility**: CPU must support all required features (hard requirement)
- **Version match**: Higher version = better score (prefer newer optimizations)
- **Feature overlap**: More features in common with CPU = better score
- **Size penalty**: Slightly penalize larger binaries (1 point per MB)

The binary with the highest score is selected.

## Compression

Default: **zstd** (level 3)
- Best balance of compression ratio and speed
- Typically 60-70% compression for similar binaries
- Fast decompression (~500 MB/s)

Other options: `lz4` (faster), `gzip` (standard), `none`

## Metadata Format

### Format Version 1

Introduced in v0.5.0 with library path support.

**Changes from version 0**:
- Added `MetadataVersion` uint32 field (4 bytes)
- Reduced `Reserved` space from 136 to 132 bytes
- Library path stored in Reserved[0:130]: 2 bytes length + 128 bytes path

**Binary Metadata Structure** (256 bytes fixed):
```
Offset  Size  Field
0       4     Architecture (CPU architecture enum)
4       4     ArchVersion (microarchitecture version)
8       4     MetadataVersion (format version, v1 = 1)
12      8     RequiredFeatures (primary feature bitmask)
20      32    ExtFeatures (4x 8-byte extended feature bitmasks)
52      8     OriginalSize (uncompressed binary size)
60      8     CompressedSize (compressed binary size)
68      8     DataOffset (offset from start of file)
76      4     Compression (compression algorithm)
80      32    Checksum (SHA-256 of uncompressed binary)
112     4     Priority (selection priority)
116     4     Format (binary format: ELF/Mach-O/PE)
120     132   Reserved (library path + future use)
```

### Reserved Space Layout (Metadata Version 1)

```
Reserved[0:2]     LibraryPathLen uint16  - Length of library path (0 = not set)
Reserved[2:130]   LibraryPath [128]byte  - Null-terminated absolute path
Reserved[130:132] Reserved for future     - 2 bytes for flags/extensions
```

Maximum library path length: 128 bytes

## Environment Variables

```bash
ADIPO_VERBOSE=1      # Enable verbose output
ADIPO_DEBUG=1        # Enable debug output
ADIPO_FORCE=x86-64-v2  # Force specific version
ADIPO_PREFER_DISK=1  # Use disk instead of memory extraction
```

## Platform Support

### Operating Systems
- Linux (fully supported with memory extraction via `memfd_create`)
- macOS (supported with disk fallback)
- Windows (planned)

### Binary Formats
- ELF (Linux) - fully supported
- Mach-O (macOS) - fully supported
- PE (Windows) - planned

### Memory Extraction
- **Linux 3.17+**: Uses `memfd_create` for in-memory extraction (zero disk I/O)
- **macOS/older Linux**: Falls back to temporary file extraction
- No performance difference in practice; disk fallback is transparent

## Performance

Typical overhead:
- **Startup time**: ~10ms (CPU detection + decompression)
- **Memory**: ~2-3MB stub + decompressed binary size
- **Disk I/O**: None (memory extraction) or one temp file (fallback)

Space efficiency:
- Stub: ~2-3MB
- Compressed binaries: ~65% of original size (zstd)
- Example: 4 versions of a 10MB binary = 2MB stub + 26MB compressed = **28MB total** vs 40MB for separate files

## Security Considerations

- **Checksums**: SHA-256 verification of decompressed binaries
- **Format validation**: Strict parsing of metadata and headers
- **No code injection**: All binaries are pre-compiled and checksummed
- **Temporary files**: Cleaned up after execution (configurable)
- **SIP on macOS**: System Integrity Protection may prevent DYLD_LIBRARY_PATH injection for signed system binaries
