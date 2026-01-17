// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"strings"

	"github.com/DataDog/adipo/internal/features"
)

// armInstructionMappings maps ARM64 instruction mnemonics to required CPU features
var armInstructionMappings = []InstructionMapping{
	// NEON/ASIMD baseline instructions (v registers for vector operations)
	// These are universal in ARMv8.0+, mark as ASIMD/NEON
	{Prefix: "fadd", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fsub", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fmul", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fdiv", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fmax", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fmin", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fmla", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fmls", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fneg", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fabs", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "fsqrt", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "frecpe", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "frsqrte", RequireReg: "v", Features: features.ARM_ASIMD},

	// NEON integer vector operations
	{Prefix: "add", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "sub", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "mul", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "mla", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "mls", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "smax", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "smin", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "umax", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "umin", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "addp", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "smull", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "umull", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "saddl", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "uaddl", RequireReg: "v", Features: features.ARM_ASIMD},

	// NEON load/store (v register forms, not SVE z registers)
	{Prefix: "ld1", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "ld2", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "ld3", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "ld4", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "st1", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "st2", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "st3", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "st4", RequireReg: "v", Features: features.ARM_ASIMD},

	// NEON vector manipulation
	{Prefix: "dup", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "mov", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "ext", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "zip", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "uzp", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "trn", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "ins", RequireReg: "v", Features: features.ARM_ASIMD},
	{Prefix: "tbl", Features: features.ARM_ASIMD},
	{Prefix: "tbx", Features: features.ARM_ASIMD},
	{Prefix: "rev", RequireReg: "v", Features: features.ARM_ASIMD},

	// FP16 (Half-precision floating point) instructions
	{Prefix: "fcvt", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fadd", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fsub", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fmul", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fdiv", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fmax", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fmin", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fmaxnm", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fminnm", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fabs", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fneg", RequireReg: ".h", Features: features.ARM_FP16},
	{Prefix: "fsqrt", RequireReg: ".h", Features: features.ARM_FP16},

	// FCMA (Complex number arithmetic)
	{Prefix: "fcmla", Features: features.ARM_FCMA},
	{Prefix: "fcadd", Features: features.ARM_FCMA},

	// JSCVT (JavaScript conversion)
	{Prefix: "fjcvtzs", Features: features.ARM_JSCVT},

	// LRCPC (Load-acquire/store-release)
	{Prefix: "ldapr", Features: features.ARM_LRCPC},
	{Prefix: "ldaprb", Features: features.ARM_LRCPC},
	{Prefix: "ldaprh", Features: features.ARM_LRCPC},

	// FlagM (Flag manipulation)
	{Prefix: "cfinv", Features: features.ARM_FLAGM},
	{Prefix: "rmif", Features: features.ARM_FLAGM},
	{Prefix: "setf8", Features: features.ARM_FLAGM},
	{Prefix: "setf16", Features: features.ARM_FLAGM},

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
