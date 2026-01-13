// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package runner

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/DataDog/adipo/internal/extractor"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/hwcaps"
)

// PrepareEnvironmentWithLibPath prepares the environment with library path from binary metadata.
// Evaluates templates at runtime and configures library paths.
//
// Parameters:
//   - metadata: Binary metadata containing library path templates
//   - verbose: If true, prints library path configuration to stderr
//
// Returns the modified environment or the original environment if no library path is set.
func PrepareEnvironmentWithLibPath(metadata *format.BinaryMetadata, verbose bool) []string {
	env := extractor.GetEnvironment()

	// Get templates from metadata
	templates := metadata.GetLibraryPathTemplates()
	if len(templates) == 0 {
		return env
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Evaluating %d library path templates...\n", len(templates))
	}

	// Create evaluator for current CPU
	evaluator, err := hwcaps.NewTemplateEvaluator(
		metadata.Architecture,
		metadata.ArchVersion,
	)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: Failed to create template evaluator: %v\n", err)
		}
		return env
	}

	// Evaluate templates and get ranked paths
	validPaths := evaluator.EvaluateTemplates(templates)
	if len(validPaths) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "No valid library paths found\n")
		}
		return env
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found %d valid library paths:\n", len(validPaths))
		for i, path := range validPaths {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, path)
		}
	}

	// Join paths with colon separator
	libraryPath := strings.Join(validPaths, ":")

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
