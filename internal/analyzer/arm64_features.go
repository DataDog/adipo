package analyzer

import (
	"strings"

	"github.com/DataDog/adipo/internal/features"
)

// armInstructionMappings maps ARM64 instruction mnemonics to required CPU features
var armInstructionMappings = []InstructionMapping{
	// SVE2 instructions
	{Prefix: "sqrdmlah", Features: features.ARM_SVE2},
	{Prefix: "sqrdmlsh", Features: features.ARM_SVE2},

	// SVE instructions (check for z registers or SVE-specific mnemonics)
	{Prefix: "ptrue", Features: features.ARM_SVE},
	{Prefix: "pfalse", Features: features.ARM_SVE},
	{Prefix: "whilelt", Features: features.ARM_SVE},
	{Prefix: "rdvl", Features: features.ARM_SVE},
	{Prefix: "addvl", Features: features.ARM_SVE},
	{Prefix: "mov", RequireReg: "z", Features: features.ARM_SVE},
	{Prefix: "dup", RequireReg: "z", Features: features.ARM_SVE},
	{Prefix: "ld1", RequireReg: "z", Features: features.ARM_SVE},
	{Prefix: "st1", RequireReg: "z", Features: features.ARM_SVE},
	{Prefix: "fadd", RequireReg: "z", Features: features.ARM_SVE},
	{Prefix: "add", RequireReg: "z", Features: features.ARM_SVE},

	// Crypto extensions - AES
	{Prefix: "aese", Features: features.ARM_AES},
	{Prefix: "aesd", Features: features.ARM_AES},
	{Prefix: "aesmc", Features: features.ARM_AES},
	{Prefix: "aesimc", Features: features.ARM_AES},

	// Crypto extensions - PMULL
	{Prefix: "pmull", Features: features.ARM_PMULL},
	{Prefix: "pmull2", Features: features.ARM_PMULL},

	// Crypto extensions - SHA1
	{Prefix: "sha1c", Features: features.ARM_SHA1},
	{Prefix: "sha1p", Features: features.ARM_SHA1},
	{Prefix: "sha1m", Features: features.ARM_SHA1},
	{Prefix: "sha1h", Features: features.ARM_SHA1},

	// Crypto extensions - SHA256 (SHA2)
	{Prefix: "sha256h", Features: features.ARM_SHA2},
	{Prefix: "sha256h2", Features: features.ARM_SHA2},
	{Prefix: "sha256su0", Features: features.ARM_SHA2},
	{Prefix: "sha256su1", Features: features.ARM_SHA2},

	// Crypto extensions - SHA512
	{Prefix: "sha512h", Features: features.ARM_SHA512},
	{Prefix: "sha512h2", Features: features.ARM_SHA512},

	// Crypto extensions - SHA3
	{Prefix: "bcax", Features: features.ARM_SHA3},
	{Prefix: "eor3", Features: features.ARM_SHA3},

	// CRC32
	{Prefix: "crc32b", Features: features.ARM_CRC32},
	{Prefix: "crc32h", Features: features.ARM_CRC32},
	{Prefix: "crc32w", Features: features.ARM_CRC32},
	{Prefix: "crc32x", Features: features.ARM_CRC32},
	{Prefix: "crc32cb", Features: features.ARM_CRC32},
	{Prefix: "crc32ch", Features: features.ARM_CRC32},
	{Prefix: "crc32cw", Features: features.ARM_CRC32},
	{Prefix: "crc32cx", Features: features.ARM_CRC32},

	// Atomics (LSE - Large System Extensions)
	{Prefix: "ldadd", Features: features.ARM_ATOMICS},
	{Prefix: "stadd", Features: features.ARM_ATOMICS},
	{Prefix: "ldclr", Features: features.ARM_ATOMICS},
	{Prefix: "ldeor", Features: features.ARM_ATOMICS},
	{Prefix: "ldset", Features: features.ARM_ATOMICS},
	{Prefix: "ldsmax", Features: features.ARM_ATOMICS},
	{Prefix: "ldsmin", Features: features.ARM_ATOMICS},
	{Prefix: "ldumax", Features: features.ARM_ATOMICS},
	{Prefix: "ldumin", Features: features.ARM_ATOMICS},
	{Prefix: "swp", Features: features.ARM_ATOMICS},
	{Prefix: "cas", Features: features.ARM_ATOMICS},

	// ASIMD Dot Product
	{Prefix: "sdot", Features: features.ARM_ASIMDDP},
	{Prefix: "udot", Features: features.ARM_ASIMDDP},

	// BFloat16
	{Prefix: "bfdot", Features: features.ARM_BF16},
	{Prefix: "bfcvt", Features: features.ARM_BF16},

	// Int8 matrix multiply
	{Prefix: "smmla", Features: features.ARM_I8MM},
	{Prefix: "ummla", Features: features.ARM_I8MM},

	// Branch target identification
	{Prefix: "bti", Features: features.ARM_BTI},

	// Pointer authentication
	{Prefix: "paciasp", Features: features.ARM_PACA},
	{Prefix: "pacibsp", Features: features.ARM_PACA},
	{Prefix: "autiasp", Features: features.ARM_PACA},
	{Prefix: "autibsp", Features: features.ARM_PACA},
}

// MapARMInstructionToFeatures maps an ARM64 instruction to required CPU features
func MapARMInstructionToFeatures(insn Instruction) uint64 {
	mnemonic := strings.ToLower(insn.Mnemonic)
	operands := strings.ToLower(insn.Operands)

	// Try to match against known instruction mappings
	for _, mapping := range armInstructionMappings {
		if !strings.HasPrefix(mnemonic, mapping.Prefix) {
			continue
		}

		// If register requirement specified, check operands
		if mapping.RequireReg != "" {
			if !strings.Contains(operands, mapping.RequireReg) {
				continue
			}
		}

		return mapping.Features
	}

	// No specific features required (baseline ARMv8.0-A)
	return 0
}
