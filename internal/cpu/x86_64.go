//go:build amd64

package cpu

import (
	"github.com/DataDog/adipo/internal/format"
	"golang.org/x/sys/cpu"
)

// DetectX86_64 detects x86-64 CPU capabilities
func DetectX86_64() (*Capabilities, error) {
	caps := NewCapabilities("x86-64")
	caps.ArchType = format.ArchX86_64

	// Build feature mask and map
	var featureMask uint64

	// SSE3
	if cpu.X86.HasSSE3 {
		featureMask |= X86_SSE3
		caps.Features["sse3"] = struct{}{}
	}

	// SSSE3
	if cpu.X86.HasSSSE3 {
		featureMask |= X86_SSSE3
		caps.Features["ssse3"] = struct{}{}
	}

	// SSE4.1
	if cpu.X86.HasSSE41 {
		featureMask |= X86_SSE4_1
		caps.Features["sse4.1"] = struct{}{}
	}

	// SSE4.2
	if cpu.X86.HasSSE42 {
		featureMask |= X86_SSE4_2
		caps.Features["sse4.2"] = struct{}{}
	}

	// POPCNT
	if cpu.X86.HasPOPCNT {
		featureMask |= X86_POPCNT
		caps.Features["popcnt"] = struct{}{}
	}

	// AVX
	if cpu.X86.HasAVX {
		featureMask |= X86_AVX
		caps.Features["avx"] = struct{}{}
	}

	// AVX2
	if cpu.X86.HasAVX2 {
		featureMask |= X86_AVX2
		caps.Features["avx2"] = struct{}{}
	}

	// FMA
	if cpu.X86.HasFMA {
		featureMask |= X86_FMA
		caps.Features["fma"] = struct{}{}
	}

	// BMI1
	if cpu.X86.HasBMI1 {
		featureMask |= X86_BMI1
		caps.Features["bmi1"] = struct{}{}
	}

	// BMI2
	if cpu.X86.HasBMI2 {
		featureMask |= X86_BMI2
		caps.Features["bmi2"] = struct{}{}
	}

	// LZCNT (part of ABM)
	if cpu.X86.HasBMI1 { // LZCNT is typically part of BMI1
		featureMask |= X86_LZCNT
		caps.Features["lzcnt"] = struct{}{}
	}

	// OSXSAVE
	if cpu.X86.HasOSXSAVE {
		featureMask |= X86_OSXSAVE
		caps.Features["osxsave"] = struct{}{}
	}

	// Infer missing features that golang.org/x/sys/cpu doesn't expose
	// These are virtually universal on CPUs that have the marker features we check

	// LAHF-SAHF: Present on all x86-64 CPUs (part of V1 originally, required for V2 in spec)
	// Since we're on x86-64, we can safely assume LAHF is present
	featureMask |= X86_LAHF
	caps.Features["lahf"] = struct{}{}

	// CMPXCHG16B: Present on all x86-64 CPUs since ~2005 (required for V2)
	// If we have SSE4.2 (V2 marker), we definitely have CMPXCHG16B
	if cpu.X86.HasSSE42 {
		featureMask |= X86_CMPXCHG16B
		caps.Features["cmpxchg16b"] = struct{}{}
	}

	// F16C: Virtually always present with AVX (required for V3)
	// If we have AVX, we almost certainly have F16C
	if cpu.X86.HasAVX {
		featureMask |= X86_F16C
		caps.Features["f16c"] = struct{}{}
	}

	// MOVBE: Virtually always present with AVX2 (required for V3)
	// If we have AVX2, we almost certainly have MOVBE
	if cpu.X86.HasAVX2 {
		featureMask |= X86_MOVBE
		caps.Features["movbe"] = struct{}{}
	}

	// AVX-512 features
	if cpu.X86.HasAVX512F {
		featureMask |= X86_AVX512F
		caps.Features["avx512f"] = struct{}{}
	}

	if cpu.X86.HasAVX512DQ {
		featureMask |= X86_AVX512DQ
		caps.Features["avx512dq"] = struct{}{}
	}

	if cpu.X86.HasAVX512CD {
		featureMask |= X86_AVX512CD
		caps.Features["avx512cd"] = struct{}{}
	}

	if cpu.X86.HasAVX512BW {
		featureMask |= X86_AVX512BW
		caps.Features["avx512bw"] = struct{}{}
	}

	if cpu.X86.HasAVX512VL {
		featureMask |= X86_AVX512VL
		caps.Features["avx512vl"] = struct{}{}
	}

	if cpu.X86.HasAVX512IFMA {
		featureMask |= X86_AVX512IFMA
		caps.Features["avx512ifma"] = struct{}{}
	}

	if cpu.X86.HasAVX512VBMI {
		featureMask |= X86_AVX512VBMI
		caps.Features["avx512vbmi"] = struct{}{}
	}

	if cpu.X86.HasAVX512VBMI2 {
		featureMask |= X86_AVX512VBMI2
		caps.Features["avx512vbmi2"] = struct{}{}
	}

	if cpu.X86.HasAVX512VNNI {
		featureMask |= X86_AVX512VNNI
		caps.Features["avx512vnni"] = struct{}{}
	}

	if cpu.X86.HasAVX512BITALG {
		featureMask |= X86_AVX512BITALG
		caps.Features["avx512bitalg"] = struct{}{}
	}

	if cpu.X86.HasAVX512VPOPCNTDQ {
		featureMask |= X86_AVX512VPOPCNTDQ
		caps.Features["avx512vpopcntdq"] = struct{}{}
	}

	caps.FeatureMask = featureMask

	// Determine x86-64 microarchitecture level
	caps.Version, caps.VersionStr = detectX86Level(featureMask)

	return caps, nil
}

// detectX86Level determines the x86-64 microarchitecture level
func detectX86Level(features uint64) (format.ArchVersion, string) {
	// Check for v4 (AVX-512)
	if (features & X86_64_V4_Features) == X86_64_V4_Features {
		return format.X86_64_V4, "v4"
	}

	// Check for v3 (AVX2)
	if (features & X86_64_V3_Features) == X86_64_V3_Features {
		return format.X86_64_V3, "v3"
	}

	// Check for v2 (SSE4.2)
	if (features & X86_64_V2_Features) == X86_64_V2_Features {
		return format.X86_64_V2, "v2"
	}

	// Default to v1 (baseline)
	return format.X86_64_V1, "v1"
}
