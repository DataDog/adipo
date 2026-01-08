package macho

import (
	"debug/macho"
	"encoding/binary"
	"fmt"

	"github.com/corentin-chary/adipo/internal/format"
)

// BinaryInfo contains information about a Mach-O binary
type BinaryInfo struct {
	Path           string
	Architecture   format.Architecture
	ArchVersion    format.ArchVersion
	Cpu            macho.Cpu
	SubCpu         uint32
	ByteOrder      binary.ByteOrder
	IsExecutable   bool
	EntryPoint     uint64
}

// Analyze analyzes a Mach-O binary and extracts architecture information
func Analyze(path string) (*BinaryInfo, error) {
	f, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer f.Close()

	info := &BinaryInfo{
		Path:         path,
		Cpu:          f.Cpu,
		SubCpu:       f.SubCpu,
		ByteOrder:    f.ByteOrder,
		IsExecutable: f.Type == macho.TypeExec,
	}

	// Determine architecture
	switch f.Cpu {
	case macho.CpuAmd64:
		info.Architecture = format.ArchX86_64
		// Default to v1 unless we can detect otherwise
		// Note: Architecture version detection from binary is complex
		// For now, we'll default to v1 and let the user specify explicitly
		info.ArchVersion = format.X86_64_V1
	case macho.CpuArm64:
		info.Architecture = format.ArchARM64
		// Default to v8.0
		info.ArchVersion = format.ARM64_V8_0
	default:
		return nil, fmt.Errorf("unsupported architecture: %v", f.Cpu)
	}

	// Validate that it's executable
	if !info.IsExecutable {
		return nil, fmt.Errorf("file is not an executable (type: %v)", f.Type)
	}

	return info, nil
}

// String returns a string representation of the binary info
func (b *BinaryInfo) String() string {
	return fmt.Sprintf("%s (%s, %s)",
		b.Architecture.String(),
		b.Cpu.String(),
		byteOrderString(b.ByteOrder),
	)
}

func byteOrderString(order binary.ByteOrder) string {
	if order == binary.LittleEndian {
		return "little-endian"
	}
	return "big-endian"
}
