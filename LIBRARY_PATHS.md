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
3. **Version Fallback** - Creates a list of compatible versions (current, then older versions)
4. **Expands Templates** - For each version, expands all templates with that version's variables
5. **Filters Paths** - Keeps only paths that actually exist on the system
6. **Deduplicates** - Removes duplicate paths that may result from template expansion
7. **Sets Environment** - Prepends paths to `LD_LIBRARY_PATH`/`DYLD_LIBRARY_PATH` in priority order

### Example: x86-64 v3 CPU

For a CPU with x86-64 v3 support, templates expand in version-first order:

```
# Version v3 (exact match) - all templates
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3     (v3, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/x86-64-v3             (v3, template 1: RedHat lib64)
/opt/x86-64/lib                                (v3, template 2: /opt)

# Version v2 (fallback) - all templates
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2     (v2, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/x86-64-v2             (v2, template 1: RedHat lib64)

# Version v1 (fallback) - all templates
/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v1     (v1, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/x86-64-v1             (v1, template 1: RedHat lib64)
```

Only paths that exist on disk are included. For example, if only v3 and v2 Debian paths exist:
```bash
LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3:/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2
```

**Note on deduplication**: If multiple templates expand to the same physical path, it will only appear once in the final `LD_LIBRARY_PATH`. For example, if both `/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}` and `/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.ArchVersion}}` expand to the same directory for certain version formats, the duplicate is automatically removed.

### Example: ARM64 v9.4 with Version Fallback

For ARM64 v9.4, version fallback order: v9.4 → v9.5 → v9.3 → v9.2 → v9.1 → v9.0 → v8.9 → ... → v8.0

Templates expand in version-first order:

```
# Version v9.4 (exact) - all templates
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.4  (v9.4, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/aarch64-v9.4          (v9.4, template 1: RedHat lib64)
/opt/aarch64/lib                               (v9.4, template 2: /opt)

# Version v9.0 (fallback) - all templates
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.0  (v9.0, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/aarch64-v9.0          (v9.0, template 1: RedHat lib64)

# Version v8.9 (fallback) - all templates
/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v8.9  (v8.9, template 0: Debian multiarch)
/usr/lib64/glibc-hwcaps/aarch64-v8.9          (v8.9, template 1: RedHat lib64)
...
```

This ensures binaries compiled for v9.4 can use libraries in v9.0 or older directories.

## Priority Order

Library paths are prioritized by:
1. **Version match** - Exact version matches come first, then fallback to older versions
2. **Template order** - Within each version, templates are evaluated in order
3. **Existence** - Only paths that exist on disk are included

For example, with default templates on x86-64 v3:
1. All v3 paths: Debian multiarch (template 0), RedHat lib64 (template 1), /opt (template 2)
2. All v2 paths: Debian multiarch (template 0), RedHat lib64 (template 1)
3. All v1 paths: Debian multiarch (template 0), RedHat lib64 (template 1)

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

- Library path templates are stored in metadata Reserved field (388 bytes)
- Each template is length-prefixed (2-byte length + template string)
- Paths must be absolute (relative paths not supported)
- macOS SIP restrictions apply to system binaries

## Testing

Run integration tests with real CPU detection:

```bash
go test -tags=integration ./internal/hwcaps/...
```

## Implementation Details

- Binary metadata size: 512 bytes (FormatVersion 1)
- Metadata version: `MetadataVersionV1` indicates template-based library paths
- Templates stored in Reserved field: 388 bytes available
- Each template is length-prefixed (2 bytes) for efficient parsing
- Version fallback uses explicit ARM64 version list (format.ARM64VersionFallbackOrder)
- Paths are collected in version-first order: for each version, expand all templates
- Template evaluation happens at runtime, not build time
- Duplicate paths are automatically removed using a seen map during evaluation
