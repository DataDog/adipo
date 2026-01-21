// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

//go:build capstone

package analyzer

import (
	"fmt"

	"github.com/DataDog/adipo/internal/features"
	"github.com/DataDog/adipo/internal/format"
	"github.com/bnagy/gapstone"
)

// CapstoneDisassembler wraps Capstone for disassembly
type CapstoneDisassembler struct {
	arch format.Architecture
}

// NewCapstoneDisassembler creates a new Capstone-based disassembler
func NewCapstoneDisassembler(arch format.Architecture) (*CapstoneDisassembler, error) {
	// Verify Capstone can be initialized for this architecture
	var csArch uint
	switch arch {
	case format.ArchX86_64:
		csArch = gapstone.CS_ARCH_X86
	case format.ArchARM64:
		csArch = gapstone.CS_ARCH_ARM64
	default:
		return nil, fmt.Errorf("unsupported architecture for Capstone: %s", arch)
	}

	// Test initialization
	engine, err := gapstone.New(int(csArch), gapstone.CS_MODE_LITTLE_ENDIAN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Capstone: %w", err)
	}
	engine.Close()

	return &CapstoneDisassembler{
		arch: arch,
	}, nil
}

// DisassembleBytes disassembles raw binary bytes and returns instructions with features
func (d *CapstoneDisassembler) DisassembleBytes(data []byte, maxInstructions int) ([]Instruction, error) {
	var csArch uint
	var csMode uint

	switch d.arch {
	case format.ArchX86_64:
		csArch = gapstone.CS_ARCH_X86
		csMode = gapstone.CS_MODE_64
	case format.ArchARM64:
		csArch = gapstone.CS_ARCH_ARM64
		csMode = gapstone.CS_MODE_ARM
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", d.arch)
	}

	engine, err := gapstone.New(int(csArch), int(csMode))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Capstone: %w", err)
	}
	defer engine.Close()

	// Enable detail mode to get instruction groups
	if err := engine.SetOption(gapstone.CS_OPT_DETAIL, gapstone.CS_OPT_ON); err != nil {
		return nil, fmt.Errorf("failed to enable detail mode: %w", err)
	}

	// Disassemble
	insns, err := engine.Disasm(data, 0x0, 0) // 0 = all instructions
	if err != nil {
		return nil, fmt.Errorf("disassembly failed: %w", err)
	}

	// Convert to our instruction format
	var instructions []Instruction
	for i, insn := range insns {
		if maxInstructions > 0 && i >= maxInstructions {
			break
		}

		instructions = append(instructions, Instruction{
			Address:  insn.Address,
			Mnemonic: insn.Mnemonic,
			Operands: insn.OpStr,
		})
	}

	return instructions, nil
}

// MapCapstoneGroupsToFeatures maps Capstone instruction groups to our feature flags
func MapCapstoneGroupsToFeatures(arch format.Architecture, groups []uint8) uint64 {
	var featureMask uint64

	for _, group := range groups {
		switch arch {
		case format.ArchX86_64:
			featureMask |= mapX86CapstoneGroup(group)
		case format.ArchARM64:
			featureMask |= mapARM64CapstoneGroup(group)
		}
	}

	return featureMask
}

