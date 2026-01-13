# hwcaps-exec - Hardware Capabilities Executor

## Overview

`hwcaps-exec` executes programs with `LD_LIBRARY_PATH` (Linux) or `DYLD_LIBRARY_PATH` (macOS) configured based on CPU capabilities. It replicates glibc hwcaps functionality for platforms without native support.

The tool automatically:
- Detects CPU architecture and features
- Scans for compatible library directories
- Selects the best available library paths
- Executes the program with the modified environment

Available as both:
- `adipo hwcaps-exec` - subcommand of adipo
- `hwcaps-exec` - standalone binary

## Usage

### Basic Usage

```bash
# Auto-detect CPU and use standard paths
hwcaps-exec myprogram arg1 arg2

# As adipo subcommand
adipo hwcaps-exec myprogram arg1 arg2
```

The tool will automatically scan for compatible libraries and execute your program with the appropriate library paths set.

### Dry Run (Preview)

See what would be executed without actually running the program:

```bash
hwcaps-exec --dry-run myprogram
```

Output example:
```
LD_LIBRARY_PATH=/usr/lib64/glibc-hwcaps/x86-64-v3:/opt/x86-64/lib
[would execute: myprogram]
```

### Verbose Mode

See detailed information about the scanning and selection process:

```bash
hwcaps-exec --verbose myprogram
```

Output example:
```
Detecting CPU capabilities...
CPU: x86-64 v3

Scanning for library directories...

Scanned directories:
  [✗ missing] /usr/lib64/glibc-hwcaps/x86-64-v4 (source: standard-hwcaps, priority: 10004)
  [✓ compatible] /usr/lib64/glibc-hwcaps/x86-64-v3 (source: standard-hwcaps, priority: 10003)
  [✓ compatible] /usr/lib64/glibc-hwcaps/x86-64-v2 (source: standard-hwcaps, priority: 10002)
  [✓ compatible] /opt/x86-64/lib (source: opt-pattern, priority: 103)

Selected paths (2):
  /usr/lib64/glibc-hwcaps/x86-64-v3
  /opt/x86-64/lib

Final LD_LIBRARY_PATH:
  /usr/lib64/glibc-hwcaps/x86-64-v3:/opt/x86-64/lib

Executing: myprogram arg1
```

### Custom Templates

Use custom directory templates with variable expansion:

```bash
hwcaps-exec --lib-path-template "/custom/{{.ArchVersion}}/lib" myprogram
```

Template variables:
- `{{.Arch}}` → Base architecture: "x86-64" or "aarch64"
- `{{.Version}}` → Version only: "v1", "v2", "v8.0", "v9.0", etc.
- `{{.ArchVersion}}` → Full: "x86-64-v1", "aarch64-v8.0", etc.

Example expansions for x86-64 v3 CPU:
- `/opt/{{.ArchVersion}}` → `/opt/x86-64-v3`, `/opt/x86-64-v2`, `/opt/x86-64-v1`
- `/libs/{{.Arch}}/{{.Version}}` → `/libs/x86-64/v3`, `/libs/x86-64/v2`, `/libs/x86-64/v1`

### Additional Directories

Scan additional directories beyond standard paths:

```bash
hwcaps-exec --scan-dir /opt/mylibs --scan-dir /usr/local/mylibs myprogram
```

The `--scan-dir` flag can be repeated multiple times.

### Selective Scanning

Control which directory types are scanned:

```bash
# Disable standard glibc-hwcaps paths
hwcaps-exec --include-standard-hwcaps=false myprogram

# Disable /opt pattern paths
hwcaps-exec --include-opt-pattern=false myprogram

# Only use custom directories
hwcaps-exec --include-standard-hwcaps=false --include-opt-pattern=false \
  --scan-dir /my/custom/libs myprogram
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--lib-path-template` | string | "" | Template with {{.Arch}}, {{.Version}}, {{.ArchVersion}} variables |
| `--scan-dir` | string[] | [] | Additional directories to scan (can be repeated) |
| `--include-standard-hwcaps` | bool | true | Scan /usr/lib64/glibc-hwcaps directories |
| `--include-opt-pattern` | bool | true | Scan /opt/<arch>/lib directories |
| `--dry-run` | bool | false | Show LD_LIBRARY_PATH without executing |
| `--verbose` | bool | false | Show detailed scanning process |

