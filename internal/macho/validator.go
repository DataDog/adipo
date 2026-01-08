package macho

import (
	"debug/macho"
	"fmt"
	"os"
)

// Validate validates that a file is a valid Mach-O executable
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

	// Try to open as Mach-O
	f, err := macho.Open(path)
	if err != nil {
		return fmt.Errorf("not a valid Mach-O file: %w", err)
	}
	defer f.Close()

	// Check if it's executable
	if f.Type != macho.TypeExec {
		return fmt.Errorf("not an executable (type: %v)", f.Type)
	}

	// Check architecture is supported
	switch f.Cpu {
	case macho.CpuAmd64, macho.CpuArm64:
		// Supported
	default:
		return fmt.Errorf("unsupported architecture: %v", f.Cpu)
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
func ValidateArchitecture(path string, expectedCpu macho.Cpu) error {
	f, err := macho.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if f.Cpu != expectedCpu {
		return fmt.Errorf("architecture mismatch: expected %v, got %v",
			expectedCpu, f.Cpu)
	}

	return nil
}
