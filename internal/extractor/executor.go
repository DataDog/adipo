// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


//go:build unix

package extractor

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Execute executes a binary at the given path with args and environment
// This function does not return if successful (it replaces the current process)
func Execute(path string, args []string, env []string) error {
	// Prepare arguments (first arg should be the program name)
	argv := make([]string, len(args)+1)
	argv[0] = path
	copy(argv[1:], args)

	// Execute the binary (replaces current process)
	err := unix.Exec(path, argv, env)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	// This line should never be reached if exec succeeds
	return fmt.Errorf("exec returned unexpectedly")
}

// ExecuteWithPath is similar to Execute but allows specifying a custom argv[0]
func ExecuteWithPath(path string, argv0 string, args []string, env []string) error {
	// Prepare arguments
	argv := make([]string, len(args)+1)
	argv[0] = argv0
	copy(argv[1:], args)

	// Execute the binary
	err := unix.Exec(path, argv, env)
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return fmt.Errorf("exec returned unexpectedly")
}

// GetEnvironment returns the current environment variables
// Can be modified by the caller before passing to Execute
func GetEnvironment() []string {
	return os.Environ()
}

// SetupEnvironment adds or overrides environment variables
func SetupEnvironment(base []string, overrides map[string]string) []string {
	env := make(map[string]string)

	// Parse base environment
	for _, entry := range base {
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				key := entry[:i]
				value := entry[i+1:]
				env[key] = value
				break
			}
		}
	}

	// Apply overrides
	for key, value := range overrides {
		env[key] = value
	}

	// Convert back to []string
	result := make([]string, 0, len(env))
	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}
