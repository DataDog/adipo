// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package elf

import (
	"debug/elf"
	"encoding/binary"
	"fmt"

	"github.com/DataDog/adipo/internal/format"
)

// BinaryInfo contains information about an ELF binary
type BinaryInfo struct {
	Path           string
	Architecture   format.Architecture
	ArchVersion    format.ArchVersion
	Machine        elf.Machine
	Class          elf.Class
	ByteOrder      binary.ByteOrder
	IsExecutable   bool
	HasInterpreter bool
	EntryPoint     uint64
}

// Analyze analyzes an ELF binary and extracts architecture information
func Analyze(path string) (*BinaryInfo, error) {
	f, err := elf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ELF file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info := &BinaryInfo{
		Path:         path,
		Machine:      f.Machine,
		Class:        f.Class,
		ByteOrder:    f.ByteOrder,
		IsExecutable: f.Type == elf.ET_EXEC || f.Type == elf.ET_DYN,
		EntryPoint:   f.Entry,
	}

	// Determine architecture
	switch f.Machine {
	case elf.EM_X86_64:
		info.Architecture = format.ArchX86_64
		// Default to v1 unless we can detect otherwise
		// Note: Architecture version detection from binary is complex
		// For now, we'll default to v1 and let the user specify explicitly
		info.ArchVersion = format.X86_64_V1
	case elf.EM_AARCH64:
		info.Architecture = format.ArchARM64
		// Default to v8.0
		info.ArchVersion = format.ARM64_V8_0
	default:
		return nil, fmt.Errorf("unsupported architecture: %v", f.Machine)
	}

	// Check for interpreter (dynamic linking)
	for _, prog := range f.Progs {
		if prog.Type == elf.PT_INTERP {
			info.HasInterpreter = true
			break
		}
	}

	// Validate that it's executable
	if !info.IsExecutable {
		return nil, fmt.Errorf("file is not an executable (type: %v)", f.Type)
	}

	return info, nil
}

// String returns a string representation of the binary info
func (b *BinaryInfo) String() string {
	return fmt.Sprintf("%s (%s, %s, %s)",
		b.Architecture.String(),
		b.Machine.String(),
		b.Class.String(),
		byteOrderString(b.ByteOrder),
	)
}

func byteOrderString(order binary.ByteOrder) string {
	if order == binary.LittleEndian {
		return "little-endian"
	}
	return "big-endian"
}
