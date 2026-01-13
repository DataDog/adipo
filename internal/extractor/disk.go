// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiskExtractor extracts binaries to disk (temp files or deterministic paths)
type DiskExtractor struct {
	TempDir      string
	FileTemplate string // Optional: template for deterministic filename
	file         *os.File
}

// ValidatePath ensures targetPath stays within baseDir and prevents directory traversal attacks.
// It resolves both paths to absolute form and verifies the target doesn't escape the base directory.
func ValidatePath(baseDir, targetPath string) error {
	// Resolve to absolute paths
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Use filepath.Rel to check containment
	// If target is within base, rel won't start with ".."
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Check for escape attempt
	if strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, string(filepath.Separator)) {
		return fmt.Errorf("path traversal blocked: %q escapes base directory %q", targetPath, baseDir)
	}

	return nil
}

// Extract extracts the binary to a file (temp or deterministic path)
func (d *DiskExtractor) Extract(data []byte, name string) (string, func(), error) {
	// Determine temp directory
	tempDir := d.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	var extractDir string
	var tempFile string
	var err error

	// If FileTemplate is provided, use it for deterministic path
	if d.FileTemplate != "" {
		// FileTemplate must be a filename only (no directories) for security
		// This prevents TOCTOU vulnerabilities and path traversal attacks
		if strings.ContainsRune(d.FileTemplate, filepath.Separator) ||
			strings.ContainsRune(d.FileTemplate, '/') { // Check both native and forward slash
			return "", nil, fmt.Errorf("FileTemplate must be a filename only (no directories): %q", d.FileTemplate)
		}

		// Use template as the filename directly (already expanded by caller)
		tempFile = filepath.Join(tempDir, d.FileTemplate)
		extractDir = tempDir

		// Validate path to prevent directory traversal (defense in depth)
		if err := ValidatePath(tempDir, tempFile); err != nil {
			return "", nil, fmt.Errorf("invalid extraction path: %w", err)
		}

		// No directory creation needed - file goes directly in tempDir
		// This eliminates TOCTOU race conditions with symlink attacks
	} else {
		// Create random temp directory with secure permissions (legacy behavior)
		extractDir, err = os.MkdirTemp(tempDir, "adipo-*")
		if err != nil {
			return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		// Create temp file in the directory
		tempFile = filepath.Join(extractDir, name)

		// Validate path to prevent directory traversal (defense in depth)
		if err := ValidatePath(extractDir, tempFile); err != nil {
			_ = os.RemoveAll(extractDir)
			return "", nil, fmt.Errorf("invalid extraction path: %w", err)
		}
	}

	// Create/open the file
	file, err := os.OpenFile(tempFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		if d.FileTemplate == "" {
			_ = os.RemoveAll(extractDir)
		}
		return "", nil, fmt.Errorf("failed to create file: %w", err)
	}

	d.file = file

	// Write binary data
	_, err = file.Write(data)
	if err != nil {
		_ = file.Close()
		if d.FileTemplate == "" {
			_ = os.RemoveAll(extractDir)
		} else {
			_ = os.Remove(tempFile)
		}
		return "", nil, fmt.Errorf("failed to write to file: %w", err)
	}

	// Close file (we'll reopen for execution)
	if err := file.Close(); err != nil {
		if d.FileTemplate == "" {
			_ = os.RemoveAll(extractDir)
		} else {
			_ = os.Remove(tempFile)
		}
		return "", nil, fmt.Errorf("failed to close file: %w", err)
	}

	// Ensure executable permissions
	if err := os.Chmod(tempFile, 0755); err != nil {
		if d.FileTemplate == "" {
			_ = os.RemoveAll(extractDir)
		} else {
			_ = os.Remove(tempFile)
		}
		return "", nil, fmt.Errorf("failed to chmod file: %w", err)
	}

	// Cleanup function
	cleanup := func() {
		if d.FileTemplate == "" {
			// Remove the entire random temp directory
			_ = os.RemoveAll(extractDir)
		} else {
			// Remove just the file (deterministic path)
			_ = os.Remove(tempFile)
		}
	}

	return tempFile, cleanup, nil
}

// Name returns the extractor name
func (d *DiskExtractor) Name() string {
	return "disk"
}
