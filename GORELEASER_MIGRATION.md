# GoReleaser Migration Guide

## Overview

This document explains how migrating to GoReleaser would simplify the adipo project's build and release process.

## Current Setup vs GoReleaser

### Current Complexity

**Makefile (176 lines):**
- Manual multi-arch build logic
- Complex integration test targets
- Stub building for each platform
- Manual stub discovery logic

**GitHub Actions (.github/workflows/release.yaml):**
- Multiple steps for building stubs
- Matrix builds for different platforms
- Manual archive creation
- Manual checksum generation
- Custom release note formatting

**Total maintenance:** ~250 lines across Makefile + workflow

### With GoReleaser

**`.goreleaser.yaml` (150 lines):** Single config file for everything

**GitHub Actions (15 lines):** Simple workflow that just runs GoReleaser

**Makefile (60 lines):** Simplified to development tasks only

**Total maintenance:** ~225 lines, but much clearer and standardized

## Key Benefits

### 1. **Automatic Multi-Arch Builds**

**Before:**
```makefile
build-all-arch:
	@GOOS=linux GOARCH=amd64 go build ...
	@GOOS=linux GOARCH=arm64 go build ...
	@GOOS=darwin GOARCH=amd64 go build ...
	@GOOS=darwin GOARCH=arm64 go build ...
```

**After:**
```yaml
builds:
  - goos: [linux, darwin]
    goarch: [amd64, arm64]
```

### 2. **Automatic Release Creation**

**Before:**
- Manual GitHub CLI commands
- Custom release note generation
- Manual archive creation
- Manual checksum calculation

**After:**
- Automatic GitHub release
- Auto-generated changelog from commits
- Automatic archives with correct naming
- Automatic checksums

### 3. **Stub Distribution**

GoReleaser handles stub building and distribution simply:

```yaml
builds:
  # Build stub with platform name in binary
  - id: adipo-stub
    binary: adipo-stub-{{ .Os }}-{{ .Arch }}

# Include stub in the same archive as adipo
archives:
  - builds: [adipo, hwcaps-exec, adipo-stub]
```

This ensures each archive contains `adipo`, `hwcaps-exec`, and the matching `adipo-stub-{os}-{arch}` for that platform. The adipo binary's auto-discovery finds the stub next to it.

### 4. **Better Changelog Management**

**Before:** Manual changelog in release notes

**After:** Automatic changelog from commit messages using conventional commits:
```
feat: Add hwcaps-exec command
fix: Address compatibility issues
perf: Optimize binary selection
```

GoReleaser groups these into sections automatically.

### 5. **Consistency**

GoReleaser is an industry-standard tool with:
- Extensive documentation
- Large community
- Battle-tested patterns
- Regular updates

## Installation

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser/v2@latest

# Or use the GitHub Action (recommended for CI)
```

## Testing Locally

```bash
# Test release process without publishing
goreleaser release --snapshot --clean

# Build for current platform only (fast)
goreleaser build --snapshot --clean --single-target

# Build for all platforms
goreleaser build --snapshot --clean

