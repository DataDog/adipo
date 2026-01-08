#!/bin/bash
# Workspace status script for stamping version information into binaries

set -euo pipefail

# Get version info
VERSION="${BUILD_SCM_VERSION:-dev}"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Output stable status variables (changes trigger rebuild)
cat <<EOF
STABLE_VERSION ${VERSION}
STABLE_COMMIT ${COMMIT}
STABLE_DATE ${DATE}
EOF
