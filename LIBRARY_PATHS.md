# Library Path Support

Adipo supports automatic library path configuration to ensure binaries find the correct optimized libraries for their target architecture. This is particularly useful when:
- glibc hwcaps doesn't support your architecture (e.g., ARM64 as of today)
- You need custom library directories for optimized builds
- Different binary variants require different library dependencies

## Quick Start

Create a fat binary with automatic library path configuration:

```bash
adipo create -o app.fat --enable-lib-path app-v1 app-v2 app-v3
```

When executed, the fat binary will automatically:
1. Detect the CPU architecture and version
2. Select the best matching binary
3. Evaluate library path templates for your system
4. Set `LD_LIBRARY_PATH` (Linux) or `DYLD_LIBRARY_PATH` (macOS)
5. Execute the binary with optimized library paths

## Template-Based Library Paths

Library paths are stored as **templates** in the fat binary metadata and evaluated at runtime. This allows a single fat binary to work across different Linux distributions (Debian/Ubuntu, RedHat/Fedora) without hardcoding paths at build time.

### Enabling Library Path Configuration

Library path configuration is **disabled by default**. Enable it with `--enable-lib-path`:

```bash
# Use default templates (Debian multiarch + RedHat)
adipo create -o app.fat --enable-lib-path app-v1 app-v2 app-v3
```

### Default Templates

When enabled without custom templates, adipo uses these defaults:

```bash
/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}  # Debian/Ubuntu multiarch
/usr/lib64/glibc-hwcaps/{{.ArchVersion}}                       # RedHat/Fedora lib64
/opt/{{.Arch}}/lib                                              # Custom /opt installations
```

### Template Variables

Templates support these variables:

| Variable | x86-64 v3 Example | ARM64 v9.4 Example | Description |
|----------|-------------------|---------------------|-------------|
| `{{.Arch}}` | `x86-64` | `aarch64` | Base architecture |
| `{{.ArchTriple}}` | `x86_64` | `aarch64` | Debian multiarch triple |
| `{{.Version}}` | `v3` | `v9.4` | Version string |
| `{{.ArchVersion}}` | `x86-64-v3` | `aarch64-v9.4` | Full arch-version |

### Custom Templates

Specify your own templates with `--lib-path-template` (can be repeated):

```bash
adipo create -o app.fat --enable-lib-path \
  --lib-path-template "/opt/glibc-{{.Version}}/lib" \
  --lib-path-template "/custom/{{.ArchVersion}}/lib" \
  app-v1 app-v2
```

## Runtime Evaluation

### How It Works

At runtime, the fat binary:

1. **Detects CPU** - Determines architecture (x86-64/ARM64) and version (v3, v9.0, etc.)
2. **Selects Binary** - Chooses the best matching binary from the fat archive
3. **Expands Templates** - Replaces template variables with actual values
4. **Version Fallback** - Tries current version, then falls back to older versions
5. **Scores Paths** - Ranks paths by template priority, version match, and path patterns
6. **Filters Paths** - Keeps only paths that actually exist on the system
7. **Sets Environment** - Prepends ranked paths to `LD_LIBRARY_PATH`/`DYLD_LIBRARY_PATH`

### Example: x86-64 v3 CPU

For a CPU with x86-64 v3 support, the default templates expand to:

```
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3     (Score: 1250 - Debian multiarch, exact match)
/usr/lib64/glibc-hwcaps/x86-64-v3             (Score: 1230 - RedHat lib64, exact match)
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2     (Score: 1140 - Debian multiarch, v2 fallback)
/usr/lib64/glibc-hwcaps/x86-64-v2             (Score: 1120 - RedHat lib64, v2 fallback)
/opt/x86-64/lib                                (Score: 1080 - Custom /opt)
```

Only paths that exist on disk are included. The fat binary would set:
```bash
LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3:/usr/lib64/glibc-hwcaps/x86-64-v3:...
```

### Example: ARM64 v9.4 with Fuzzy Matching

For ARM64 v9.4, the evaluator tries version fallback:
- v9.4 (exact) → v9.0 → v9 → v8.9 → v8.8 → ... → v8.0

If a directory only has `v9/` (without v9.4 or v9.0), it will match via the v9.0 fallback:

```
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.0  (Score: 1240 - Debian, v9.0 fallback matches v9/)
/usr/lib64/glibc-hwcaps/aarch64-v8.9          (Score: 1020 - RedHat, v8.9 fallback)
```

