// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


// go:build integration
// +build integration

package hwcaps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

// TestTemplateEvaluationWithRealCPU tests template evaluation with actual CPU detection
func TestTemplateEvaluationWithRealCPU(t *testing.T) {
	// Detect actual CPU capabilities
	caps, err := cpu.Detect()
	if err != nil {
		t.Fatalf("failed to detect CPU: %v", err)
	}

	t.Logf("Detected CPU: %s %s", caps.Architecture, caps.VersionStr)

	// Get the architecture and version from capabilities
	var arch format.Architecture
	var version format.ArchVersion

	switch caps.Architecture {
	case "x86-64", "amd64":
		arch = format.ArchX86_64
		// Parse version from VersionStr (e.g., "v3" -> X86_64_V3)
		switch caps.VersionStr {
		case "v1":
			version = format.X86_64_V1
		case "v2":
			version = format.X86_64_V2
		case "v3":
			version = format.X86_64_V3
		case "v4":
			version = format.X86_64_V4
		default:
			t.Fatalf("unknown x86-64 version: %s", caps.VersionStr)
		}
	case "aarch64", "arm64":
		arch = format.ArchARM64
		// Parse ARM64 version
		spec, err := format.ParseArchSpec("aarch64-" + caps.VersionStr)
		if err != nil {
			t.Fatalf("failed to parse ARM64 version %s: %v", caps.VersionStr, err)
		}
		version = spec.ArchVersion
	default:
		t.Skip("unsupported architecture for this test")
	}

	// Create temporary directory structure mimicking real system
	tmpDir := t.TempDir()

	// Create directories based on detected architecture
	var testDirs []string
	if arch == format.ArchX86_64 {
		// Create Debian multiarch and RedHat style directories
		testDirs = []string{
			filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps", caps.VersionStr),
			filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-"+caps.VersionStr),
		}
	} else if arch == format.ArchARM64 {
		// Create ARM64 directories
		testDirs = []string{
			filepath.Join(tmpDir, "usr/lib/aarch64-linux-gnu/glibc-hwcaps", caps.VersionStr),
			filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/aarch64-"+caps.VersionStr),
		}
	}

	// Create the directories
	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Test templates with tmpDir prefix
	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
		tmpDir + "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
		tmpDir + "/opt/{{.Arch}}/lib", // This won't exist, testing filtering
	}

	// Create template evaluator
	evaluator, err := NewTemplateEvaluator(arch, version)
	if err != nil {
		t.Fatalf("failed to create template evaluator: %v", err)
	}

	// Evaluate templates
	validPaths := evaluator.EvaluateTemplates(templates)

	// Should find the directories we created
	if len(validPaths) == 0 {
		t.Fatal("expected to find valid paths, got none")
	}

	t.Logf("Found %d valid library paths:", len(validPaths))
	for i, path := range validPaths {
		t.Logf("  %d. %s", i+1, path)

		// Verify path exists
		if !evaluator.pathExists(path) {
			t.Errorf("path %s does not exist but was returned", path)
		}
	}

	// Verify /opt path was filtered out (doesn't exist)
	for _, path := range validPaths {
		if filepath.HasPrefix(path, tmpDir+"/opt/") {
			t.Errorf("non-existent /opt path should not be in results: %s", path)
		}
	}

	// Verify paths are sorted by score (Debian multiarch should be first)
	if arch == format.ArchX86_64 || arch == format.ArchARM64 {
		if len(validPaths) > 0 {
			// First path should be the Debian multiarch path (higher score)
			expectedPrefix := tmpDir + "/usr/lib/"
			if !filepath.HasPrefix(validPaths[0], expectedPrefix) {
				t.Errorf("first path should be Debian multiarch path (starts with %s), got: %s",
					expectedPrefix, validPaths[0])
			}
		}
	}
}

// TestDefaultTemplatesWithRealCPU tests default templates work with real CPU
func TestDefaultTemplatesWithRealCPU(t *testing.T) {
	// Detect actual CPU capabilities
	caps, err := cpu.Detect()
	if err != nil {
		t.Fatalf("failed to detect CPU: %v", err)
	}

	t.Logf("Detected CPU: %s %s", caps.Architecture, caps.VersionStr)

	// Get architecture from capabilities
	var arch format.Architecture
	var version format.ArchVersion

	switch caps.Architecture {
	case "x86-64", "amd64":
		arch = format.ArchX86_64
		switch caps.VersionStr {
		case "v1":
			version = format.X86_64_V1
		case "v2":
			version = format.X86_64_V2
		case "v3":
			version = format.X86_64_V3
		case "v4":
			version = format.X86_64_V4
		default:
			t.Fatalf("unknown x86-64 version: %s", caps.VersionStr)
		}
	case "aarch64", "arm64":
		arch = format.ArchARM64
		spec, err := format.ParseArchSpec("aarch64-" + caps.VersionStr)
		if err != nil {
			t.Fatalf("failed to parse ARM64 version %s: %v", caps.VersionStr, err)
		}
		version = spec.ArchVersion
	default:
		t.Skip("unsupported architecture for this test")
	}

	// Get default templates
	templates := GetDefaultTemplates()

	if len(templates) == 0 {
		t.Fatal("expected default templates, got none")
	}

	t.Logf("Testing %d default templates", len(templates))
	for i, tmpl := range templates {
		t.Logf("  Template %d: %s", i+1, tmpl)
	}

	// Create template evaluator
	evaluator, err := NewTemplateEvaluator(arch, version)
	if err != nil {
		t.Fatalf("failed to create template evaluator: %v", err)
	}

	// Evaluate default templates (will look for real system directories)
	validPaths := evaluator.EvaluateTemplates(templates)

	// On a real system with glibc-hwcaps support, we might find paths
	// But this isn't guaranteed, so we just log what we find
	t.Logf("Found %d library paths on real system:", len(validPaths))
	for i, path := range validPaths {
		t.Logf("  %d. %s", i+1, path)
	}

	// Verify all returned paths actually exist
	for _, path := range validPaths {
		if !evaluator.pathExists(path) {
			t.Errorf("path %s does not exist but was returned", path)
		}
	}
}
