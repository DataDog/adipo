package analyzer

import (
	"fmt"
	"strings"

	"github.com/DataDog/adipo/internal/features"
	"github.com/DataDog/adipo/internal/format"
)

// MarchGenerator generates GCC-style -march flags from detected features
type MarchGenerator struct {
	arch format.Architecture
}

// NewMarchGenerator creates a new march string generator for the given architecture
func NewMarchGenerator(arch format.Architecture) *MarchGenerator {
	return &MarchGenerator{arch: arch}
}

// Generate creates a GCC-style march string from the detected version and features
func (g *MarchGenerator) Generate(version format.ArchVersion, detectedFeatures uint64) string {
	switch g.arch {
	case format.ArchX86_64:
		return g.generateX86March(version, detectedFeatures)
	case format.ArchARM64:
		return g.generateARMMarch(version, detectedFeatures)
	default:
		return fmt.Sprintf("-march=unknown")
	}
}

// generateX86March generates a march string for x86-64
// Format: -march=x86-64-v3 or -march=x86-64-v3+avx512f+avx512vnni
func (g *MarchGenerator) generateX86March(version format.ArchVersion, detectedFeatures uint64) string {
	// Base march string
	var base string
	switch version {
	case format.X86_64_V1:
		base = "x86-64"
	case format.X86_64_V2:
		base = "x86-64-v2"
	case format.X86_64_V3:
		base = "x86-64-v3"
	case format.X86_64_V4:
		base = "x86-64-v4"
	default:
		base = "x86-64"
	}

	// Find features beyond the base version
	var baseFeatures uint64
	switch version {
	case format.X86_64_V1:
		baseFeatures = features.X86_64_V1_Features
	case format.X86_64_V2:
		baseFeatures = features.X86_64_V2_Features
	case format.X86_64_V3:
		baseFeatures = features.X86_64_V3_Features
	case format.X86_64_V4:
		baseFeatures = features.X86_64_V4_Features
	default:
		baseFeatures = 0
	}

	extraFeatures := detectedFeatures & ^baseFeatures
	if extraFeatures == 0 {
		return fmt.Sprintf("-march=%s", base)
	}

	// Map extra features to GCC names
	var extensions []string

	// AVX-512 extensions
	if extraFeatures&features.X86_AVX512F != 0 {
		extensions = append(extensions, "avx512f")
	}
	if extraFeatures&features.X86_AVX512DQ != 0 {
		extensions = append(extensions, "avx512dq")
	}
	if extraFeatures&features.X86_AVX512CD != 0 {
		extensions = append(extensions, "avx512cd")
	}
	if extraFeatures&features.X86_AVX512BW != 0 {
		extensions = append(extensions, "avx512bw")
	}
	if extraFeatures&features.X86_AVX512VL != 0 {
		extensions = append(extensions, "avx512vl")
	}
	if extraFeatures&features.X86_AVX512IFMA != 0 {
		extensions = append(extensions, "avx512ifma")
	}
	if extraFeatures&features.X86_AVX512VBMI != 0 {
		extensions = append(extensions, "avx512vbmi")
	}
	if extraFeatures&features.X86_AVX512VBMI2 != 0 {
		extensions = append(extensions, "avx512vbmi2")
	}
	if extraFeatures&features.X86_AVX512VNNI != 0 {
		extensions = append(extensions, "avx512vnni")
	}
	if extraFeatures&features.X86_AVX512BITALG != 0 {
		extensions = append(extensions, "avx512bitalg")
	}
	if extraFeatures&features.X86_AVX512VPOPCNTDQ != 0 {
		extensions = append(extensions, "avx512vpopcntdq")
	}

	if len(extensions) == 0 {
		return fmt.Sprintf("-march=%s", base)
	}

	return fmt.Sprintf("-march=%s+%s", base, strings.Join(extensions, "+"))
}

