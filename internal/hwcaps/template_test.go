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
			name:          "ARM64 v9.4 fallback to v8.0",
			arch:          format.ArchARM64,
			version:       format.ARM64_V9_4,
			expectedFirst: format.ARM64_V9_4,
			expectedLast:  format.ARM64_V8_0,
			expectedContains: []format.ArchVersion{
				format.ARM64_V9_4,
				format.ARM64_V9_0, // .0 variant
				format.ARM64_V8_9,
				format.ARM64_V8_0,
			},
		},
		{
			name:          "ARM64 v9.0 fallback",
			arch:          format.ArchARM64,
			version:       format.ARM64_V9_0,
			expectedFirst: format.ARM64_V9_0,
			expectedLast:  format.ARM64_V8_0,
			expectedContains: []format.ArchVersion{
				format.ARM64_V9_0,
				format.ARM64_V8_9,
				format.ARM64_V8_0,
			},
		},
		{
			name:          "ARM64 v8.5 fallback",
			arch:          format.ArchARM64,
			version:       format.ARM64_V8_5,
			expectedFirst: format.ARM64_V8_5,
			expectedLast:  format.ARM64_V8_0,
			expectedContains: []format.ArchVersion{
				format.ARM64_V8_5,
				format.ARM64_V8_0, // .0 variant
				format.ARM64_V8_4,
				format.ARM64_V8_3,
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

// TestScoring tests the path scoring mechanism
func TestScoring(t *testing.T) {
	tests := []struct {
		name         string
		arch         format.Architecture
		version      format.ArchVersion
		templateIdx  int
		versionIdx   int
		template     string
		testVersion  format.ArchVersion
		expectHigher int // Expected score should be > this value
	}{
		{
			name:         "exact version match bonus",
			arch:         format.ArchX86_64,
			version:      format.X86_64_V3,
			templateIdx:  0,
			versionIdx:   0,
			template:     "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
			testVersion:  format.X86_64_V3, // Exact match
			expectHigher: 1000,             // Should get exact match bonus
		},
		{
			name:         "Debian multiarch bonus",
			arch:         format.ArchX86_64,
			version:      format.X86_64_V3,
			templateIdx:  0,
			versionIdx:   0,
			template:     "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
			testVersion:  format.X86_64_V3,
			expectHigher: 1200, // Template + version + multiarch bonus
		},
		{
			name:         "lib64 bonus",
			arch:         format.ArchX86_64,
			version:      format.X86_64_V3,
			templateIdx:  1,
			versionIdx:   0,
			template:     "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
			testVersion:  format.X86_64_V3,
			expectHigher: 1000, // Template + version + lib64 bonus
		},
		{
			name:         "template priority matters",
			arch:         format.ArchARM64,
			version:      format.ARM64_V8_2,
			templateIdx:  0, // First template
			versionIdx:   0,
			template:     "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
			testVersion:  format.ARM64_V8_2,
			expectHigher: 1200,
		},
		{
			name:         "version fallback penalty",
			arch:         format.ArchX86_64,
			version:      format.X86_64_V3,
			templateIdx:  0,
			versionIdx:   1, // Second version in fallback chain
			template:     "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
			testVersion:  format.X86_64_V2, // Fallback version
			expectHigher: 900,              // Lower than exact match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := &TemplateEvaluator{
				arch:    tt.arch,
				version: tt.version,
			}

			score := evaluator.calculateScore(tt.templateIdx, tt.versionIdx, tt.template, tt.testVersion)

			if score <= tt.expectHigher {
				t.Errorf("calculateScore() = %d, want > %d", score, tt.expectHigher)
			}
		})
	}
}

// TestTemplateEvaluationWithDirectories tests template evaluation with real directories
func TestTemplateEvaluationWithDirectories(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create mock directory structures for x86-64
	x86v3Path := filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3")
	x86v2Path := filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v2")

	if err := os.MkdirAll(x86v3Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(x86v2Path, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Test templates with tmpDir prefix
	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
		tmpDir + "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates(templates)

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

	validPaths := evaluator.EvaluateTemplates(templates)

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

	validPaths := evaluator.EvaluateTemplates([]string{})

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

	validPaths := evaluator.EvaluateTemplates(templates)

	if len(validPaths) != 0 {
		t.Errorf("expected 0 paths for non-existent directories, got %d: %v", len(validPaths), validPaths)
	}
}

// TestPathOrdering tests that paths are returned in correct priority order
func TestPathOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple directories
	paths := []string{
		filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/v3"), // Debian multiarch, exact
		filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v3"),         // RedHat, exact
		filepath.Join(tmpDir, "usr/lib/x86_64-linux-gnu/glibc-hwcaps/v2"), // Debian multiarch, fallback
		filepath.Join(tmpDir, "usr/lib64/glibc-hwcaps/x86-64-v2"),         // RedHat, fallback
	}

	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
	}

	templates := []string{
		tmpDir + "/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
		tmpDir + "/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
	}

	evaluator := &TemplateEvaluator{
		arch:    format.ArchX86_64,
		version: format.X86_64_V3,
	}

	validPaths := evaluator.EvaluateTemplates(templates)

	if len(validPaths) != 4 {
		t.Fatalf("expected 4 paths, got %d", len(validPaths))
	}

	// First should be Debian multiarch v3 (template 0, version 0, multiarch bonus, exact match)
	if validPaths[0] != paths[0] {
		t.Errorf("path[0] = %s, want %s", validPaths[0], paths[0])
	}

	// Second should be RedHat v3 (template 1, version 0, lib64 bonus, exact match)
	if validPaths[1] != paths[1] {
		t.Errorf("path[1] = %s, want %s", validPaths[1], paths[1])
	}
}
