// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package hwcaps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DataDog/adipo/internal/format"
)

// TestTemplateExpansion tests template variable expansion
func TestTemplateExpansion(t *testing.T) {
	tests := []struct {
		name     string
		arch     format.Architecture
		version  format.ArchVersion
		template string
		expected string
	}{
		{
			name:     "x86-64 v3 with ArchVersion",
			arch:     format.ArchX86_64,
			version:  format.X86_64_V3,
			template: "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
			expected: "/usr/lib64/glibc-hwcaps/x86-64-v3",
		},
		{
			name:     "x86-64 v3 with Version only",
			arch:     format.ArchX86_64,
			version:  format.X86_64_V3,
			template: "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
			expected: "/usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3",
		},
		{
			name:     "ARM64 v8.2 with ArchTriple",
			arch:     format.ArchARM64,
			version:  format.ARM64_V8_2,
			template: "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.ArchVersion}}",
			expected: "/usr/lib/aarch64-linux-gnu/glibc-hwcaps/aarch64-v8.2",
		},
		{
			name:     "ARM64 v9.4 with Version fuzzy matching",
			arch:     format.ArchARM64,
			version:  format.ARM64_V9_4,
			template: "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
			expected: "/usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.4",
		},
		{
			name:     "x86-64 with Arch variable",
			arch:     format.ArchX86_64,
			version:  format.X86_64_V2,
			template: "/opt/{{.Arch}}/lib",
			expected: "/opt/x86-64/lib",
		},
		{
			name:     "ARM64 with Arch variable",
			arch:     format.ArchARM64,
			version:  format.ARM64_V8_0,
			template: "/opt/{{.Arch}}/lib",
			expected: "/opt/aarch64/lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := &TemplateEvaluator{
				arch:    tt.arch,
				version: tt.version,
			}

			result := evaluator.expandTemplate(tt.template, tt.version)
			if result != tt.expected {
				t.Errorf("expandTemplate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestVersionFallbackChain tests version fallback logic
func TestVersionFallbackChain(t *testing.T) {
	tests := []struct {
		name             string
		arch             format.Architecture
		version          format.ArchVersion
		expectedFirst    format.ArchVersion
		expectedLast     format.ArchVersion
		expectedContains []format.ArchVersion
	}{
		{
			name:          "x86-64 v3 fallback",
			arch:          format.ArchX86_64,
			version:       format.X86_64_V3,
			expectedFirst: format.X86_64_V3,
			expectedLast:  format.X86_64_V1,
			expectedContains: []format.ArchVersion{
				format.X86_64_V3,
				format.X86_64_V2,
				format.X86_64_V1,
			},
		},
		{
			name:          "x86-64 v1 no fallback",
			arch:          format.ArchX86_64,
			version:       format.X86_64_V1,
			expectedFirst: format.X86_64_V1,
			expectedLast:  format.X86_64_V1,
			expectedContains: []format.ArchVersion{
				format.X86_64_V1,
			},
		},
		{
			name:          "ARM64 v9.4 fallback to v8",
			arch:          format.ArchARM64,
			version:       format.ARM64_V9_4,
			expectedFirst: format.ARM64_V9_4,
			expectedLast:  format.ARM64_V8,
			expectedContains: []format.ArchVersion{
				format.ARM64_V9_4,
				format.ARM64_V9_0,
				format.ARM64_V9, // glibc-hwcaps alias
				format.ARM64_V8_9,
				format.ARM64_V8_0,
				format.ARM64_V8, // glibc-hwcaps alias
			},
		},
		{
			name:          "ARM64 v9.0 fallback",
			arch:          format.ArchARM64,
			version:       format.ARM64_V9_0,
			expectedFirst: format.ARM64_V9_0,
			expectedLast:  format.ARM64_V8,
			expectedContains: []format.ArchVersion{
				format.ARM64_V9_0,
				format.ARM64_V9, // glibc-hwcaps alias
				format.ARM64_V8_9,
				format.ARM64_V8_0,
				format.ARM64_V8, // glibc-hwcaps alias
			},
		},
		{
			name:          "ARM64 v8.5 fallback",
			arch:          format.ArchARM64,
			version:       format.ARM64_V8_5,
			expectedFirst: format.ARM64_V8_5,
			expectedLast:  format.ARM64_V8,
			expectedContains: []format.ArchVersion{
				format.ARM64_V8_5,
				format.ARM64_V8_4,
				format.ARM64_V8_3,
				format.ARM64_V8_0,
				format.ARM64_V8, // glibc-hwcaps alias
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := &TemplateEvaluator{
				arch:    tt.arch,
				version: tt.version,
			}

			chain := evaluator.getVersionFallbackChain()

			if len(chain) == 0 {
				t.Fatal("version fallback chain is empty")
			}

			if chain[0] != tt.expectedFirst {
				t.Errorf("first version = %v, want %v", chain[0], tt.expectedFirst)
			}

			if chain[len(chain)-1] != tt.expectedLast {
				t.Errorf("last version = %v, want %v", chain[len(chain)-1], tt.expectedLast)
			}

			// Check that all expected versions are present
			chainMap := make(map[format.ArchVersion]bool)
			for _, v := range chain {
				chainMap[v] = true
			}

			for _, expected := range tt.expectedContains {
				if !chainMap[expected] {
					t.Errorf("version chain missing expected version: %v", expected)
				}
			}
		})
	}
}

// TestTemplateEvaluationWithDirectories tests template evaluation with real directories
func TestTemplateEvaluationWithDirectories(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create mock directory structures for x86-64 (using real glibc-hwcaps naming)
	x86v3Path := filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/x86-64-v3")
	x86v2Path := filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v2")

	if err := os.MkdirAll(x86v3Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(x86v2Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Test templates with tmpDir prefix (using ArchVersion for x86-64)
	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.ArchVersion}}",
		tmpDir + "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates(templates, "")

	// Should find both v3 path (exact match) and v2 path (fallback)
	if len(validPaths) != 2 {
		t.Errorf("expected 2 valid paths, got %d: %v", len(validPaths), validPaths)
	}

	// First path should be v3 (higher priority - Debian multiarch + exact match)
	if len(validPaths) > 0 && validPaths[0] != x86v3Path {
		t.Errorf("first path = %s, want %s", validPaths[0], x86v3Path)
	}

	// Second path should be v2 (fallback version)
	if len(validPaths) > 1 && validPaths[1] != x86v2Path {
		t.Errorf("second path = %s, want %s", validPaths[1], x86v2Path)
	}
}

// TestARMVersionFuzzyMatching tests ARM64 v9.x → v9 directory matching
func TestARMVersionFuzzyMatching(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create v9 directory (without minor version) and v8.2 directory
	armV9Path := filepath.Join(tmpDir, "usr/lib/aarch64-linux-gnu/glibc-hwcaps/v9.0")
	armV82Path := filepath.Join(tmpDir, "usr/lib/aarch64-linux-gnu/glibc-hwcaps/v8.2")

	if err := os.MkdirAll(armV9Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(armV82Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Binary is v9.4 but only v9.0 directory exists
	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchARM64,
		version: format.ARM64_V9_4,
	}

	validPaths := evaluator.EvaluateTemplates(templates, "")

	// Should find v9.0 path (via v9.0 fallback from v9.4) and v8.2 path
	if len(validPaths) == 0 {
		t.Fatal("expected at least one valid path, got none")
	}

	// First path should be v9.0 (closer version match)
	if validPaths[0] != armV9Path {
		t.Errorf("first path = %s, want %s", validPaths[0], armV9Path)
	}

	// Should also find v8.2 as fallback
	foundV82 := false
	for _, p := range validPaths {
		if p == armV82Path {
			foundV82 = true
			break
		}
	}
	if !foundV82 {
		t.Error("expected to find v8.2 fallback path")
	}
}

// TestEmptyTemplates tests behavior with no templates
func TestEmptyTemplates(t *testing.T) {
	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates([]string{}, "")

	if len(validPaths) != 0 {
		t.Errorf("expected 0 paths with empty templates, got %d", len(validPaths))
	}
}

// TestNonExistentPaths tests that non-existent paths are filtered out
func TestNonExistentPaths(t *testing.T) {
	templates := []string{
		"/nonexistent/path/{{.ArchVersion}}",
		"/another/fake/path/{{.Version}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates(templates, "")

	if len(validPaths) != 0 {
		t.Errorf("expected 0 paths for non-existent directories, got %d: %v", len(validPaths), validPaths)
	}
}

// TestPathOrdering tests that paths are returned in correct priority order
// TestCPUAliasPriority tests that paths with {{.CPUAlias}} are prioritized when hint matches
func TestCPUAliasPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure:
	// /opt/zen3/lib (CPU alias path)
	// /opt/x86-64-v3/lib (version path)
	// /opt/x86-64-v2/lib (fallback version path)
	zen3Path := filepath.Join(tmpDir, "opt/zen3/lib")
	v3Path := filepath.Join(tmpDir, "opt/x86-64-v3/lib")
	v2Path := filepath.Join(tmpDir, "opt/x86-64-v2/lib")

	if err := os.MkdirAll(zen3Path, 0755); err != nil {
		t.Fatalf("failed to create zen3 directory: %v", err)
	}
	if err := os.MkdirAll(v3Path, 0755); err != nil {
		t.Fatalf("failed to create v3 directory: %v", err)
	}
	if err := os.MkdirAll(v2Path, 0755); err != nil {
		t.Fatalf("failed to create v2 directory: %v", err)
	}

	templates := []string{
		tmpDir + "/opt/{{.CPUAlias}}/lib",
		tmpDir + "/opt/{{.ArchVersion}}/lib",
	}

	evaluator := &TemplateEvaluator{
		arch:     format.ArchX86_64,
		version:  format.X86_64_V3,
		cpuAlias: "zen3", // Detected CPU alias
	}

	// Test 1: With matching hint, CPU alias path should be first
	validPaths := evaluator.EvaluateTemplates(templates, "zen3")

	if len(validPaths) != 3 {
		t.Errorf("expected 3 valid paths, got %d: %v", len(validPaths), validPaths)
	}

	// First path should be zen3 (CPU alias match with priority boost)
	if len(validPaths) > 0 && validPaths[0] != zen3Path {
		t.Errorf("first path = %s, want %s (CPU alias should be prioritized)", validPaths[0], zen3Path)
	}

	// Second path should be v3 (version)
	if len(validPaths) > 1 && validPaths[1] != v3Path {
		t.Errorf("second path = %s, want %s", validPaths[1], v3Path)
	}

	// Test 2: Without matching hint, no priority boost
	// Template order still matters within each version level
	validPaths2 := evaluator.EvaluateTemplates(templates, "")

	if len(validPaths2) != 3 {
		t.Errorf("expected 3 valid paths, got %d: %v", len(validPaths2), validPaths2)
	}

	// Without priority boost, all 3 paths are still found
	// The key difference is that matching hint evaluates templates BEFORE version fallback
	// Here we verify the paths are all present (order may vary by template order)
	pathSet := make(map[string]bool)
	for _, p := range validPaths2 {
		pathSet[p] = true
	}
	if !pathSet[zen3Path] || !pathSet[v3Path] || !pathSet[v2Path] {
		t.Errorf("missing expected paths. Got: %v", validPaths2)
	}

	// Test 3: Wrong hint, same as no hint (no priority boost)
	validPaths3 := evaluator.EvaluateTemplates(templates, "skylake")

	if len(validPaths3) != 3 {
		t.Errorf("expected 3 valid paths with wrong hint, got %d: %v", len(validPaths3), validPaths3)
	}
}

func TestPathOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple directories (using real x86-64 glibc-hwcaps naming)
	paths := []string{
		filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/x86-64-v3"), // Debian multiarch, exact
		filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v3"),                // RedHat, exact
		filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/x86-64-v2"), // Debian multiarch, fallback
		filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v2"),                // RedHat, fallback
	}

	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
	}

	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.ArchVersion}}",
		tmpDir + "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates(templates, "")

	if len(validPaths) != 4 {
		t.Fatalf("expected 4 paths, got %d", len(validPaths))
	}

	// Paths are prioritized by version first, then template order
	// Version v3: Template 0 (Debian), then Template 1 (RedHat)
	// Version v2: Template 0 (Debian), then Template 1 (RedHat)
	expected := []string{
		paths[0], // v3, Template 0 (Debian)
		paths[1], // v3, Template 1 (RedHat)
		paths[2], // v2, Template 0 (Debian)
		paths[3], // v2, Template 1 (RedHat)
	}

	for i, want := range expected {
		if validPaths[i] != want {
			t.Errorf("path[%d] = %s, want %s", i, validPaths[i], want)
		}
	}
}
