// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package cpu

import "github.com/DataDog/adipo/internal/features"

// Re-export feature constants from features package
const (
	X86_SSE3                = features.X86_SSE3
	X86_SSSE3               = features.X86_SSSE3
	X86_SSE4_1              = features.X86_SSE4_1
	X86_SSE4_2              = features.X86_SSE4_2
	X86_POPCNT              = features.X86_POPCNT
	X86_AVX                 = features.X86_AVX
	X86_AVX2                = features.X86_AVX2
	X86_FMA                 = features.X86_FMA
	X86_BMI1                = features.X86_BMI1
	X86_BMI2                = features.X86_BMI2
	X86_LZCNT               = features.X86_LZCNT
	X86_MOVBE               = features.X86_MOVBE
	X86_AVX512F             = features.X86_AVX512F
	X86_AVX512DQ            = features.X86_AVX512DQ
	X86_AVX512CD            = features.X86_AVX512CD
	X86_AVX512BW            = features.X86_AVX512BW
	X86_AVX512VL            = features.X86_AVX512VL
	X86_AVX512IFMA          = features.X86_AVX512IFMA
	X86_AVX512VBMI          = features.X86_AVX512VBMI
	X86_AVX512VBMI2         = features.X86_AVX512VBMI2
	X86_AVX512VNNI          = features.X86_AVX512VNNI
	X86_AVX512BITALG        = features.X86_AVX512BITALG
	X86_AVX512VPOPCNTDQ     = features.X86_AVX512VPOPCNTDQ
	X86_F16C                = features.X86_F16C
	X86_OSXSAVE             = features.X86_OSXSAVE
	X86_CMPXCHG16B          = features.X86_CMPXCHG16B
	X86_LAHF                = features.X86_LAHF

	ARM_FP         = features.ARM_FP
	ARM_ASIMD      = features.ARM_ASIMD
	ARM_AES        = features.ARM_AES
	ARM_PMULL      = features.ARM_PMULL
	ARM_SHA1       = features.ARM_SHA1
	ARM_SHA2       = features.ARM_SHA2
	ARM_CRC32      = features.ARM_CRC32
	ARM_ATOMICS    = features.ARM_ATOMICS
	ARM_FP16       = features.ARM_FP16
	ARM_ASIMDHP    = features.ARM_ASIMDHP
	ARM_CPUID      = features.ARM_CPUID
	ARM_ASIMDRDM   = features.ARM_ASIMDRDM
	ARM_JSCVT      = features.ARM_JSCVT
	ARM_FCMA       = features.ARM_FCMA
	ARM_LRCPC      = features.ARM_LRCPC
	ARM_DCPOP      = features.ARM_DCPOP
	ARM_SHA3       = features.ARM_SHA3
	ARM_SM3        = features.ARM_SM3
	ARM_SM4        = features.ARM_SM4
	ARM_ASIMDDP    = features.ARM_ASIMDDP
	ARM_SHA512     = features.ARM_SHA512
	ARM_SVE        = features.ARM_SVE
	ARM_ASIMDFHM   = features.ARM_ASIMDFHM
	ARM_DIT        = features.ARM_DIT
	ARM_USCAT      = features.ARM_USCAT
	ARM_ILRCPC     = features.ARM_ILRCPC
	ARM_FLAGM      = features.ARM_FLAGM
	ARM_SSBS       = features.ARM_SSBS
	ARM_SB         = features.ARM_SB
	ARM_PACA       = features.ARM_PACA
	ARM_PACG       = features.ARM_PACG
	ARM_DPB2       = features.ARM_DPB2
	ARM_SVE2       = features.ARM_SVE2
	ARM_SVEAES     = features.ARM_SVEAES
	ARM_SVEPMULL   = features.ARM_SVEPMULL
	ARM_SVEBITPERM = features.ARM_SVEBITPERM
	ARM_SVESHA3    = features.ARM_SVESHA3
	ARM_SVESM4     = features.ARM_SVESM4
	ARM_FLAGM2     = features.ARM_FLAGM2
	ARM_FRINT      = features.ARM_FRINT
	ARM_SVEI8MM    = features.ARM_SVEI8MM
	ARM_SVEF32MM   = features.ARM_SVEF32MM
	ARM_SVEF64MM   = features.ARM_SVEF64MM
	ARM_SVEBF16    = features.ARM_SVEBF16
	ARM_I8MM       = features.ARM_I8MM
	ARM_BF16       = features.ARM_BF16
	ARM_DGH        = features.ARM_DGH
	ARM_RNG        = features.ARM_RNG
	ARM_BTI        = features.ARM_BTI
	ARM_MTE        = features.ARM_MTE
)

// Re-export feature level requirements
var (
	X86_64_V1_Features = features.X86_64_V1_Features
	X86_64_V2_Features = features.X86_64_V2_Features
	X86_64_V3_Features = features.X86_64_V3_Features
	X86_64_V4_Features = features.X86_64_V4_Features
)

// Re-export feature name functions
var (
	ParseX86FeatureName = features.ParseX86FeatureName
	ParseARMFeatureName = features.ParseARMFeatureName
	FormatX86Features   = features.FormatX86Features
	FormatARMFeatures   = features.FormatARMFeatures
)
