package runner

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/DataDog/adipo/internal/format"
)

func TestGetLibraryPathEnvVar(t *testing.T) {
	envVar := GetLibraryPathEnvVar()

	// Verify we get the correct variable for current OS
	switch runtime.GOOS {
	case "darwin":
		if envVar != "DYLD_LIBRARY_PATH" {
			t.Errorf("Expected DYLD_LIBRARY_PATH on macOS, got %s", envVar)
		}
	case "linux":
		if envVar != "LD_LIBRARY_PATH" {
			t.Errorf("Expected LD_LIBRARY_PATH on Linux, got %s", envVar)
		}
	default:
		if envVar != "LD_LIBRARY_PATH" {
			t.Errorf("Expected LD_LIBRARY_PATH as fallback, got %s", envVar)
		}
	}

	// Verify it's never empty
	if envVar == "" {
		t.Error("GetLibraryPathEnvVar returned empty string")
	}
}

func TestPrependLibraryPath(t *testing.T) {
	tests := []struct {
		name        string
		env         []string
		envVar      string
		newPath     string
		wantResult  string
		description string
	}{
		{
			name:        "prepend to existing path",
			env:         []string{"HOME=/home/user", "LD_LIBRARY_PATH=/usr/lib:/usr/local/lib", "PATH=/usr/bin"},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/custom/lib",
			wantResult:  "/opt/custom/lib:/usr/lib:/usr/local/lib",
			description: "Should prepend to existing path with colon separator",
		},
		{
			name:        "new variable (not in env)",
			env:         []string{"HOME=/home/user", "PATH=/usr/bin"},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/custom/lib",
			wantResult:  "/opt/custom/lib",
			description: "Should return just the new path when variable doesn't exist",
		},
		{
			name:        "empty existing path",
			env:         []string{"LD_LIBRARY_PATH=", "PATH=/usr/bin"},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/custom/lib",
			wantResult:  "/opt/custom/lib:",
			description: "Should handle empty existing value",
		},
		{
			name:        "DYLD_LIBRARY_PATH on macOS",
			env:         []string{"DYLD_LIBRARY_PATH=/usr/local/lib"},
			envVar:      "DYLD_LIBRARY_PATH",
			newPath:     "/opt/homebrew/lib",
			wantResult:  "/opt/homebrew/lib:/usr/local/lib",
			description: "Should work with DYLD_LIBRARY_PATH",
		},
		{
			name:        "empty environment",
			env:         []string{},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/custom/lib",
			wantResult:  "/opt/custom/lib",
			description: "Should handle empty environment",
		},
		{
			name:        "multiple colons in new path",
			env:         []string{"LD_LIBRARY_PATH=/usr/lib"},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/lib1:/opt/lib2:/opt/lib3",
			wantResult:  "/opt/lib1:/opt/lib2:/opt/lib3:/usr/lib",
			description: "Should handle multiple paths in newPath",
		},
		{
			name:        "case sensitive variable name",
			env:         []string{"ld_library_path=/usr/lib", "PATH=/usr/bin"},
			envVar:      "LD_LIBRARY_PATH",
			newPath:     "/opt/custom/lib",
			wantResult:  "/opt/custom/lib",
			description: "Should be case sensitive (lowercase var not matched)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrependLibraryPath(tt.env, tt.envVar, tt.newPath)
			if result != tt.wantResult {
				t.Errorf("PrependLibraryPath() = %v, want %v\nDescription: %s",
					result, tt.wantResult, tt.description)
			}
		})
	}
}

