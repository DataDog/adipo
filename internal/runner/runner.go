package runner

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/DataDog/adipo/internal/extractor"
	"github.com/DataDog/adipo/internal/format"
)

// PrepareEnvironmentWithLibPath prepares the environment with library path from binary metadata.
// It reads the library path from the metadata and prepends it to LD_LIBRARY_PATH (Linux)
// or DYLD_LIBRARY_PATH (macOS) before execution.
//
// Parameters:
//   - metadata: Binary metadata containing the library path
//   - verbose: If true, prints library path configuration to stderr
//
// Returns the modified environment or the original environment if no library path is set.
func PrepareEnvironmentWithLibPath(metadata *format.BinaryMetadata, verbose bool) []string {
	libraryPath := metadata.GetLibraryPath()
	env := extractor.GetEnvironment()

	if libraryPath == "" {
		return env
	}

	// Determine environment variable based on OS
	libEnvVar := GetLibraryPathEnvVar()

	if verbose {
		fmt.Fprintf(os.Stderr, "Setting %s=%s\n", libEnvVar, libraryPath)
	}

	// Prepend library path to existing value (if any)
	overrides := make(map[string]string)
	overrides[libEnvVar] = PrependLibraryPath(env, libEnvVar, libraryPath)

	return extractor.SetupEnvironment(env, overrides)
}

// GetLibraryPathEnvVar returns the appropriate library path environment variable
// for the current operating system.
//
// Returns:
//   - "DYLD_LIBRARY_PATH" on macOS
//   - "LD_LIBRARY_PATH" on Linux and other platforms
func GetLibraryPathEnvVar() string {
	switch runtime.GOOS {
	case "darwin":
		return "DYLD_LIBRARY_PATH"
	case "linux":
		return "LD_LIBRARY_PATH"
	default:
		// Other platforms: use LD_LIBRARY_PATH as fallback
		return "LD_LIBRARY_PATH"
	}
}

// PrependLibraryPath prepends newPath to the existing value of envVar in the environment.
// If envVar doesn't exist in the environment, returns just newPath.
// If envVar exists, returns "newPath:existingPath".
//
// Parameters:
//   - env: The environment as a slice of "KEY=VALUE" strings
//   - envVar: The environment variable name to modify
//   - newPath: The path to prepend
//
// Returns the new value for the environment variable (without the "KEY=" prefix).
func PrependLibraryPath(env []string, envVar string, newPath string) string {
	// Find existing value in environment
	prefix := envVar + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			existingPath := strings.TrimPrefix(e, prefix)
			// Prepend new path with colon separator
			return newPath + ":" + existingPath
		}
	}

	// Environment variable doesn't exist, return just the new path
	return newPath
}
