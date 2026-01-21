// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"debug/elf"
	"debug/macho"
	"fmt"
)

// DetectFormat detects the binary format of a file (ELF, Mach-O, PE, etc.)
func DetectFormat(path string) (BinaryFormat, error) {
	// Try ELF first
	if _, err := elf.Open(path); err == nil {
		return FormatELF, nil
	}

	// Try Mach-O
	if _, err := macho.Open(path); err == nil {
		return FormatMachO, nil
	}

	// TODO: Add PE detection when Windows support is added

	return FormatUnknown, fmt.Errorf("unknown binary format")
}