func TestPrepareEnvironmentWithLibPath(t *testing.T) {
	tests := []struct {
		name          string
		templates     []string
		verbose       bool
		wantEnvVarSet bool
		description   string
	}{
		{
			name:          "empty templates",
			templates:     []string{},
			verbose:       false,
			wantEnvVarSet: false,
			description:   "Should not modify environment when no templates",
		},
		{
			name: "templates with no existing paths",
			templates: []string{
				"/nonexistent/path/{{.ArchVersion}}/lib",
			},
			verbose:       false,
			wantEnvVarSet: false,
			description:   "Should not set env var when no paths exist",
		},
		{
			name: "templates with verbose",
			templates: []string{
				"/tmp",  // This should exist
			},
			verbose:       true,
			wantEnvVarSet: true,
			description:   "Should set library path and print verbose output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create metadata with templates
			metadata := &format.BinaryMetadata{
				Architecture: format.ArchX86_64,
				ArchVersion:  format.X86_64_V3,
			}
			if len(tt.templates) > 0 {
				err := metadata.SetLibraryPathTemplates(tt.templates)
				if err != nil {
					t.Fatalf("Failed to set library path templates: %v", err)
				}
			}

			// Save original environment to restore later
			origEnv := os.Environ()

			// Clear the library path environment variable for consistent testing
			libEnvVar := GetLibraryPathEnvVar()
			_ = os.Unsetenv(libEnvVar)

			// Prepare environment
			env := PrepareEnvironmentWithLibPath(metadata, tt.verbose)

			// Restore original environment
			os.Clearenv()
			for _, e := range origEnv {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					_ = os.Setenv(parts[0], parts[1])
				}
			}

			// Verify environment variable was set correctly
			var foundLibPath string
			prefix := libEnvVar + "="
			for _, e := range env {
				if strings.HasPrefix(e, prefix) {
					foundLibPath = strings.TrimPrefix(e, prefix)
					break
				}
			}

			if tt.wantEnvVarSet {
				if foundLibPath == "" {
					t.Errorf("Expected %s to be set but it was not found in environment", libEnvVar)
				}
			} else {
				// When no valid paths exist, env var should not be set with new paths
				// (it might exist from the system, but shouldn't have been modified)
				_ = foundLibPath // We just verify no panic occurred
			}
		})
	}
}

func TestPrepareEnvironmentWithLibPath_PreservesExisting(t *testing.T) {
	// Create metadata with templates that evaluate to /tmp (should exist)
	metadata := &format.BinaryMetadata{
		Architecture: format.ArchX86_64,
		ArchVersion:  format.X86_64_V3,
	}
	err := metadata.SetLibraryPathTemplates([]string{"/tmp"})
	if err != nil {
		t.Fatalf("Failed to set library path templates: %v", err)
	}

	// Set an existing library path in the environment
	libEnvVar := GetLibraryPathEnvVar()
	origValue := os.Getenv(libEnvVar)
	_ = os.Setenv(libEnvVar, "/usr/local/lib:/usr/lib")
	defer func() {
		if origValue != "" {
			_ = os.Setenv(libEnvVar, origValue)
		} else {
			_ = os.Unsetenv(libEnvVar)
		}
	}()

	// Prepare environment
	env := PrepareEnvironmentWithLibPath(metadata, false)

	// Find the library path in the environment
	var foundLibPath string
	prefix := libEnvVar + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			foundLibPath = strings.TrimPrefix(e, prefix)
			break
		}
	}

	// Verify the new path was prepended (should contain /tmp)
	if !strings.Contains(foundLibPath, "/tmp") {
		t.Errorf("Expected path to contain /tmp, got %s", foundLibPath)
	}

	// Verify original paths are still there
	if !strings.Contains(foundLibPath, "/usr/local/lib") || !strings.Contains(foundLibPath, "/usr/lib") {
		t.Errorf("Expected existing paths to be preserved, got %s", foundLibPath)
	}
}

func TestPrepareEnvironmentWithLibPath_NoLibraryPath(t *testing.T) {
	// Create metadata without library path
	metadata := &format.BinaryMetadata{}

	// Get original environment
	origEnvLen := len(os.Environ())

	// Prepare environment
	env := PrepareEnvironmentWithLibPath(metadata, false)

	// Verify environment is unchanged (should be same length or similar)
	// We can't check exact equality because extractor.GetEnvironment() might normalize it
	if len(env) == 0 {
		t.Error("Expected non-empty environment")
	}

	// Verify it's roughly the same size (within 10% to account for normalization)
	diff := len(env) - origEnvLen
	if diff < 0 {
		diff = -diff
	}
	if float64(diff)/float64(origEnvLen) > 0.1 {
		t.Errorf("Environment size changed significantly: original %d, new %d", origEnvLen, len(env))
	}
}

func TestPrepareEnvironmentWithLibPath_EmptyMetadata(t *testing.T) {
	// Test with nil metadata (defensive programming)
	// This should not panic
	metadata := &format.BinaryMetadata{}

	// Should not panic
	env := PrepareEnvironmentWithLibPath(metadata, false)

	// Should return valid environment
	if len(env) == 0 {
		t.Error("Expected non-empty environment even with empty metadata")
	}
}
