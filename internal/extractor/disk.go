package extractor

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiskExtractor extracts binaries to disk (temp files)
type DiskExtractor struct {
	TempDir string
	file    *os.File
}

// Extract extracts the binary to a temporary file
func (d *DiskExtractor) Extract(data []byte, name string) (string, func(), error) {
	// Determine temp directory
	tempDir := d.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Create temp directory with secure permissions
	extractDir, err := os.MkdirTemp(tempDir, "adipo-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temp file in the directory
	tempFile := filepath.Join(extractDir, name)
	file, err := os.OpenFile(tempFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		os.RemoveAll(extractDir)
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	d.file = file

	// Write binary data
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		os.RemoveAll(extractDir)
		return "", nil, fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Close file (we'll reopen for execution)
	if err := file.Close(); err != nil {
		os.RemoveAll(extractDir)
		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Ensure executable permissions
	if err := os.Chmod(tempFile, 0755); err != nil {
		os.RemoveAll(extractDir)
		return "", nil, fmt.Errorf("failed to chmod temp file: %w", err)
	}

	// Cleanup function
	cleanup := func() {
		// Remove the entire extract directory
		os.RemoveAll(extractDir)
	}

	return tempFile, cleanup, nil
}

// Name returns the extractor name
func (d *DiskExtractor) Name() string {
	return "disk"
}
