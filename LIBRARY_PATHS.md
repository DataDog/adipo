# Library Path Support

For binaries that require specific library versions, you can specify library paths that will be prepended to `LD_LIBRARY_PATH` (Linux) or `DYLD_LIBRARY_PATH` (macOS) before execution. This is particularly useful when:
- glibc hwcaps doesn't support your architecture (e.g., ARM64 as of today)
- Your system has older versions of system libraries
- Different binary variants need different library dependencies

## Automatic Library Paths (Default Two-Path Format)

Enable automatic library path generation for all binaries using the standard glibc-hwcaps format:

```bash
adipo create -o app.fat --enable-auto-lib \
  --binary app-v1:x86-64-v1 \
  --binary app-v2:x86-64-v2 \
  --binary app-v4:x86-64-v4

# Results in:
# app-v1 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v1
# app-v2 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v2
# app-v4 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v4
```

For ARM64:
```bash
adipo create -o app.fat --enable-auto-lib \
  --binary app-v80:aarch64-v8.0 \
  --binary app-v90:aarch64-v9.0

# Results in:
# app-v80 → /opt/aarch64/lib:/usr/lib64/glibc-hwcaps/aarch64-v8.0
# app-v90 → /opt/aarch64/lib:/usr/lib64/glibc-hwcaps/aarch64-v9.0
```

This works seamlessly with:
- `/opt/<arch>/lib` - Custom optimized libraries
- `/usr/lib<width>/glibc-hwcaps/<arch-version>` - System glibc-hwcaps directory

## Custom Template Paths

Use templates to generate custom library paths:

```bash
adipo create -o app.fat \
  --auto-lib-path "/opt/glibc-{{.Version}}/lib" \
  --binary app-v1:x86-64-v1 \
  --binary app-v2:x86-64-v2

# Results in:
# app-v1 → /opt/glibc-v1/lib
# app-v2 → /opt/glibc-v2/lib
```

**Template variables:**
- `{{.Arch}}` - Base architecture (e.g., `x86-64`, `aarch64`)
- `{{.Version}}` - Version only (e.g., `v1`, `v4`, `v8.0`)
- `{{.ArchVersion}}` - Full architecture-version (e.g., `x86-64-v4`, `aarch64-v9.0`)

## Per-Binary Library Paths

Override library paths for specific binaries:

```bash
adipo create -o app.fat \
  --binary app-v1:x86-64-v1 --binary-lib app-v1:/custom/path/v1 \
  --binary app-v3:x86-64-v3 --binary-lib app-v3:/custom/path/v3
```

## Fixed Library Path

Set the same library path for all binaries:

```bash
adipo create -o app.fat --lib-path /opt/myapp/lib app-v1 app-v2
```

## Priority Order

When multiple library path options are specified, the priority is:
1. Per-binary specification (`--binary-lib FILE:PATH`)
2. Auto-generated path (`--auto-lib-path template` or `--enable-auto-lib`)
3. Default path (`--lib-path PATH`)

## Platform Support

- **Linux**: Sets `LD_LIBRARY_PATH`
- **macOS**: Sets `DYLD_LIBRARY_PATH` (Note: SIP-protected binaries ignore this)
- Library paths must be absolute (starting with `/`)
- Multiple paths can be specified using colon separators (`:`)
- Paths are prepended to existing environment variable values

## How It Works

When the stub or `adipo run` executes a binary, it:
1. Reads the library path from the selected binary's metadata
2. Prepends it to the existing `LD_LIBRARY_PATH` or `DYLD_LIBRARY_PATH`
3. Executes the binary with the modified environment

For example, if the existing `LD_LIBRARY_PATH=/usr/local/lib` and the binary specifies `/opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v4`, the final value will be:
```
LD_LIBRARY_PATH=/opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v4:/usr/local/lib
```

## Bazel Support

Library path support is also available in Bazel. See [BAZEL.md](BAZEL.md) for details on using library paths in Bazel builds.
