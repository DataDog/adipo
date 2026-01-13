// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"strings"

	"github.com/DataDog/adipo/internal/features"
)

// InstructionMapping maps instruction patterns to CPU feature requirements
type InstructionMapping struct {
	Prefix     string // Instruction mnemonic prefix to match
	RequireReg string // Optional register name requirement (e.g., "zmm" for AVX-512)
	Features   uint64 // Feature bitmask required for this instruction
}

// x86InstructionMappings maps x86-64 instruction mnemonics to required CPU features
// Ordered from most specific to least specific to ensure correct matching
var x86InstructionMappings = []InstructionMapping{
	// AVX-512 instructions (check for zmm registers first)
	{Prefix: "vadd", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vsub", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vmul", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vdiv", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vmov", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpand", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpor", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpxor", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpadd", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpsub", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpmul", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpmin", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpmax", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vpcmp", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vperm", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vshuf", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vbroadcast", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vextract", RequireReg: "zmm", Features: features.X86_AVX512F},
	{Prefix: "vinsert", RequireReg: "zmm", Features: features.X86_AVX512F},

	// AVX-512 specific instructions
	{Prefix: "vpdpbusd", Features: features.X86_AVX512VNNI},
	{Prefix: "vpdpwssd", Features: features.X86_AVX512VNNI},
	{Prefix: "vpopcnt", Features: features.X86_AVX512VPOPCNTDQ},
	{Prefix: "vpconflict", Features: features.X86_AVX512CD},
	{Prefix: "vplzcnt", Features: features.X86_AVX512CD},

	// FMA instructions (vfmadd*, vfmsub*, vfnmadd*, vfnmsub*)
	{Prefix: "vfmadd", Features: features.X86_FMA},
	{Prefix: "vfmsub", Features: features.X86_FMA},
	{Prefix: "vfnmadd", Features: features.X86_FMA},
	{Prefix: "vfnmsub", Features: features.X86_FMA},

	// AVX2 instructions
	{Prefix: "vperm2i128", Features: features.X86_AVX2},
	{Prefix: "vinserti128", Features: features.X86_AVX2},
	{Prefix: "vextracti128", Features: features.X86_AVX2},
	{Prefix: "vpbroadcast", Features: features.X86_AVX2},
	{Prefix: "vpsllv", Features: features.X86_AVX2},
	{Prefix: "vpsrlv", Features: features.X86_AVX2},
	{Prefix: "vpsrav", Features: features.X86_AVX2},
	{Prefix: "vpmaskmov", Features: features.X86_AVX2},
	{Prefix: "vpgather", Features: features.X86_AVX2},

	// AVX/AVX2 instructions (could be either, mark both)
	// These use ymm registers but could be AVX or AVX2
	{Prefix: "vpadd", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpsub", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpmul", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpmin", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpmax", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpand", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpor", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpxor", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpcmp", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpshuf", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpsll", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpsrl", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpsra", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpunpck", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpack", Features: features.X86_AVX | features.X86_AVX2},
	{Prefix: "vpmovmskb", Features: features.X86_AVX | features.X86_AVX2},

	// AVX floating point instructions
	{Prefix: "vadd", Features: features.X86_AVX},
	{Prefix: "vsub", Features: features.X86_AVX},
	{Prefix: "vmul", Features: features.X86_AVX},
	{Prefix: "vdiv", Features: features.X86_AVX},
	{Prefix: "vsqrt", Features: features.X86_AVX},
	{Prefix: "vmax", Features: features.X86_AVX},
	{Prefix: "vmin", Features: features.X86_AVX},
	{Prefix: "vhad", Features: features.X86_AVX},
	{Prefix: "vhsub", Features: features.X86_AVX},
	{Prefix: "vcmp", Features: features.X86_AVX},
	{Prefix: "vrcp", Features: features.X86_AVX},
	{Prefix: "vrsqrt", Features: features.X86_AVX},
	{Prefix: "vblend", Features: features.X86_AVX},
	{Prefix: "vdpp", Features: features.X86_AVX},
	{Prefix: "vround", Features: features.X86_AVX},
	{Prefix: "vinsert", Features: features.X86_AVX},
	{Prefix: "vextract", Features: features.X86_AVX},
	{Prefix: "vbroadcast", Features: features.X86_AVX},
	{Prefix: "vmaskmov", Features: features.X86_AVX},
	{Prefix: "vperm", Features: features.X86_AVX},
	{Prefix: "vtest", Features: features.X86_AVX},
	{Prefix: "vzero", Features: features.X86_AVX},

	// AVX movs
	{Prefix: "vmovaps", Features: features.X86_AVX},
	{Prefix: "vmovapd", Features: features.X86_AVX},
	{Prefix: "vmovups", Features: features.X86_AVX},
	{Prefix: "vmovupd", Features: features.X86_AVX},
	{Prefix: "vmovss", Features: features.X86_AVX},
	{Prefix: "vmovsd", Features: features.X86_AVX},
	{Prefix: "vmovdqa", Features: features.X86_AVX},
	{Prefix: "vmovdqu", Features: features.X86_AVX},
	{Prefix: "vmovhlps", Features: features.X86_AVX},
	{Prefix: "vmovlhps", Features: features.X86_AVX},
	{Prefix: "vmovshdup", Features: features.X86_AVX},
	{Prefix: "vmovsldup", Features: features.X86_AVX},

	// F16C instructions
	{Prefix: "vcvtph2ps", Features: features.X86_F16C},
	{Prefix: "vcvtps2ph", Features: features.X86_F16C},

	// BMI1 instructions
	{Prefix: "andn", Features: features.X86_BMI1},
	{Prefix: "bextr", Features: features.X86_BMI1},
	{Prefix: "blsi", Features: features.X86_BMI1},
	{Prefix: "blsmsk", Features: features.X86_BMI1},
	{Prefix: "blsr", Features: features.X86_BMI1},
	{Prefix: "tzcnt", Features: features.X86_BMI1},

	// BMI2 instructions
	{Prefix: "bzhi", Features: features.X86_BMI2},
	{Prefix: "mulx", Features: features.X86_BMI2},
	{Prefix: "pdep", Features: features.X86_BMI2},
	{Prefix: "pext", Features: features.X86_BMI2},
	{Prefix: "rorx", Features: features.X86_BMI2},
	{Prefix: "sarx", Features: features.X86_BMI2},
	{Prefix: "shlx", Features: features.X86_BMI2},
	{Prefix: "shrx", Features: features.X86_BMI2},

	// Other x86-64-v3 instructions
	{Prefix: "lzcnt", Features: features.X86_LZCNT},
	{Prefix: "movbe", Features: features.X86_MOVBE},

	// x86-64-v2 instructions
	{Prefix: "popcnt", Features: features.X86_POPCNT},
	{Prefix: "cmpxchg16b", Features: features.X86_CMPXCHG16B},

	// SSE4.2 instructions
	{Prefix: "crc32", Features: features.X86_SSE4_2},
	{Prefix: "pcmpgtq", Features: features.X86_SSE4_2},
	{Prefix: "pcmpestri", Features: features.X86_SSE4_2},
	{Prefix: "pcmpestrm", Features: features.X86_SSE4_2},
	{Prefix: "pcmpistri", Features: features.X86_SSE4_2},
	{Prefix: "pcmpistrm", Features: features.X86_SSE4_2},

	// SSE4.1 instructions
	{Prefix: "pblendvb", Features: features.X86_SSE4_1},
	{Prefix: "pblendw", Features: features.X86_SSE4_1},
	{Prefix: "pblendvps", Features: features.X86_SSE4_1},
	{Prefix: "pblendvpd", Features: features.X86_SSE4_1},
	{Prefix: "pminsb", Features: features.X86_SSE4_1},
	{Prefix: "pminsd", Features: features.X86_SSE4_1},
	{Prefix: "pminuw", Features: features.X86_SSE4_1},
	{Prefix: "pminud", Features: features.X86_SSE4_1},
	{Prefix: "pmaxsb", Features: features.X86_SSE4_1},
	{Prefix: "pmaxsd", Features: features.X86_SSE4_1},
	{Prefix: "pmaxuw", Features: features.X86_SSE4_1},
	{Prefix: "pmaxud", Features: features.X86_SSE4_1},
	{Prefix: "pmulld", Features: features.X86_SSE4_1},
	{Prefix: "pmuldq", Features: features.X86_SSE4_1},
	{Prefix: "roundps", Features: features.X86_SSE4_1},
	{Prefix: "roundpd", Features: features.X86_SSE4_1},
	{Prefix: "roundss", Features: features.X86_SSE4_1},
	{Prefix: "roundsd", Features: features.X86_SSE4_1},
	{Prefix: "dpps", Features: features.X86_SSE4_1},
	{Prefix: "dppd", Features: features.X86_SSE4_1},
	{Prefix: "mpsadbw", Features: features.X86_SSE4_1},
	{Prefix: "ptest", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxbw", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxbd", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxbq", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxwd", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxwq", Features: features.X86_SSE4_1},
	{Prefix: "pmovsxdq", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxbw", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxbd", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxbq", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxwd", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxwq", Features: features.X86_SSE4_1},
	{Prefix: "pmovzxdq", Features: features.X86_SSE4_1},
	{Prefix: "pinsrb", Features: features.X86_SSE4_1},
	{Prefix: "pinsrd", Features: features.X86_SSE4_1},
	{Prefix: "pinsrq", Features: features.X86_SSE4_1},
	{Prefix: "pextrb", Features: features.X86_SSE4_1},
	{Prefix: "pextrd", Features: features.X86_SSE4_1},
	{Prefix: "pextrq", Features: features.X86_SSE4_1},
	{Prefix: "extractps", Features: features.X86_SSE4_1},
	{Prefix: "insertps", Features: features.X86_SSE4_1},
	{Prefix: "movntdqa", Features: features.X86_SSE4_1},
	{Prefix: "packusdw", Features: features.X86_SSE4_1},

	// SSSE3 instructions
	{Prefix: "pshufb", Features: features.X86_SSSE3},
	{Prefix: "phaddw", Features: features.X86_SSSE3},
	{Prefix: "phaddd", Features: features.X86_SSSE3},
	{Prefix: "phaddsw", Features: features.X86_SSSE3},
	{Prefix: "phsubw", Features: features.X86_SSSE3},
	{Prefix: "phsubd", Features: features.X86_SSSE3},
	{Prefix: "phsubsw", Features: features.X86_SSSE3},
	{Prefix: "pmaddubsw", Features: features.X86_SSSE3},
	{Prefix: "pmulhrsw", Features: features.X86_SSSE3},
	{Prefix: "psignb", Features: features.X86_SSSE3},
	{Prefix: "psignw", Features: features.X86_SSSE3},
	{Prefix: "psignd", Features: features.X86_SSSE3},
	{Prefix: "pabsb", Features: features.X86_SSSE3},
	{Prefix: "pabsw", Features: features.X86_SSSE3},
	{Prefix: "pabsd", Features: features.X86_SSSE3},
	{Prefix: "palignr", Features: features.X86_SSSE3},

	// SSE3 instructions
	{Prefix: "addsubps", Features: features.X86_SSE3},
	{Prefix: "addsubpd", Features: features.X86_SSE3},
	{Prefix: "haddps", Features: features.X86_SSE3},
	{Prefix: "haddpd", Features: features.X86_SSE3},
	{Prefix: "hsubps", Features: features.X86_SSE3},
	{Prefix: "hsubpd", Features: features.X86_SSE3},
	{Prefix: "movshdup", Features: features.X86_SSE3},
	{Prefix: "movsldup", Features: features.X86_SSE3},
	{Prefix: "movddup", Features: features.X86_SSE3},
	{Prefix: "lddqu", Features: features.X86_SSE3},

	// SSE/SSE2 are baseline for x86-64, so no feature flag needed
}

// MapX86InstructionToFeatures maps an x86-64 instruction to required CPU features
func MapX86InstructionToFeatures(insn Instruction) uint64 {
	mnemonic := strings.ToLower(insn.Mnemonic)
	operands := strings.ToLower(insn.Operands)

	// Try to match against known instruction mappings
	for _, mapping := range x86InstructionMappings {
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

	// No specific features required (baseline SSE/SSE2)
	return 0
}
