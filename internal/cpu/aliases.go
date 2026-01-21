// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package cpu

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DataDog/adipo/internal/format"
)

// CPUAlias represents a CPU alias mapping
type CPUAlias struct {
	Name         string                // GCC-style CPU name (e.g., "skylake", "zen3", "apple-m3")
	Architecture format.Architecture   // Architecture this alias applies to
	MinVersion   format.ArchVersion    // Minimum version this CPU supports
}

// X86CPUAlias maps x86-64 CPU family:model to aliases
type X86CPUAlias struct {
	Family      int      // CPU family
	Model       int      // CPU model (0 means match any model in family)
	NamePattern string   // Substring match in model name (empty = match any)
	Alias       CPUAlias
}

// X86CPUAliases maps Intel/AMD CPUs to GCC-style aliases
var X86CPUAliases = []X86CPUAlias{
	// Intel Haswell (Family 6, Model 60, 69, 70, 71)
	{6, 60, "", CPUAlias{"haswell", format.ArchX86_64, format.X86_64_V3}},
	{6, 69, "", CPUAlias{"haswell", format.ArchX86_64, format.X86_64_V3}},
	{6, 70, "", CPUAlias{"haswell", format.ArchX86_64, format.X86_64_V3}},
	{6, 71, "", CPUAlias{"haswell", format.ArchX86_64, format.X86_64_V3}},

	// Intel Broadwell (Family 6, Model 61, 71, 79, 86)
	{6, 61, "", CPUAlias{"broadwell", format.ArchX86_64, format.X86_64_V3}},
	{6, 79, "", CPUAlias{"broadwell", format.ArchX86_64, format.X86_64_V3}},
	{6, 86, "", CPUAlias{"broadwell", format.ArchX86_64, format.X86_64_V3}},

	// Intel Skylake (Family 6, Model 78, 94)
	{6, 78, "", CPUAlias{"skylake", format.ArchX86_64, format.X86_64_V3}},
	{6, 94, "", CPUAlias{"skylake", format.ArchX86_64, format.X86_64_V3}},

	// Intel Skylake-X/Cascade Lake (Family 6, Model 85) - with AVX-512
	{6, 85, "Platinum", CPUAlias{"skylake-avx512", format.ArchX86_64, format.X86_64_V4}},
	{6, 85, "Gold", CPUAlias{"skylake-avx512", format.ArchX86_64, format.X86_64_V4}},
	{6, 85, "Silver", CPUAlias{"skylake-avx512", format.ArchX86_64, format.X86_64_V4}},
	{6, 85, "W", CPUAlias{"skylake-avx512", format.ArchX86_64, format.X86_64_V4}},
	{6, 85, "", CPUAlias{"skylake", format.ArchX86_64, format.X86_64_V3}}, // Fallback for non-AVX512

	// Intel Ice Lake (Family 6, Model 106, 108, 125, 126)
	{6, 106, "", CPUAlias{"icelake", format.ArchX86_64, format.X86_64_V4}},
	{6, 108, "", CPUAlias{"icelake", format.ArchX86_64, format.X86_64_V4}},
	{6, 125, "", CPUAlias{"icelake", format.ArchX86_64, format.X86_64_V4}},
	{6, 126, "", CPUAlias{"icelake", format.ArchX86_64, format.X86_64_V4}},

	// AMD Zen (Family 23, Model 1, 17, 24)
	{23, 1, "", CPUAlias{"zen", format.ArchX86_64, format.X86_64_V3}},
	{23, 17, "", CPUAlias{"zen", format.ArchX86_64, format.X86_64_V3}},
	{23, 24, "", CPUAlias{"zen", format.ArchX86_64, format.X86_64_V3}},

	// AMD Zen 2 (Family 23, Model 49, 96, 113)
	{23, 49, "", CPUAlias{"zen2", format.ArchX86_64, format.X86_64_V3}},
	{23, 96, "", CPUAlias{"zen2", format.ArchX86_64, format.X86_64_V3}},
	{23, 113, "", CPUAlias{"zen2", format.ArchX86_64, format.X86_64_V3}},

	// AMD Zen 3 (Family 25, Model 33, 80)
	{25, 33, "", CPUAlias{"zen3", format.ArchX86_64, format.X86_64_V3}},
	{25, 80, "", CPUAlias{"zen3", format.ArchX86_64, format.X86_64_V3}},

	// AMD Zen 4 (Family 25, Model 97, 104)
	{25, 97, "", CPUAlias{"zen4", format.ArchX86_64, format.X86_64_V4}},
	{25, 104, "", CPUAlias{"zen4", format.ArchX86_64, format.X86_64_V4}},
}

// ARMCPUAlias maps ARM64 implementer:part to aliases
type ARMCPUAlias struct {
	Implementer int      // ARM implementer ID
	PartNum     int      // ARM part number
	Alias       CPUAlias
}

