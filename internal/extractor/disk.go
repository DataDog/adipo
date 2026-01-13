// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package extractor

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiskExtractor extracts binaries to disk (temp files or deterministic paths)
type DiskExtractor struct {
	TempDir      string
	FileTemplate string // Optional: template for deterministic filename
	file         *os.File
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
		// Use template as the filename directly (already expanded by caller)
		tempFile = filepath.Join(tempDir, d.FileTemplate)
		extractDir = tempDir

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(tempFile), 0755); err != nil {
			return "", nil, fmt.Errorf("failed to create directory: %w", err)
		}
	} else {
		// Create random temp directory with secure permissions (legacy behavior)
		extractDir, err = os.MkdirTemp(tempDir, "adipo-*")
		if err != nil {
			return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		// Create temp file in the directory
		tempFile = filepath.Join(extractDir, name)
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
