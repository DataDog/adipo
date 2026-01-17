// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package features

// X86-64 feature bitmasks (for RequiredFeatures uint64)
const (
	X86_SSE3         uint64 = 1 << 0
	X86_SSSE3        uint64 = 1 << 1
	X86_SSE4_1       uint64 = 1 << 2
	X86_SSE4_2       uint64 = 1 << 3
	X86_POPCNT       uint64 = 1 << 4
	X86_AVX          uint64 = 1 << 5
	X86_AVX2         uint64 = 1 << 6
	X86_FMA          uint64 = 1 << 7
	X86_BMI1         uint64 = 1 << 8
	X86_BMI2         uint64 = 1 << 9
	X86_LZCNT        uint64 = 1 << 10
	X86_MOVBE        uint64 = 1 << 11
	X86_AVX512F      uint64 = 1 << 12
	X86_AVX512DQ     uint64 = 1 << 13
	X86_AVX512CD     uint64 = 1 << 14
	X86_AVX512BW     uint64 = 1 << 15
	X86_AVX512VL     uint64 = 1 << 16
	X86_AVX512IFMA   uint64 = 1 << 17
	X86_AVX512VBMI   uint64 = 1 << 18
	X86_AVX512VBMI2  uint64 = 1 << 19
	X86_AVX512VNNI   uint64 = 1 << 20
	X86_AVX512BITALG uint64 = 1 << 21
	X86_AVX512VPOPCNTDQ uint64 = 1 << 22
	X86_F16C         uint64 = 1 << 23
	X86_OSXSAVE      uint64 = 1 << 24
	X86_CMPXCHG16B   uint64 = 1 << 25
	X86_LAHF         uint64 = 1 << 26
	X86_SHA          uint64 = 1 << 27  // SHA-NI
	X86_GFNI         uint64 = 1 << 28  // Galois Field instructions
	X86_VAES         uint64 = 1 << 29  // Vector AES
	X86_VPCLMULQDQ   uint64 = 1 << 30  // Vector carry-less multiply
	X86_AVX512BF16   uint64 = 1 << 31  // AVX-512 BFloat16
)

// ARM64 feature bitmasks (for RequiredFeatures uint64)
const (
	ARM_FP           uint64 = 1 << 0
	ARM_ASIMD        uint64 = 1 << 1
	ARM_NEON         uint64 = ARM_ASIMD  // Alias for ASIMD
	ARM_AES          uint64 = 1 << 2
	ARM_PMULL        uint64 = 1 << 3
	ARM_SHA1         uint64 = 1 << 4
	ARM_SHA2         uint64 = 1 << 5
	ARM_CRC32        uint64 = 1 << 6
	ARM_ATOMICS      uint64 = 1 << 7  // LSE
	ARM_FP16         uint64 = 1 << 8  // FPHP
	ARM_ASIMDHP      uint64 = 1 << 9  // ASIMD HP
	ARM_CPUID        uint64 = 1 << 10
	ARM_ASIMDRDM     uint64 = 1 << 11
	ARM_JSCVT        uint64 = 1 << 12
	ARM_FCMA         uint64 = 1 << 13
	ARM_LRCPC        uint64 = 1 << 14
	ARM_DCPOP        uint64 = 1 << 15
	ARM_SHA3         uint64 = 1 << 16
	ARM_SM3          uint64 = 1 << 17
	ARM_SM4          uint64 = 1 << 18
	ARM_ASIMDDP      uint64 = 1 << 19
	ARM_SHA512       uint64 = 1 << 20
	ARM_SVE          uint64 = 1 << 21
	ARM_ASIMDFHM     uint64 = 1 << 22
	ARM_DIT          uint64 = 1 << 23
	ARM_USCAT        uint64 = 1 << 24
	ARM_ILRCPC       uint64 = 1 << 25
	ARM_FLAGM        uint64 = 1 << 26
	ARM_SSBS         uint64 = 1 << 27
	ARM_SB           uint64 = 1 << 28
	ARM_PACA         uint64 = 1 << 29
	ARM_PACG         uint64 = 1 << 30
	ARM_DPB2         uint64 = 1 << 31
	ARM_SVE2         uint64 = 1 << 32
	ARM_SVEAES       uint64 = 1 << 33
	ARM_SVEPMULL     uint64 = 1 << 34
	ARM_SVEBITPERM   uint64 = 1 << 35
	ARM_SVESHA3      uint64 = 1 << 36
	ARM_SVESM4       uint64 = 1 << 37
	ARM_FLAGM2       uint64 = 1 << 38
	ARM_FRINT        uint64 = 1 << 39
	ARM_SVEI8MM      uint64 = 1 << 40
	ARM_SVEF32MM     uint64 = 1 << 41
	ARM_SVEF64MM     uint64 = 1 << 42
	ARM_SVEBF16      uint64 = 1 << 43
	ARM_I8MM         uint64 = 1 << 44
	ARM_BF16         uint64 = 1 << 45
	ARM_DGH          uint64 = 1 << 46
	ARM_RNG          uint64 = 1 << 47
	ARM_BTI          uint64 = 1 << 48
	ARM_MTE          uint64 = 1 << 49
)

// x86-64 microarchitecture level feature requirements
var (
	// X86_64_V1 is the baseline x86-64 (x87, SSE, SSE2)
	X86_64_V1_Features uint64 = 0

	// X86_64_V2 adds CMPXCHG16B, LAHF-SAHF, POPCNT, SSE3, SSE4.1, SSE4.2, SSSE3
	X86_64_V2_Features uint64 = X86_CMPXCHG16B | X86_LAHF | X86_POPCNT | X86_SSE3 | X86_SSSE3 | X86_SSE4_1 | X86_SSE4_2

	// X86_64_V3 adds AVX, AVX2, BMI1, BMI2, F16C, FMA, LZCNT, MOVBE, OSXSAVE
	X86_64_V3_Features uint64 = X86_64_V2_Features | X86_AVX | X86_AVX2 | X86_BMI1 | X86_BMI2 | X86_F16C | X86_FMA | X86_LZCNT | X86_MOVBE | X86_OSXSAVE

	// X86_64_V4 adds AVX512F, AVX512BW, AVX512CD, AVX512DQ, AVX512VL
	X86_64_V4_Features uint64 = X86_64_V3_Features | X86_AVX512F | X86_AVX512BW | X86_AVX512CD | X86_AVX512DQ | X86_AVX512VL
)
