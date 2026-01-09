# Release Process

This document describes how to create and publish a new release of adipo.

## Prerequisites

- Push access to the repository
- Ability to create and push tags
- All changes committed and pushed to `main`
- Tests passing on CI

## Release Steps

### 1. Ensure Code Quality

Before creating a release, ensure:
- All tests pass: `make test`
- Code builds successfully: `make build`
- Integration tests pass: `make integration-test-linux` and/or `make integration-test-macos`
- No uncommitted changes: `git status`

### 2. Update Version Information

If applicable, update version strings in:
- `cmd/adipo/main.go` (version variables)
- `README.md` (if version is mentioned)

Commit any version updates:
```bash
git add -A
git commit -m "Bump version to vX.Y.Z"
git push
```

### 3. Create and Push Tag

Create an annotated tag following semantic versioning (vMAJOR.MINOR.PATCH):

```bash
# For a new feature release
git tag -a v0.2.0 -m "Release v0.2.0"

# For a patch release
git tag -a v0.1.1 -m "Release v0.1.1 - Bug fixes"

# For a major release
git tag -a v1.0.0 -m "Release v1.0.0 - First stable release"
```

Push the tag to trigger the release workflow:
```bash
git push origin v0.2.0
```

### 4. Monitor Release Build

The GitHub Actions workflow will automatically:
1. Build binaries for all supported platforms:
   - `linux/amd64`
   - `linux/arm64`
   - `darwin/amd64` (Intel Mac)
   - `darwin/arm64` (Apple Silicon)

2. Create archives with the naming pattern:
   - `adipo-vX.Y.Z-linux-amd64.tar.gz`
   - `adipo-vX.Y.Z-linux-arm64.tar.gz`
   - `adipo-vX.Y.Z-darwin-amd64.tar.gz`
   - `adipo-vX.Y.Z-darwin-arm64.tar.gz`

3. Generate SHA256 checksums for each archive

4. Create a GitHub release with:
   - Release notes with installation instructions
   - All binary archives
   - Checksum files

Monitor the workflow at: `https://github.com/DataDog/adipo/actions`

### 5. Verify Release

Once the workflow completes:

1. Visit the releases page: `https://github.com/DataDog/adipo/releases`
2. Verify all 4 platform archives are attached
3. Verify checksums are present
4. Download and test one of the binaries:

```bash
# Example for Linux AMD64
VERSION=v0.2.0
curl -LO https://github.com/DataDog/adipo/releases/download/${VERSION}/adipo-${VERSION}-linux-amd64.tar.gz
tar xzf adipo-${VERSION}-linux-amd64.tar.gz
./adipo --version
```

### 6. Update Release Notes (Optional)

Edit the release on GitHub to add:
- Detailed changelog
- Breaking changes (if any)
- Notable features
- Known issues
- Migration guide (for major versions)

## Release Checklist

- [ ] All tests pass
- [ ] Code builds successfully
- [ ] Version information updated (if applicable)
- [ ] Changes committed and pushed to main
- [ ] Tag created with appropriate version
- [ ] Tag pushed to GitHub
- [ ] GitHub Actions workflow completed successfully
- [ ] All 4 platform binaries present in release
- [ ] Checksums generated correctly
- [ ] Release notes reviewed and updated
- [ ] At least one binary downloaded and tested

## Versioning Guidelines

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (vX.0.0): Incompatible API changes or major redesign
- **MINOR** version (v0.X.0): New features, backward compatible
- **PATCH** version (v0.0.X): Bug fixes, backward compatible

### When to increment:

**MAJOR (v1.0.0, v2.0.0):**
- Breaking changes to command-line interface
- Incompatible fat binary format changes
- Removal of deprecated features

**MINOR (v0.1.0, v0.2.0):**
- New commands or flags
- New architecture support
- New compression algorithms
- Performance improvements

**PATCH (v0.1.1, v0.1.2):**
- Bug fixes
- Documentation updates
- Security patches
- Build improvements

## Troubleshooting

### Workflow Fails

If the release workflow fails:

1. Check the workflow logs in GitHub Actions
2. Fix the issue locally and commit
3. Delete the tag locally and remotely:
   ```bash
   git tag -d v0.2.0
   git push origin :refs/tags/v0.2.0
   ```
4. Re-create and push the tag after fixing

### Missing Binaries

If some platform binaries are missing:
1. Check the build matrix in `.github/workflows/release.yml`
2. Verify the platform is supported in the build configuration
3. Re-run the workflow from the Actions tab

### Checksum Mismatch

If checksums don't match:
1. Re-download the binary
2. Verify you're using the correct checksum file
3. Report the issue if problem persists

## Post-Release

After a successful release:

1. Announce the release (if applicable):
   - Update project documentation
   - Notify users through appropriate channels
   - Update installation instructions

2. Monitor for issues:
   - Watch GitHub issues for bug reports
   - Check discussion forums
   - Be prepared for quick patch releases if needed

3. Plan next release:
   - Review roadmap
   - Prioritize issues and features
   - Update milestones
