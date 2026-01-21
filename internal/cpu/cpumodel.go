// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package cpu

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// CPUModel represents detected CPU model information
type CPUModel struct {
	// x86-64 fields (Linux)
	Vendor    string // "GenuineIntel", "AuthenticAMD"
	ModelName string // "Intel(R) Xeon(R) Platinum 8259CL CPU @ 2.50GHz"
	Family    int    // CPU family
	Model     int    // CPU model number

	// ARM64 fields (Linux)
	Implementer int // ARM implementer ID (0x41 = ARM, 0x43 = Cavium)
	PartNum     int // ARM part number (0xd0c = Neoverse N1)

	// macOS fields (both x86-64 and ARM64)
	BrandString string // "Apple M3 Max" or Intel brand string from sysctl
}

// DetectCPUModel detects CPU model information based on the operating system
// Returns nil on error (non-fatal, graceful degradation)
func DetectCPUModel() (*CPUModel, error) {
	switch runtime.GOOS {
	case "linux":
		return detectCPUModelLinux()
	case "darwin":
		return detectCPUModelMacOS()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// detectCPUModelLinux detects CPU model from /proc/cpuinfo on Linux
func detectCPUModelLinux() (*CPUModel, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/cpuinfo: %w", err)
	}
	defer func() { _ = file.Close() }()

	model := &CPUModel{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Split on colon
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse fields based on architecture
		switch runtime.GOARCH {
		case "amd64", "386":
			switch key {
			case "vendor_id":
				model.Vendor = value
			case "model name":
				model.ModelName = value
			case "cpu family":
				if f, err := strconv.Atoi(value); err == nil {
					model.Family = f
				}
			case "model":
				if m, err := strconv.Atoi(value); err == nil {
					model.Model = m
				}
			}

		case "arm64", "arm":
			switch key {
			case "CPU implementer":
				// Parse hex value (e.g., "0x41")
				if imp, err := strconv.ParseInt(value, 0, 64); err == nil {
					model.Implementer = int(imp)
				}
			case "CPU part":
				// Parse hex value (e.g., "0xd0c")
				if part, err := strconv.ParseInt(value, 0, 64); err == nil {
					model.PartNum = int(part)
				}
			}
		}

		// Stop after first processor (all cores have same info)
		if model.hasMinimalInfo() {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/cpuinfo: %w", err)
	}

	if !model.hasMinimalInfo() {
		return nil, fmt.Errorf("failed to parse CPU info from /proc/cpuinfo")
	}

	return model, nil
}

// detectCPUModelMacOS detects CPU model using sysctl on macOS
func detectCPUModelMacOS() (*CPUModel, error) {
	model := &CPUModel{}

	// Get brand string
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run sysctl for brand string: %w", err)
	}

	model.BrandString = strings.TrimSpace(string(output))

	if model.BrandString == "" {
		return nil, fmt.Errorf("empty brand string from sysctl")
	}

	return model, nil
}

// hasMinimalInfo checks if the model has enough information to be useful
func (m *CPUModel) hasMinimalInfo() bool {
	switch runtime.GOARCH {
	case "amd64", "386":
		// x86-64: Need at least vendor and family/model
		return m.Vendor != "" && (m.Family != 0 || m.Model != 0)
	case "arm64", "arm":
		switch runtime.GOOS {
		case "linux":
			// ARM64 Linux: Need implementer and part number
			return m.Implementer != 0 && m.PartNum != 0
		case "darwin":
			// ARM64 macOS: Need brand string
			return m.BrandString != ""
		}
	}
	return false
}

// String returns a human-readable representation of the CPU model
func (m *CPUModel) String() string {
	switch runtime.GOARCH {
	case "amd64", "386":
		if m.ModelName != "" {
			return m.ModelName
		}
		if m.BrandString != "" {
			return m.BrandString
		}
		return fmt.Sprintf("%s Family %d Model %d", m.Vendor, m.Family, m.Model)
	case "arm64", "arm":
		if m.BrandString != "" {
			return m.BrandString
		}
		return fmt.Sprintf("Implementer 0x%x Part 0x%x", m.Implementer, m.PartNum)
	}
	return "Unknown CPU"
}
