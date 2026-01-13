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
3. **Expands Templates** - Replaces template variables with actual values for each template in order
4. **Version Fallback** - For each template, tries current version first, then falls back to older versions
5. **Filters Paths** - Keeps only paths that actually exist on the system
6. **Sets Environment** - Prepends paths to `LD_LIBRARY_PATH`/`DYLD_LIBRARY_PATH` in priority order

### Example: x86-64 v3 CPU

For a CPU with x86-64 v3 support, the default templates expand to paths in this priority order:

```
# Template 0 (Debian multiarch): /usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3     (exact match)
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2     (v2 fallback)
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v1     (v1 fallback)

# Template 1 (RedHat lib64): /usr/lib64/glibc-hwcaps/{{.ArchVersion}}
/usr/lib64/glibc-hwcaps/x86-64-v3             (exact match)
/usr/lib64/glibc-hwcaps/x86-64-v2             (v2 fallback)
/usr/lib64/glibc-hwcaps/x86-64-v1             (v1 fallback)

# Template 2 (Custom /opt): /opt/{{.Arch}}/lib
/opt/x86-64/lib                                (no version variants)
```

Only paths that exist on disk are included. For example, if only v3 and v2 paths exist:
```bash
LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3:/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2:/usr/lib64/glibc-hwcaps/x86-64-v3:...
```

### Example: ARM64 v9.4 with Version Fallback

For ARM64 v9.4, the evaluator tries each template with version fallback:
- v9.4 (exact) → v9.5, v9.3, v9.2, v9.1, v9.0 → v8.9 → v8.8 → ... → v8.0

```
# Template 0 (Debian multiarch): /usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.4  (exact match)
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.0  (v9.0 fallback)
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v8.9  (v8.9 fallback)
...

# Template 1 (RedHat lib64): /usr/lib64/glibc-hwcaps/{{.ArchVersion}}
/usr/lib64/glibc-hwcaps/aarch64-v9.4          (exact match)
/usr/lib64/glibc-hwcaps/aarch64-v9.0          (v9.0 fallback)
...
```

This ensures binaries compiled for v9.4 can still use libraries in v9.0 or older directories.

## Priority Order

Library paths are prioritized by:
1. **Template order** - Earlier templates in the list are evaluated first
2. **Version match** - Within each template, exact version matches come before fallback versions
3. **Existence** - Only paths that exist on disk are included

For example, with default templates on x86-64 v3:
1. All Debian multiarch paths (template 0): v3, v2, v1
2. All RedHat lib64 paths (template 1): v3, v2, v1
3. Custom /opt paths (template 2)

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
- Version fallback uses explicit ARM64 version list (format.ARM64VersionFallbackOrder)
- Paths are collected in template order, then version fallback order
- Template evaluation happens at runtime, not build time