## Directory Scanning

### Standard Paths (x86-64)

When `--include-standard-hwcaps` is enabled (default):

```
/usr/lib64/glibc-hwcaps/x86-64-v4
/usr/lib64/glibc-hwcaps/x86-64-v3
/usr/lib64/glibc-hwcaps/x86-64-v2
/usr/lib64/glibc-hwcaps/x86-64-v1
/usr/lib/glibc-hwcaps/x86-64-v4   (32-bit)
/usr/lib/glibc-hwcaps/x86-64-v3
/usr/lib/glibc-hwcaps/x86-64-v2
/usr/lib/glibc-hwcaps/x86-64-v1
```

### Standard Paths (ARM64)

When `--include-standard-hwcaps` is enabled (default):

```
/usr/lib64/glibc-hwcaps/aarch64-v9.5
/usr/lib64/glibc-hwcaps/aarch64-v9.4
... (all versions down to v8.0)
/usr/lib64/glibc-hwcaps/aarch64-v8.0
```

### Opt Pattern Paths

When `--include-opt-pattern` is enabled (default):

```
/opt/x86-64/lib    (for x86-64)
/opt/aarch64/lib   (for ARM64)
```

### Selection Priority

When multiple compatible directories exist, they are prioritized as follows:

1. **Highest compatible version first** - v4 > v3 > v2 > v1
2. **Source type priority** (for same version):
   - Standard glibc-hwcaps paths (highest)
   - Custom template paths
   - Opt pattern paths
   - User-specified directories (lowest)

**All compatible paths are included** in `LD_LIBRARY_PATH` to allow dynamic linker fallback.

## Graceful Degradation

If no compatible library directories are found, the program executes anyway with system defaults:

```bash
$ hwcaps-exec --verbose myprogram
Detecting CPU capabilities...
CPU: x86-64 v3

Scanning for library directories...

Scanned directories:
  [✗ missing] /usr/lib64/glibc-hwcaps/x86-64-v4 (source: standard-hwcaps, priority: 10004)
  [✗ missing] /usr/lib64/glibc-hwcaps/x86-64-v3 (source: standard-hwcaps, priority: 10003)
  ...

Selected paths (0):
  (none - executing with system defaults)

Final LD_LIBRARY_PATH:
  (not set)

Executing: myprogram
```

This ensures the tool doesn't break existing workflows when custom libraries aren't available.

## Integration with adipo

Combine `hwcaps-exec` with adipo fat binaries for a complete architecture-optimized solution:

### Workflow

1. **Create fat binary** with multiple architecture versions:
```bash
adipo create -o myapp.fat \
  --binary myapp-v1:x86-64-v1 \
  --binary myapp-v2:x86-64-v2 \
  --binary myapp-v3:x86-64-v3 \
  --binary myapp-v4:x86-64-v4
```

2. **Extract specific version** for your CPU:
```bash
adipo extract -t 2 -o myapp-v3 myapp.fat
```

3. **Run with hwcaps-selected libraries**:
```bash
hwcaps-exec myapp-v3 arg1 arg2
```

Or combine steps 2 and 3:
```bash
adipo run myapp.fat arg1 arg2  # Extracts and runs best binary
```

### Library Organization

Structure your libraries to match hwcaps conventions:

```
/usr/lib64/glibc-hwcaps/
├── x86-64-v4/
│   └── libmylib.so       # AVX-512 optimized
├── x86-64-v3/
│   └── libmylib.so       # AVX2 optimized
├── x86-64-v2/
│   └── libmylib.so       # SSE4.2 optimized
└── x86-64-v1/
    └── libmylib.so       # Baseline x86-64
```