// generateARMMarch generates a march string for ARM64
// Format: -march=armv8.2-a+crc+crypto+lse
func (g *MarchGenerator) generateARMMarch(version format.ArchVersion, detectedFeatures uint64) string {
	// Base march string
	var base string
	switch version {
	case format.ARM64_V8_0:
		base = "armv8-a"
	case format.ARM64_V8_1:
		base = "armv8.1-a"
	case format.ARM64_V8_2:
		base = "armv8.2-a"
	case format.ARM64_V8_3:
		base = "armv8.3-a"
	case format.ARM64_V8_4:
		base = "armv8.4-a"
	case format.ARM64_V8_5:
		base = "armv8.5-a"
	case format.ARM64_V8_6:
		base = "armv8.6-a"
	case format.ARM64_V8_7:
		base = "armv8.7-a"
	case format.ARM64_V8_8:
		base = "armv8.8-a"
	case format.ARM64_V8_9:
		base = "armv8.9-a"
	case format.ARM64_V9_0:
		base = "armv9-a"
	case format.ARM64_V9_1:
		base = "armv9.1-a"
	case format.ARM64_V9_2:
		base = "armv9.2-a"
	case format.ARM64_V9_3:
		base = "armv9.3-a"
	case format.ARM64_V9_4:
		base = "armv9.4-a"
	case format.ARM64_V9_5:
		base = "armv9.5-a"
	default:
		base = "armv8-a"
	}

	// Collect architecture extensions
	var extensions []string

	// CRC32
	if detectedFeatures&features.ARM_CRC32 != 0 {
		extensions = append(extensions, "crc")
	}

	// Crypto (AES+PMULL+SHA1+SHA2)
	// GCC's +crypto enables AES, PMULL, SHA1, and SHA2
	if (detectedFeatures&features.ARM_AES != 0) &&
		(detectedFeatures&features.ARM_PMULL != 0) &&
		(detectedFeatures&features.ARM_SHA1 != 0) &&
		(detectedFeatures&features.ARM_SHA2 != 0) {
		extensions = append(extensions, "crypto")
	} else {
		// Add individual crypto features
		if detectedFeatures&features.ARM_AES != 0 {
			extensions = append(extensions, "aes")
		}
		if detectedFeatures&features.ARM_PMULL != 0 {
			extensions = append(extensions, "pmull")
		}
		if detectedFeatures&features.ARM_SHA1 != 0 {
			extensions = append(extensions, "sha1")
		}
		if detectedFeatures&features.ARM_SHA2 != 0 {
			extensions = append(extensions, "sha2")
		}
	}

	// SHA-512
	if detectedFeatures&features.ARM_SHA512 != 0 {
		extensions = append(extensions, "sha512")
	}

	// SHA3
	if detectedFeatures&features.ARM_SHA3 != 0 {
		extensions = append(extensions, "sha3")
	}

	// SM3/SM4
	if detectedFeatures&features.ARM_SM3 != 0 {
		extensions = append(extensions, "sm3")
	}
	if detectedFeatures&features.ARM_SM4 != 0 {
		extensions = append(extensions, "sm4")
	}

	// LSE (Atomics)
	if detectedFeatures&features.ARM_ATOMICS != 0 {
		extensions = append(extensions, "lse")
	}

	// FP16
	if detectedFeatures&features.ARM_FP16 != 0 {
		extensions = append(extensions, "fp16")
	}

	// ASIMD HP
	if detectedFeatures&features.ARM_ASIMDHP != 0 {
		extensions = append(extensions, "fp16")
	}

	// RDM
	if detectedFeatures&features.ARM_ASIMDRDM != 0 {
		extensions = append(extensions, "rdm")
	}

	// DotProd
	if detectedFeatures&features.ARM_ASIMDDP != 0 {
		extensions = append(extensions, "dotprod")
	}

	// SVE
	if detectedFeatures&features.ARM_SVE != 0 {
		extensions = append(extensions, "sve")
	}

	// SVE2
	if detectedFeatures&features.ARM_SVE2 != 0 {
		extensions = append(extensions, "sve2")
	}

	// BFloat16
	if detectedFeatures&features.ARM_BF16 != 0 {
		extensions = append(extensions, "bf16")
	}

	// Int8 matrix multiply
	if detectedFeatures&features.ARM_I8MM != 0 {
		extensions = append(extensions, "i8mm")
	}

	// Branch target identification
	if detectedFeatures&features.ARM_BTI != 0 {
		extensions = append(extensions, "bti")
	}

	// Memory tagging
	if detectedFeatures&features.ARM_MTE != 0 {
		extensions = append(extensions, "mte")
	}

	// Random number
	if detectedFeatures&features.ARM_RNG != 0 {
		extensions = append(extensions, "rng")
	}

	// Speculation barrier
	if detectedFeatures&features.ARM_SB != 0 {
		extensions = append(extensions, "sb")
	}

	if len(extensions) == 0 {
		return fmt.Sprintf("-march=%s", base)
	}

	return fmt.Sprintf("-march=%s+%s", base, strings.Join(extensions, "+"))
}