This **fuzzy matching** solves the problem where binaries compiled for v9.4 can still use libraries in a v9/ directory.

## Scoring System

Paths are scored and ranked to prefer the best matches:

### Scoring Factors

| Factor | Points | Example |
|--------|--------|---------|
| Template priority | 1000, 900, 800... | Earlier templates score higher |
| Exact version match | +200 | CPU v3 + path v3 |
| Version fallback | +90, +80, +70... | Closer versions score higher |
| Debian multiarch pattern | +50 | `/usr/lib/x86_64-linux-gnu/` |
| RedHat lib64 pattern | +30 | `/usr/lib64/` |
| /opt pattern | -20 | Deprioritize custom paths |

### Scoring Examples

**Scenario: x86-64 v3 CPU, template order [Debian multiarch, RedHat lib64, /opt]**

```
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3
  = 1000 (template 0) + 200 (exact) + 50 (multiarch) = 1250

/usr/lib64/glibc-hwcaps/x86-64-v3
  = 900 (template 1) + 200 (exact) + 30 (lib64) = 1130

/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2
  = 1000 (template 0) + 90 (v2 fallback) + 50 (multiarch) = 1140

/opt/x86-64/lib
  = 800 (template 2) - 20 (deprioritize) = 780
```

Result: Paths returned in order `[v3 multiarch, v2 multiarch, v3 lib64, v2 lib64, /opt]`

## Platform Support

| Platform | Environment Variable | Notes |
|----------|---------------------|-------|
| Linux | `LD_LIBRARY_PATH` | Works with all binaries |
| macOS | `DYLD_LIBRARY_PATH` | SIP-protected binaries ignore this |

- All paths must be absolute (starting with `/`)
- Multiple paths are joined with colon separators (`:`)
- Paths are **prepended** to existing environment variable values

## Inspecting Templates

View stored templates with `inspect`:

```bash
adipo inspect app.fat
```

Output:
```
Library Path Templates:
  Binary 0:
    - /usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}
    - /usr/lib64/glibc-hwcaps/{{.ArchVersion}}
    - /opt/{{.Arch}}/lib
```

## Verbose Execution

See template evaluation in action:

```bash
ADIPO_VERBOSE=1 ./app.fat
```

Output:
```
adipo: Selected binary: x86-64-v3
adipo: Evaluating 3 library path templates...
adipo: Found 4 valid library paths:
  1. /usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3
  2. /usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2
  3. /usr/lib64/glibc-hwcaps/x86-64-v3
  4. /usr/lib64/glibc-hwcaps/x86-64-v2
adipo: Setting LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3:...
adipo: Executing binary...
```

## Distribution Compatibility

The template system automatically adapts to your Linux distribution:

### Debian/Ubuntu
Uses multiarch paths like `/usr/lib/x86_64-linux-gnu/glibc-hwcaps/`

### RedHat/Fedora/CentOS
Uses lib64 paths like `/usr/lib64/glibc-hwcaps/`

### Custom
Falls back to `/opt/` paths or generic `/usr/lib/glibc-hwcaps/`

## Advanced Use Cases

### Custom Directory Layout

```bash
adipo create -o app.fat --enable-lib-path \
  --lib-path-template "/mnt/optimized-libs/{{.ArchVersion}}" \
  app-v1 app-v2
```

### Multiple Custom Paths

```bash
adipo create -o app.fat --enable-lib-path \
  --lib-path-template "/opt/libs1/{{.Version}}" \
  --lib-path-template "/opt/libs2/{{.ArchVersion}}" \
  --lib-path-template "/custom/{{.Arch}}" \
  app-v1 app-v2
```

### Combining with glibc hwcaps

For x86-64, glibc automatically searches hwcaps directories. The template system complements this by:
- Supporting ARM64 (which glibc doesn't handle automatically)
- Adding custom library directories beyond system paths
- Providing explicit control over library search order

## Limitations

- Maximum 3 templates by default (can fit in metadata)
- Library paths cannot exceed 132 bytes total in metadata
- Paths must be absolute (relative paths not supported)
- macOS SIP restrictions apply to system binaries

## Testing

Run integration tests with real CPU detection:

```bash
go test -tags=integration ./internal/hwcaps/...
```

## Implementation Details

- Templates are stored in binary metadata (132-byte reserved field)
- Each template is length-prefixed for efficient parsing
- Version fallback uses explicit ARM64 version list (v9.5 → v8.0)
- Scoring algorithm mirrors binary selection logic
- Template evaluation happens at runtime, not build time
