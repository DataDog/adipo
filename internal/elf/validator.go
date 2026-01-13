// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package elf

import (
	"debug/elf"
	"fmt"
	"os"
)

// Validate validates that a file is a valid ELF executable
func Validate(path string) error {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", path)
	}

	// Try to open as ELF
	f, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("not a valid ELF file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Check if it's executable or shared object
	if f.Type != elf.ET_EXEC && f.Type != elf.ET_DYN {
		return fmt.Errorf("not an executable or shared object (type: %v)", f.Type)
	}

	// Check architecture is supported
	switch f.Machine {
	case elf.EM_X86_64, elf.EM_AARCH64:
		// Supported
	default:
		return fmt.Errorf("unsupported architecture: %v", f.Machine)
	}

	return nil
}

// ValidateExecutable validates that a file is executable (has execute permission)
func ValidateExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check for execute permission
	mode := info.Mode()
	if mode&0111 == 0 {
		return fmt.Errorf("file is not executable: %s", path)
	}

	return nil
}

// ValidateArchitecture validates that a binary matches the expected architecture
func ValidateArchitecture(path string, expectedMachine elf.Machine) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if f.Machine != expectedMachine {
		return fmt.Errorf("architecture mismatch: expected %v, got %v",
			expectedMachine, f.Machine)
	}

	return nil
}
