//go:build linux
// +build linux

package extractor

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// MemoryExtractor extracts binaries to memory using memfd_create
type MemoryExtractor struct {
	fd int
}

// Extract extracts the binary to memory
func (m *MemoryExtractor) Extract(data []byte, name string) (string, func(), error) {
	// Create anonymous file in memory
	fd, err := unix.MemfdCreate(name, 0)
	if err != nil {
		return "", nil, fmt.Errorf("memfd_create failed: %w", err)
	}

	m.fd = fd

	// Write binary data
	_, err = unix.Write(fd, data)
	if err != nil {
		unix.Close(fd)
		return "", nil, fmt.Errorf("failed to write to memfd: %w", err)
	}

	// Make executable (fchmod)
	err = unix.Fchmod(fd, 0755)
	if err != nil {
		unix.Close(fd)
		return "", nil, fmt.Errorf("failed to chmod memfd: %w", err)
	}

	// Return path as /proc/self/fd/N
	path := fmt.Sprintf("/proc/self/fd/%d", fd)

	// Cleanup function
	cleanup := func() {
		unix.Close(fd)
	}

	return path, cleanup, nil
}

// Name returns the extractor name
func (m *MemoryExtractor) Name() string {
	return "memory"
}