# Check configuration
goreleaser check
```

Built binaries will be in `dist/` directory.

## Migration Steps

### Phase 1: Test GoReleaser (Non-Breaking)

1. **Install GoReleaser:**
   ```bash
   brew install goreleaser
   ```

2. **Test build process:**
   ```bash
   goreleaser build --snapshot --clean --single-target
   ```

3. **Verify binaries work:**
   ```bash
   ./dist/adipo_darwin_arm64/adipo --version
   ./dist/hwcaps-exec_darwin_arm64/hwcaps-exec --help
   ```

4. **Test full multi-arch build:**
   ```bash
   goreleaser build --snapshot --clean
   ```

5. **Test release dry-run:**
   ```bash
   goreleaser release --snapshot --clean
   ```

### Phase 2: Create Test Release

1. **Create test tag:**
   ```bash
   git tag v0.6.0-test
   git push origin v0.6.0-test
   ```

2. **Enable new workflow** (rename `release-goreleaser.yaml` to `release.yaml.new` temporarily)

3. **Manually trigger release:**
   ```bash
   GITHUB_TOKEN=your_token goreleaser release --clean
   ```

4. **Verify release artifacts:**
   - Check GitHub release page
   - Download and test archives
   - Verify checksums
   - Check changelog formatting

5. **Delete test tag if successful:**
   ```bash
   gh release delete v0.6.0-test --yes
   git tag -d v0.6.0-test
   git push origin :refs/tags/v0.6.0-test
   ```

### Phase 3: Switch to GoReleaser

1. **Backup old workflow:**
   ```bash
   mv .github/workflows/release.yaml .github/workflows/release.yaml.old
   mv .github/workflows/release-goreleaser.yaml .github/workflows/release.yaml
   ```

2. **Update Makefile:**
   ```bash
   mv Makefile Makefile.old
   mv Makefile.goreleaser Makefile
   ```

3. **Commit changes:**
   ```bash
   git add .goreleaser.yaml .github/workflows/release.yaml Makefile
   git commit -m "feat: Migrate to GoReleaser for releases"
   ```

4. **Create real release:**
   ```bash
   git tag v0.6.0
   git push origin v0.6.0
   ```

5. **Verify and cleanup:**
   - Verify release worked correctly
   - Remove old files:
     ```bash
     git rm .github/workflows/release.yaml.old Makefile.old
     git commit -m "chore: Remove old build configuration"
     ```

## Stub Distribution: How It Works

The stub distribution is straightforward:

1. **GoReleaser builds all binaries** for each platform:
   - `adipo` (main binary)
   - `hwcaps-exec` (standalone tool)
   - `adipo-stub-{os}-{arch}` (stub for creating fat binaries)

2. **Each archive contains all three:**
   - `adipo-linux-amd64.tar.gz` contains: `adipo`, `hwcaps-exec`, `adipo-stub-linux-amd64`
   - `adipo-darwin-arm64.tar.gz` contains: `adipo`, `hwcaps-exec`, `adipo-stub-darwin-arm64`

3. **adipo automatically discovers the stub** next to it:
   - When creating fat binaries, `adipo` looks for `adipo-stub-{os}-{arch}` in the same directory
   - Users can also specify `--stub-path` explicitly if needed

## Rollback Plan

If something goes wrong:

1. **Restore old workflow:**
   ```bash
   mv .github/workflows/release.yaml.old .github/workflows/release.yaml
   ```

2. **Restore old Makefile:**
   ```bash
   mv Makefile.old Makefile
   ```

3. **Delete failed release:**
   ```bash
   gh release delete v0.6.0 --yes
   git tag -d v0.6.0
   git push origin :refs/tags/v0.6.0
   ```

4. **Retry with old system:**
   ```bash
   git tag v0.6.0
   git push origin v0.6.0
   ```

## Future Enhancements with GoReleaser

Once migrated, you can easily add:

### Homebrew Tap

```yaml
brews:
  - name: adipo
    repository:
      owner: DataDog
      name: homebrew-tap
    description: "Architecture-aware fat binaries"
    homepage: "https://github.com/DataDog/adipo"
    install: |
      bin.install "adipo"
      bin.install "hwcaps-exec"
```

Users can then: `brew install datadog/tap/adipo`

### Docker Images

```yaml
dockers:
  - image_templates:
      - "datadog/adipo:{{ .Tag }}-amd64"
      - "datadog/adipo:{{ .Tag }}-arm64"
    build_flag_templates:
      - "--platform=linux/{{ .Arch }}"
```

### AUR Package (Arch Linux)

```yaml
aurs:
  - name: adipo-bin
    homepage: "https://github.com/DataDog/adipo"
    description: "Architecture-aware fat binaries"
    maintainers:
      - 'Your Name <your@email.com>'
```

### Snap Package

```yaml
snapcrafts:
  - name: adipo
    summary: Architecture-aware fat binaries
    description: Create and run fat binaries with multiple architecture versions
    grade: stable
    confinement: classic
```

## Recommended Commit Message Convention

To get the most out of GoReleaser's changelog generation:

```
feat: Add new feature
fix: Bug fix
perf: Performance improvement
docs: Documentation update
chore: Maintenance task
ci: CI/CD changes
test: Test updates
refactor: Code refactoring
```

## Comparison Table

| Feature | Current Setup | With GoReleaser |
|---------|--------------|-----------------|
| Multi-arch builds | Manual in Makefile | Automatic |
| Release creation | Manual with gh CLI | Automatic |
| Changelog | Manual | Auto-generated |
| Archives | Manual tar/zip | Automatic |
| Checksums | Manual | Automatic |
| Stub embedding | Complex Makefile | Hooks |
| Local testing | Full build only | Snapshot builds |
| CI complexity | ~100 lines | ~15 lines |
| Learning curve | Project-specific | Industry standard |
| Extensibility | Add more Make targets | Add GoReleaser features |

## Verdict

**Recommendation:** Migrate to GoReleaser

**Reasons:**
1. ✅ Significantly simpler CI/CD
2. ✅ Industry-standard tool with good docs
3. ✅ Handles stub embedding correctly
4. ✅ Future extensibility (Homebrew, Docker, etc.)
5. ✅ Better release automation
6. ⚠️ Small learning curve, but worth it

**When to migrate:**
- After v0.5.0 release is stable
- When you have time to test thoroughly
- Before adding more platforms/architectures

**Effort:** ~2-4 hours including testing
**Payoff:** Saves hours on every release + better maintainability
