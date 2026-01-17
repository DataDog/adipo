// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package features

// X86FeatureNames maps feature bits to human-readable names
var X86FeatureNames = map[uint64]string{
	X86_SSE3:         "sse3",
	X86_SSSE3:        "ssse3",
	X86_SSE4_1:       "sse4.1",
	X86_SSE4_2:       "sse4.2",
	X86_POPCNT:       "popcnt",
	X86_AVX:          "avx",
	X86_AVX2:         "avx2",
	X86_FMA:          "fma",
	X86_BMI1:         "bmi1",
	X86_BMI2:         "bmi2",
	X86_LZCNT:        "lzcnt",
	X86_MOVBE:        "movbe",
	X86_AVX512F:      "avx512f",
	X86_AVX512DQ:     "avx512dq",
	X86_AVX512CD:     "avx512cd",
	X86_AVX512BW:     "avx512bw",
	X86_AVX512VL:     "avx512vl",
	X86_AVX512IFMA:   "avx512ifma",
	X86_AVX512VBMI:   "avx512vbmi",
	X86_AVX512VBMI2:  "avx512vbmi2",
	X86_AVX512VNNI:   "avx512vnni",
	X86_AVX512BITALG: "avx512bitalg",
	X86_AVX512VPOPCNTDQ: "avx512vpopcntdq",
	X86_F16C:         "f16c",
	X86_OSXSAVE:      "osxsave",
	X86_CMPXCHG16B:   "cmpxchg16b",
	X86_LAHF:         "lahf",
	X86_SHA:          "sha",
	X86_GFNI:         "gfni",
	X86_VAES:         "vaes",
	X86_VPCLMULQDQ:   "vpclmulqdq",
	X86_AVX512BF16:   "avx512bf16",
}

// ARMFeatureNames maps feature bits to human-readable names
// Note: ARM_NEON is an alias for ARM_ASIMD, so we only include ASIMD in the map
var ARMFeatureNames = map[uint64]string{
	ARM_FP:           "fp",
	ARM_ASIMD:        "asimd",  // Also known as NEON
	ARM_AES:          "aes",
	ARM_PMULL:        "pmull",
	ARM_SHA1:         "sha1",
	ARM_SHA2:         "sha2",
	ARM_CRC32:        "crc32",
	ARM_ATOMICS:      "atomics",
	ARM_FP16:         "fphp",
	ARM_ASIMDHP:      "asimdhp",
	ARM_CPUID:        "cpuid",
	ARM_ASIMDRDM:     "asimdrdm",
	ARM_JSCVT:        "jscvt",
	ARM_FCMA:         "fcma",
	ARM_LRCPC:        "lrcpc",
	ARM_DCPOP:        "dcpop",
	ARM_SHA3:         "sha3",
	ARM_SM3:          "sm3",
	ARM_SM4:          "sm4",
	ARM_ASIMDDP:      "asimddp",
	ARM_SHA512:       "sha512",
	ARM_SVE:          "sve",
	ARM_ASIMDFHM:     "asimdfhm",
	ARM_DIT:          "dit",
	ARM_USCAT:        "uscat",
	ARM_ILRCPC:       "ilrcpc",
	ARM_FLAGM:        "flagm",
	ARM_SSBS:         "ssbs",
	ARM_SB:           "sb",
	ARM_PACA:         "paca",
	ARM_PACG:         "pacg",
	ARM_DPB2:         "dpb2",
	ARM_SVE2:         "sve2",
	ARM_SVEAES:       "sveaes",
	ARM_SVEPMULL:     "svepmull",
	ARM_SVEBITPERM:   "svebitperm",
	ARM_SVESHA3:      "svesha3",
	ARM_SVESM4:       "svesm4",
	ARM_FLAGM2:       "flagm2",
	ARM_FRINT:        "frint",
	ARM_SVEI8MM:      "svei8mm",
	ARM_SVEF32MM:     "svef32mm",
	ARM_SVEF64MM:     "svef64mm",
	ARM_SVEBF16:      "svebf16",
	ARM_I8MM:         "i8mm",
	ARM_BF16:         "bf16",
	ARM_DGH:          "dgh",
	ARM_RNG:          "rng",
	ARM_BTI:          "bti",
	ARM_MTE:          "mte",
}

// ParseX86FeatureName converts a feature name string to its bitmask
func ParseX86FeatureName(name string) (uint64, bool) {
	for bit, fname := range X86FeatureNames {
		if fname == name {
			return bit, true
		}
	}
	return 0, false
}

// ParseARMFeatureName converts a feature name string to its bitmask
func ParseARMFeatureName(name string) (uint64, bool) {
	for bit, fname := range ARMFeatureNames {
		if fname == name {
			return bit, true
		}
	}
	return 0, false
}

// FormatX86Features converts a feature bitmask to a list of feature names
func FormatX86Features(features uint64) []string {
	result := make([]string, 0)
	for bit, name := range X86FeatureNames {
		if features&bit != 0 {
			result = append(result, name)
		}
	}
	return result
}

// FormatARMFeatures converts a feature bitmask to a list of feature names
func FormatARMFeatures(features uint64) []string {
	result := make([]string, 0)
	for bit, name := range ARMFeatureNames {
		if features&bit != 0 {
			result = append(result, name)
		}
	}
	return result
}
