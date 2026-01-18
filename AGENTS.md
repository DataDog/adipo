# AI Code Assistant Instructions

This file provides instructions for AI code assistants (Claude, GitHub Copilot, Cursor, etc.) to help maintain code quality and consistency in the adipo project.

## Project Overview

Before making any changes, read the [README.md](README.md) to understand:
- What adipo does (creates fat binaries with multiple micro-architecture versions)
- How it works (CPU detection, compression, extraction)
- Key features and architecture support

## Code Standards

### File Headers

All new `.go` files must include this header at the top:

```go
// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.
```

### Testing Requirements

**All new code must include tests:**

1. **Unit tests**: Test individual functions and methods
   - Place tests in `*_test.go` files next to the code they test
   - Use table-driven tests for multiple test cases
   - Aim for high code coverage (run `make test-coverage`)

2. **Integration tests**: Test end-to-end functionality
   - Add integration tests to `test/integration/` directory
   - Test fat binary creation, execution, and inspection
   - Test on both supported platforms when possible (Linux, macOS)

3. **Running tests**:
   ```bash
   make test              # Run all unit tests
   make test-coverage     # Run tests with coverage report
   make integration-test  # Run integration tests
   make check             # Run all checks (fmt, vet, lint, test)
   ```

### Code Quality

Before submitting changes, ensure code passes all checks:

```bash
make fmt    # Format code with gofmt
make vet    # Run go vet for potential issues
make lint   # Run golangci-lint for style issues
make check  # Run all checks together
```

### Architecture Support

When adding features, consider both supported architectures:
- **x86-64**: versions v1, v2, v3, v4
- **ARM64/AArch64**: versions v8.0, v8.1, v8.2, v9.0, etc.

Test on both platforms when possible, or clearly document platform-specific code.

### Error Handling

- Return errors rather than panicking
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Check all error returns
- Add helpful error messages for user-facing errors

### Documentation

- Add package-level documentation for new packages
- Document exported functions, types, and constants
- Include examples in documentation for complex features
- Update relevant markdown documentation files

### Maintaining CPU Alias Tables

When adding support for new CPUs, update the alias mapping tables in `internal/cpu/aliases.go`:

**For x86-64 CPUs:**
- Add entries to `X86CPUAliases` with Family:Model mappings
- Find CPU family/model with: `lscpu | grep -E 'Model|Family'` on Linux
- Common patterns:
  - Intel: Family 6, different models for each generation
  - AMD: Family 23 (Zen/Zen+/Zen2), Family 25 (Zen3/Zen4)
- Use `NamePattern` for CPUs with multiple SKUs (e.g., "Platinum", "Gold" for Skylake-AVX512)

**For ARM64 Linux CPUs:**
- Add entries to `ARMCPUAliases` with Implementer:PartNum mappings
- Find values with: `grep -E 'CPU implementer|CPU part' /proc/cpuinfo`
- Format: Implementer 0x41 = ARM Ltd, PartNum identifies specific CPU
- Examples: Neoverse N1 = 0x41:0xd0c, Neoverse V2 = 0x41:0xd4f

**For Apple Silicon CPUs:**
- Add entries to `AppleSiliconAliases` with regex patterns
- Match brand string from: `sysctl -n machdep.cpu.brand_string`
- Pattern matches any variant (base/Pro/Max/Ultra) to same alias
- Example: `Apple M3.*` → "apple-m3"

**Testing:**
1. Test detection: `adipo detect-cpu` on the target CPU
2. Test binary selection with hints in selector tests
3. Test library path priority with hints in hwcaps tests
4. Document new aliases in README.md and LIBRARY_PATHS.md

## Commit Guidelines

**Do NOT mention AI assistants in commit messages.** Write commit messages as if you wrote the code yourself:

```bash
# Good commit message
git commit -m "fix: support x86-64 glibc-hwcaps directory naming"

# Bad commit message
git commit -m "fix: support x86-64 glibc-hwcaps (Claude-assisted)"
```

Follow conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/changes
- `refactor:` for code refactoring

## Development Workflow

1. **Read the code first**: Before modifying a file, read it to understand the existing patterns
2. **Follow existing patterns**: Match the style and structure of surrounding code
3. **Keep changes focused**: Don't refactor unrelated code in the same PR
4. **Test your changes**: Always run tests before submitting
5. **Check for regressions**: Run integration tests to ensure nothing breaks

## Security Considerations

- Validate user input, especially file paths and architecture specifications
- Be careful with file operations (prevent path traversal, symlink attacks)
- Avoid command injection when executing binaries
- Don't log sensitive information

## Contributing Process

Follow the guidelines in [CONTRIBUTING.md](CONTRIBUTING.md):

1. **For new features**: Open an RFC issue first to discuss the approach
2. **For bug fixes**: Check if the issue is already reported and assigned
3. **Sign your commits**: All commits must be signed (see [GitHub's signing guide](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits))
4. **Write tests**: All PRs require tests
5. **Pass CI checks**: Ensure all automated checks pass

## Useful Resources

- [TECHNICAL.md](TECHNICAL.md) - Technical implementation details
- [LIBRARY_PATHS.md](LIBRARY_PATHS.md) - Library path configuration
- [HWCAPS.md](HWCAPS.md) - Hardware capabilities approach
- [TODO.md](TODO.md) - Planned features and improvements

## Questions?

If you're unsure about anything:
1. Check existing code for similar patterns
2. Review the documentation files listed above
3. Ask in your PR or issue for clarification
