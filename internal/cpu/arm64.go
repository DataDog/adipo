// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


//go:build arm64

package cpu

import (
	"bufio"
	"os"
	"strings"

	"github.com/DataDog/adipo/internal/format"
	"golang.org/x/sys/cpu"
)

// DetectARM64 detects ARM64 CPU capabilities
func DetectARM64() (*Capabilities, error) {
	caps := NewCapabilities("aarch64")
	caps.ArchType = format.ArchARM64

	// Build feature mask and map
	var featureMask uint64

	// Use golang.org/x/sys/cpu for ARM64 detection
	// Note: Not all features are available, so we'll also parse /proc/cpuinfo

	// FP (Floating Point) - baseline
	featureMask |= ARM_FP
	caps.Features["fp"] = struct{}{}

	// ASIMD (Advanced SIMD) - baseline
	featureMask |= ARM_ASIMD
	caps.Features["asimd"] = struct{}{}

	// AES
	if cpu.ARM64.HasAES {
		featureMask |= ARM_AES
		caps.Features["aes"] = struct{}{}
	}

	// PMULL
	if cpu.ARM64.HasPMULL {
		featureMask |= ARM_PMULL
		caps.Features["pmull"] = struct{}{}
	}

	// SHA1
	if cpu.ARM64.HasSHA1 {
		featureMask |= ARM_SHA1
		caps.Features["sha1"] = struct{}{}
	}

	// SHA2
	if cpu.ARM64.HasSHA2 {
		featureMask |= ARM_SHA2
		caps.Features["sha2"] = struct{}{}
	}

	// CRC32
	if cpu.ARM64.HasCRC32 {
		featureMask |= ARM_CRC32
		caps.Features["crc32"] = struct{}{}
	}

	// ATOMICS (LSE)
	if cpu.ARM64.HasATOMICS {
		featureMask |= ARM_ATOMICS
		caps.Features["atomics"] = struct{}{}
	}

	// CPUID
	if cpu.ARM64.HasCPUID {
		featureMask |= ARM_CPUID
		caps.Features["cpuid"] = struct{}{}
	}

	// SVE
	if cpu.ARM64.HasSVE {
		featureMask |= ARM_SVE
		caps.Features["sve"] = struct{}{}
	}

	// SVE2
	if cpu.ARM64.HasSVE2 {
		featureMask |= ARM_SVE2
		caps.Features["sve2"] = struct{}{}
	}

	// Additional features from /proc/cpuinfo
	// Non-fatal, continue with what we have
	_ = detectFromCPUInfo(caps, &featureMask)

	caps.FeatureMask = featureMask

	// Determine ARM version
	caps.Version, caps.VersionStr = detectARMVersion(featureMask)

	return caps, nil
}

// detectFromCPUInfo reads additional features from /proc/cpuinfo
func detectFromCPUInfo(caps *Capabilities, featureMask *uint64) error {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Features") {
			// Parse features line
			// Format: "Features	: fp asimd evtstrm aes pmull sha1 sha2 crc32 atomics ..."
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}

			features := strings.Fields(parts[1])
			for _, feature := range features {
				feature = strings.ToLower(strings.TrimSpace(feature))
				parseFeatureFromName(feature, caps, featureMask)
			}
		}
	}

	return scanner.Err()
}