Or use custom structure:
```
/opt/x86-64/lib/
└── libmylib.so           # Generic optimized version
```

## Platform Support

| Platform | Binary Format | Env Variable | Status |
|----------|---------------|--------------|--------|
| Linux x86-64 | ELF | `LD_LIBRARY_PATH` | ✓ Supported |
| Linux ARM64 | ELF | `LD_LIBRARY_PATH` | ✓ Supported |
| macOS Intel | Mach-O | `DYLD_LIBRARY_PATH` | ✓ Supported |
| macOS Apple Silicon | Mach-O | `DYLD_LIBRARY_PATH` | ✓ Supported |

## Examples

### Example 1: Simple Execution

```bash
hwcaps-exec /usr/local/bin/myapp
```

### Example 2: With Arguments

```bash
hwcaps-exec myapp --config /etc/myapp.conf --verbose
```

### Example 3: Custom Library Location

```bash
hwcaps-exec --lib-path-template "/opt/libs/{{.ArchVersion}}" myapp
```

### Example 4: Multiple Custom Directories

```bash
hwcaps-exec \
  --scan-dir /opt/vendor/libs \
  --scan-dir /usr/local/custom/libs \
  myapp
```

### Example 5: Debugging Library Issues

```bash
# See what libraries would be used
hwcaps-exec --dry-run --verbose myapp

# Run with verbose output
hwcaps-exec --verbose myapp 2>&1 | tee hwcaps-exec.log
```

### Example 6: Integration with Build Systems

```bash
# In a shell script
#!/bin/bash
set -e

# Extract best binary from fat binary
adipo extract --best -o myapp myapp.fat

# Run with optimized libraries
hwcaps-exec --verbose myapp "$@"
```

## Comparison with glibc hwcaps

| Feature | glibc hwcaps | hwcaps-exec |
|---------|--------------|-------------|
| Automatic detection | ✓ | ✓ |
| Standard paths | ✓ | ✓ |
| Custom paths | Limited | ✓ Full templates |
| Platform support | Linux only | Linux + macOS |
| Custom logic | No | Yes |
| Dry-run mode | No | ✓ |
| Verbose output | No | ✓ |

`hwcaps-exec` is designed to complement glibc hwcaps by providing:
- Support for platforms without native hwcaps (e.g., musl libc, macOS)
- Custom directory templates and scanning patterns
- Debugging and preview capabilities
- Integration with adipo fat binaries

## Troubleshooting

### Libraries Not Found

If libraries aren't being detected:

1. **Verify directory structure**:
```bash
ls -la /usr/lib64/glibc-hwcaps/
```

2. **Check CPU detection**:
```bash
adipo cpu
```

3. **Use verbose mode**:
```bash
hwcaps-exec --verbose --dry-run myapp
```

### Wrong Libraries Selected

If incorrect libraries are being used:

1. **Check priority**:
```bash
hwcaps-exec --verbose --dry-run myapp
```

Look at the priority values - higher priority paths are selected first.

2. **Disable unwanted sources**:
```bash
hwcaps-exec --include-opt-pattern=false myapp
```

### Program Fails to Execute

If the program fails after library path setup:

1. **Test without hwcaps-exec**:
```bash
./myapp  # Does it work without library path modification?
```

2. **Check library compatibility**:
```bash
ldd myapp
```

3. **Try system defaults**:
```bash
hwcaps-exec --include-standard-hwcaps=false --include-opt-pattern=false myapp
```

## See Also

- [HWCAPS.md](HWCAPS.md) - Hardware capabilities overview for library authors
- [LIBRARY_PATHS.md](LIBRARY_PATHS.md) - Library path support in adipo fat binaries
- [README.md](README.md) - Main adipo documentation