// ARMCPUAliases maps ARM CPUs to GCC-style aliases
var ARMCPUAliases = []ARMCPUAlias{
	// ARM Cortex
	{0x41, 0xd07, CPUAlias{"cortex-a57", format.ArchARM64, format.ARM64_V8_0}},
	{0x41, 0xd08, CPUAlias{"cortex-a72", format.ArchARM64, format.ARM64_V8_0}},
	{0x41, 0xd09, CPUAlias{"cortex-a73", format.ArchARM64, format.ARM64_V8_0}},
	{0x41, 0xd0a, CPUAlias{"cortex-a75", format.ArchARM64, format.ARM64_V8_2}},
	{0x41, 0xd0b, CPUAlias{"cortex-a76", format.ArchARM64, format.ARM64_V8_2}},

	// ARM Neoverse
	{0x41, 0xd0c, CPUAlias{"neoverse-n1", format.ArchARM64, format.ARM64_V8_2}},
	{0x41, 0xd40, CPUAlias{"neoverse-v1", format.ArchARM64, format.ARM64_V8_4}},
	{0x41, 0xd49, CPUAlias{"neoverse-n2", format.ArchARM64, format.ARM64_V9_0}},
	{0x41, 0xd4f, CPUAlias{"neoverse-v2", format.ArchARM64, format.ARM64_V9_0}},

	// AWS Graviton (same as Neoverse, but keep separate alias)
	{0x41, 0xd0c, CPUAlias{"graviton2", format.ArchARM64, format.ARM64_V8_2}},
	{0x41, 0xd40, CPUAlias{"graviton3", format.ArchARM64, format.ARM64_V8_4}},

	// Ampere
	{0xc0, 0xac3, CPUAlias{"ampere1", format.ArchARM64, format.ARM64_V8_2}},
}

// AppleSiliconAlias represents Apple Silicon CPU aliases
type AppleSiliconAlias struct {
	Pattern string   // Regex pattern to match in brand string
	Alias   CPUAlias
}

// AppleSiliconAliases maps Apple Silicon brand strings to aliases
var AppleSiliconAliases = []AppleSiliconAlias{
	{`Apple M1`, CPUAlias{"apple-m1", format.ArchARM64, format.ARM64_V8_0}},
	{`Apple M2`, CPUAlias{"apple-m2", format.ArchARM64, format.ARM64_V8_0}},
	{`Apple M3`, CPUAlias{"apple-m3", format.ArchARM64, format.ARM64_V8_0}},
	{`Apple M4`, CPUAlias{"apple-m4", format.ArchARM64, format.ARM64_V8_0}},
}

// DetectCPUAlias attempts to determine CPU alias from model information
// Returns empty string if no match found
func DetectCPUAlias(model *CPUModel, arch format.Architecture) string {
	if model == nil {
		return ""
	}

	switch arch {
	case format.ArchX86_64:
		return detectX86Alias(model)
	case format.ArchARM64:
		return detectARMAlias(model)
	default:
		return ""
	}
}

// detectX86Alias detects x86-64 CPU alias from model info
func detectX86Alias(model *CPUModel) string {
	// Try to match family:model with optional name pattern
	for _, alias := range X86CPUAliases {
		// Check family and model match
		if model.Family != alias.Family {
			continue
		}
		if alias.Model != 0 && model.Model != alias.Model {
			continue
		}

		// Check name pattern if specified
		if alias.NamePattern != "" {
			if model.ModelName == "" || !strings.Contains(model.ModelName, alias.NamePattern) {
				continue
			}
		}

		// Found a match
		return alias.Alias.Name
	}

	return ""
}

// detectARMAlias detects ARM64 CPU alias from model info
func detectARMAlias(model *CPUModel) string {
	// macOS: Try Apple Silicon brand string patterns
	if model.BrandString != "" {
		for _, alias := range AppleSiliconAliases {
			matched, _ := regexp.MatchString(alias.Pattern, model.BrandString)
			if matched {
				return alias.Alias.Name
			}
		}
	}

	// Linux: Try implementer:part lookup
	if model.Implementer != 0 && model.PartNum != 0 {
		for _, alias := range ARMCPUAliases {
			if model.Implementer == alias.Implementer && model.PartNum == alias.PartNum {
				return alias.Alias.Name
			}
		}
	}

	return ""
}

// ValidateCPUHint validates a user-provided CPU hint
// Returns the corresponding CPUAlias or error if hint is unknown or incompatible
func ValidateCPUHint(hint string, arch format.Architecture) (*CPUAlias, error) {
	hint = strings.ToLower(strings.TrimSpace(hint))
	if hint == "" {
		return nil, fmt.Errorf("CPU hint cannot be empty")
	}

	// Search for hint in all alias tables
	switch arch {
	case format.ArchX86_64:
		for _, x86Alias := range X86CPUAliases {
			if x86Alias.Alias.Name == hint {
				return &x86Alias.Alias, nil
			}
		}

	case format.ArchARM64:
		// Check ARM Linux aliases
		for _, armAlias := range ARMCPUAliases {
			if armAlias.Alias.Name == hint {
				return &armAlias.Alias, nil
			}
		}

		// Check Apple Silicon aliases
		for _, appleAlias := range AppleSiliconAliases {
			if appleAlias.Alias.Name == hint {
				return &appleAlias.Alias, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown CPU hint %q for architecture %s", hint, arch.String())
}

// ListValidAliases returns all valid CPU aliases for the given architecture
func ListValidAliases(arch format.Architecture) []string {
	var aliases []string
	seen := make(map[string]bool)

	switch arch {
	case format.ArchX86_64:
		for _, x86Alias := range X86CPUAliases {
			if !seen[x86Alias.Alias.Name] {
				aliases = append(aliases, x86Alias.Alias.Name)
				seen[x86Alias.Alias.Name] = true
			}
		}

	case format.ArchARM64:
		for _, armAlias := range ARMCPUAliases {
			if !seen[armAlias.Alias.Name] {
				aliases = append(aliases, armAlias.Alias.Name)
				seen[armAlias.Alias.Name] = true
			}
		}
		for _, appleAlias := range AppleSiliconAliases {
			if !seen[appleAlias.Alias.Name] {
				aliases = append(aliases, appleAlias.Alias.Name)
				seen[appleAlias.Alias.Name] = true
			}
		}
	}

	return aliases
}