// mapX86CapstoneGroup maps x86-64 Capstone groups to our feature flags
func mapX86CapstoneGroup(group uint8) uint64 {
	// Capstone group constants (from gapstone)
	const (
		X86_GRP_SSE1              = 131
		X86_GRP_SSE2              = 132
		X86_GRP_SSE3              = 133
		X86_GRP_SSSE3             = 134
		X86_GRP_SSE41             = 135
		X86_GRP_SSE42             = 136
		X86_GRP_AVX               = 137
		X86_GRP_AVX2              = 138
		X86_GRP_AVX512            = 139
		X86_GRP_FMA               = 140
		X86_GRP_BMI               = 141
		X86_GRP_BMI2              = 142
		X86_GRP_F16C              = 143
		X86_GRP_AES               = 144
		X86_GRP_SHA               = 145
		X86_GRP_ADX               = 146
		X86_GRP_LZCNT             = 147
		X86_GRP_MOVBE             = 148
		X86_GRP_POPCNT            = 149
		X86_GRP_CLFLUSHOPT        = 150
		X86_GRP_CLWB              = 151
		X86_GRP_PCLMUL            = 152
		X86_GRP_GFNI              = 153
		X86_GRP_VAES              = 154
		X86_GRP_VPCLMULQDQ        = 155
		X86_GRP_AVX512_BW         = 156
		X86_GRP_AVX512_DQ         = 157
		X86_GRP_AVX512_VL         = 158
		X86_GRP_AVX512_VBMI       = 159
		X86_GRP_AVX512_IFMA       = 160
		X86_GRP_AVX512_VPOPCNTDQ  = 161
		X86_GRP_AVX512_VBMI2      = 162
		X86_GRP_AVX512_BITALG     = 163
		X86_GRP_AVX512_VNNI       = 164
		X86_GRP_AVX512_BF16       = 165
		X86_GRP_AVX512_FP16       = 166
	)

	switch group {
	case X86_GRP_SSE3:
		return features.X86_SSE3
	case X86_GRP_SSSE3:
		return features.X86_SSSE3
	case X86_GRP_SSE41:
		return features.X86_SSE4_1
	case X86_GRP_SSE42:
		return features.X86_SSE4_2
	case X86_GRP_AVX:
		return features.X86_AVX
	case X86_GRP_AVX2:
		return features.X86_AVX2
	case X86_GRP_AVX512:
		return features.X86_AVX512F
	case X86_GRP_FMA:
		return features.X86_FMA
	case X86_GRP_BMI:
		return features.X86_BMI1
	case X86_GRP_BMI2:
		return features.X86_BMI2
	case X86_GRP_F16C:
		return features.X86_F16C
	case X86_GRP_LZCNT:
		return features.X86_LZCNT
	case X86_GRP_MOVBE:
		return features.X86_MOVBE
	case X86_GRP_POPCNT:
		return features.X86_POPCNT
	case X86_GRP_SHA:
		return features.X86_SHA
	// AVX-512 subextensions - we'll need to add these to features package
	case X86_GRP_AVX512_BW:
		return features.X86_AVX512BW
	case X86_GRP_AVX512_DQ:
		return features.X86_AVX512DQ
	case X86_GRP_AVX512_VL:
		return features.X86_AVX512VL
	case X86_GRP_AVX512_VBMI:
		return features.X86_AVX512VBMI
	case X86_GRP_AVX512_IFMA:
		return features.X86_AVX512IFMA
	case X86_GRP_AVX512_VPOPCNTDQ:
		return features.X86_AVX512VPOPCNTDQ
	case X86_GRP_AVX512_VBMI2:
		return features.X86_AVX512VBMI2
	case X86_GRP_AVX512_BITALG:
		return features.X86_AVX512BITALG
	case X86_GRP_AVX512_VNNI:
		return features.X86_AVX512VNNI
	case X86_GRP_AVX512_BF16:
		return features.X86_AVX512BF16
	case X86_GRP_GFNI:
		return features.X86_GFNI
	case X86_GRP_VAES:
		return features.X86_VAES
	case X86_GRP_VPCLMULQDQ:
		return features.X86_VPCLMULQDQ
	default:
		return 0
	}
}

// mapARM64CapstoneGroup maps ARM64 Capstone groups to our feature flags
func mapARM64CapstoneGroup(group uint8) uint64 {
	// Capstone group constants (from gapstone)
	const (
		ARM64_GRP_NEON   = 128
		ARM64_GRP_CRYPTO = 129
		ARM64_GRP_CRC    = 130
		ARM64_GRP_V8_1A  = 131
		ARM64_GRP_V8_2A  = 132
		ARM64_GRP_V8_3A  = 133
		ARM64_GRP_V8_4A  = 134
	)

	switch group {
	case ARM64_GRP_NEON:
		return features.ARM_NEON
	case ARM64_GRP_CRYPTO:
		// Crypto is a general group, includes AES, SHA, etc.
		return features.ARM_AES | features.ARM_SHA1 | features.ARM_SHA2
	case ARM64_GRP_CRC:
		return features.ARM_CRC32
	case ARM64_GRP_V8_1A:
		return features.ARM_ATOMICS // v8.1 primary feature
	default:
		return 0
	}
}

// IsCapstoneAvailable returns true if Capstone is available
func IsCapstoneAvailable() bool {
	return true // This file is only compiled with capstone build tag
}