// parseFeatureFromName updates caps and featureMask based on feature name
func parseFeatureFromName(feature string, caps *Capabilities, featureMask *uint64) {
	// Map common feature names to our bitmasks
	switch feature {
	case "fp":
		*featureMask |= ARM_FP
		caps.Features["fp"] = struct{}{}
	case "asimd":
		*featureMask |= ARM_ASIMD
		caps.Features["asimd"] = struct{}{}
	case "aes":
		*featureMask |= ARM_AES
		caps.Features["aes"] = struct{}{}
	case "pmull":
		*featureMask |= ARM_PMULL
		caps.Features["pmull"] = struct{}{}
	case "sha1":
		*featureMask |= ARM_SHA1
		caps.Features["sha1"] = struct{}{}
	case "sha2":
		*featureMask |= ARM_SHA2
		caps.Features["sha2"] = struct{}{}
	case "crc32":
		*featureMask |= ARM_CRC32
		caps.Features["crc32"] = struct{}{}
	case "atomics":
		*featureMask |= ARM_ATOMICS
		caps.Features["atomics"] = struct{}{}
	case "fphp":
		*featureMask |= ARM_FP16
		caps.Features["fphp"] = struct{}{}
	case "asimdhp":
		*featureMask |= ARM_ASIMDHP
		caps.Features["asimdhp"] = struct{}{}
	case "cpuid":
		*featureMask |= ARM_CPUID
		caps.Features["cpuid"] = struct{}{}
	case "asimdrdm":
		*featureMask |= ARM_ASIMDRDM
		caps.Features["asimdrdm"] = struct{}{}
	case "jscvt":
		*featureMask |= ARM_JSCVT
		caps.Features["jscvt"] = struct{}{}
	case "fcma":
		*featureMask |= ARM_FCMA
		caps.Features["fcma"] = struct{}{}
	case "lrcpc":
		*featureMask |= ARM_LRCPC
		caps.Features["lrcpc"] = struct{}{}
	case "dcpop":
		*featureMask |= ARM_DCPOP
		caps.Features["dcpop"] = struct{}{}
	case "sha3":
		*featureMask |= ARM_SHA3
		caps.Features["sha3"] = struct{}{}
	case "sm3":
		*featureMask |= ARM_SM3
		caps.Features["sm3"] = struct{}{}
	case "sm4":
		*featureMask |= ARM_SM4
		caps.Features["sm4"] = struct{}{}
	case "asimddp", "dotprod":
		*featureMask |= ARM_ASIMDDP
		caps.Features["asimddp"] = struct{}{}
	case "sha512":
		*featureMask |= ARM_SHA512
		caps.Features["sha512"] = struct{}{}
	case "sve":
		*featureMask |= ARM_SVE
		caps.Features["sve"] = struct{}{}
	case "asimdfhm":
		*featureMask |= ARM_ASIMDFHM
		caps.Features["asimdfhm"] = struct{}{}
	case "dit":
		*featureMask |= ARM_DIT
		caps.Features["dit"] = struct{}{}
	case "ilrcpc":
		*featureMask |= ARM_ILRCPC
		caps.Features["ilrcpc"] = struct{}{}
	case "flagm":
		*featureMask |= ARM_FLAGM
		caps.Features["flagm"] = struct{}{}
	case "ssbs":
		*featureMask |= ARM_SSBS
		caps.Features["ssbs"] = struct{}{}
	case "sb":
		*featureMask |= ARM_SB
		caps.Features["sb"] = struct{}{}
	case "paca":
		*featureMask |= ARM_PACA
		caps.Features["paca"] = struct{}{}
	case "pacg":
		*featureMask |= ARM_PACG
		caps.Features["pacg"] = struct{}{}
	case "sve2":
		*featureMask |= ARM_SVE2
		caps.Features["sve2"] = struct{}{}
	case "sveaes":
		*featureMask |= ARM_SVEAES
		caps.Features["sveaes"] = struct{}{}
	case "svepmull":
		*featureMask |= ARM_SVEPMULL
		caps.Features["svepmull"] = struct{}{}
	case "svebitperm":
		*featureMask |= ARM_SVEBITPERM
		caps.Features["svebitperm"] = struct{}{}
	case "svesha3":
		*featureMask |= ARM_SVESHA3
		caps.Features["svesha3"] = struct{}{}
	case "svesm4":
		*featureMask |= ARM_SVESM4
		caps.Features["svesm4"] = struct{}{}
	case "flagm2":
		*featureMask |= ARM_FLAGM2
		caps.Features["flagm2"] = struct{}{}
	case "frint":
		*featureMask |= ARM_FRINT
		caps.Features["frint"] = struct{}{}
	case "svei8mm":
		*featureMask |= ARM_SVEI8MM
		caps.Features["svei8mm"] = struct{}{}
	case "svef32mm":
		*featureMask |= ARM_SVEF32MM
		caps.Features["svef32mm"] = struct{}{}
	case "svef64mm":
		*featureMask |= ARM_SVEF64MM
		caps.Features["svef64mm"] = struct{}{}
	case "svebf16":
		*featureMask |= ARM_SVEBF16
		caps.Features["svebf16"] = struct{}{}
	case "i8mm":
		*featureMask |= ARM_I8MM
		caps.Features["i8mm"] = struct{}{}
	case "bf16":
		*featureMask |= ARM_BF16
		caps.Features["bf16"] = struct{}{}
	case "dgh":
		*featureMask |= ARM_DGH
		caps.Features["dgh"] = struct{}{}
	case "rng":
		*featureMask |= ARM_RNG
		caps.Features["rng"] = struct{}{}
	case "bti":
		*featureMask |= ARM_BTI
		caps.Features["bti"] = struct{}{}
	case "mte":
		*featureMask |= ARM_MTE
		caps.Features["mte"] = struct{}{}
	}
}

// detectARMVersion determines the ARM version based on features
func detectARMVersion(features uint64) (format.ArchVersion, string) {
	// SVE2 implies ARMv9.0 or later
	if features&ARM_SVE2 != 0 {
		return format.ARM64_V9_0, "v9.0"
	}

	// SVE implies at least ARMv8.2
	if features&ARM_SVE != 0 {
		return format.ARM64_V8_2, "v8.2"
	}

	// Atomics (LSE) implies ARMv8.1
	if features&ARM_ATOMICS != 0 {
		return format.ARM64_V8_1, "v8.1"
	}

	// Default to ARMv8.0 baseline
	return format.ARM64_V8_0, "v8.0"
}
