// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package cpu

import (
	"testing"

	"github.com/DataDog/adipo/internal/format"
)

func TestDetectX86AliasNewCPUs(t *testing.T) {
	tests := []struct {
		name      string
		family    int
		model     int
		modelName string
		want      string
	}{
		{"cannon lake", 6, 102, "Intel(R) Core(TM) i3-8121U CPU", "cannonlake"},
		{"tiger lake", 6, 140, "Intel(R) Core(TM) i7-1165G7 CPU", "tigerlake"},
		{"rocket lake", 6, 167, "Intel(R) Core(TM) i9-11900K CPU", "rocketlake"},
		{"alder lake", 6, 151, "Intel(R) Core(TM) i9-12900K CPU", "alderlake"},
		{"raptor lake", 6, 183, "Intel(R) Core(TM) i9-13900K CPU", "raptorlake"},
		{"meteor lake", 6, 170, "Intel(R) Core(TM) Ultra 7 165H", "meteorlake"},
		{"arrow lake", 6, 181, "Intel(R) Core(TM) Ultra 9 285K", "arrowlake"},
		{"arrow lake s", 6, 198, "Intel(R) Core(TM) Ultra 7", "arrowlake-s"},
		{"lunar lake", 6, 189, "Intel(R) Core(TM) Ultra 7 258V", "lunarlake"},
		{"sapphire rapids", 6, 143, "Intel(R) Xeon(R) Platinum 8488C", "sapphirerapids"},
		{"emerald rapids", 6, 207, "Intel(R) Xeon(R) Platinum 8592+", "emeraldrapids"},
		{"granite rapids", 6, 173, "Intel(R) Xeon(R) 6980P", "graniterapids"},
		{"granite rapids d", 6, 174, "Intel(R) Xeon(R) 6740E", "graniterapids-d"},
		{"sierra forest", 6, 175, "Intel(R) Xeon(R) 6780E", "sierraforest"},
		{"zen 5", 26, 32, "AMD EPYC 9965 192-Core Processor", "zen5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &CPUModel{Family: tt.family, Model: tt.model, ModelName: tt.modelName}
			if got := DetectCPUAlias(model, format.ArchX86_64); got != tt.want {
				t.Errorf("DetectCPUAlias() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectARMAliasNewCPUs(t *testing.T) {
	tests := []struct {
		name        string
		implementer int
		partNum     int
		want        string
	}{
		{"cortex a520", 0x41, 0xd80, "cortex-a520"},
		{"cortex x925", 0x41, 0xd85, "cortex-x925"},
		{"cortex a725", 0x41, 0xd87, "cortex-a725"},
		{"neoverse v3", 0x41, 0xd84, "neoverse-v3"},
		{"neoverse n3", 0x41, 0xd8e, "neoverse-n3"},
		{"c1 pro", 0x41, 0xd8b, "c1-pro"},
		{"c1 ultra", 0x41, 0xd8c, "c1-ultra"},
		{"c1 premium", 0x41, 0xd90, "c1-premium"},
		{"ampere one a", 0xc0, 0xac4, "ampere1a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &CPUModel{Implementer: tt.implementer, PartNum: tt.partNum}
			if got := DetectCPUAlias(model, format.ArchARM64); got != tt.want {
				t.Errorf("DetectCPUAlias() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectAppleSiliconAliasM5(t *testing.T) {
	model := &CPUModel{BrandString: "Apple M5 Max"}
	if got := DetectCPUAlias(model, format.ArchARM64); got != "apple-m5" {
		t.Errorf("DetectCPUAlias() = %q, want %q", got, "apple-m5")
	}
}

func TestValidateNewCPUHints(t *testing.T) {
	tests := []struct {
		name string
		hint string
		arch format.Architecture
	}{
		{"x86 server", "graniterapids", format.ArchX86_64},
		{"x86 client", "lunarlake", format.ArchX86_64},
		{"amd zen 5", "zen5", format.ArchX86_64},
		{"arm neoverse", "neoverse-v3", format.ArchARM64},
		{"arm cloud", "graviton5", format.ArchARM64},
		{"arm cloud google", "google-axion-n4a", format.ArchARM64},
		{"arm cloud azure", "azure-cobalt100", format.ArchARM64},
		{"arm cloud nvidia", "nvidia-grace", format.ArchARM64},
		{"apple", "apple-m5", format.ArchARM64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ValidateCPUHint(tt.hint, tt.arch); err != nil {
				t.Fatalf("ValidateCPUHint(%q, %s) returned error: %v", tt.hint, tt.arch, err)
			}
		})
	}
}

func TestListValidAliasesIncludesNewCPUs(t *testing.T) {
	x86Aliases := toAliasSet(ListValidAliases(format.ArchX86_64))
	for _, alias := range []string{
		"cannonlake",
		"tigerlake",
		"rocketlake",
		"alderlake",
		"raptorlake",
		"meteorlake",
		"arrowlake",
		"arrowlake-s",
		"lunarlake",
		"sapphirerapids",
		"emeraldrapids",
		"graniterapids",
		"graniterapids-d",
		"sierraforest",
		"zen5",
	} {
		if !x86Aliases[alias] {
			t.Errorf("ListValidAliases(x86-64) missing %q", alias)
		}
	}

	armAliases := toAliasSet(ListValidAliases(format.ArchARM64))
	for _, alias := range []string{
		"cortex-a520",
		"cortex-x925",
		"cortex-a725",
		"neoverse-v3",
		"neoverse-n3",
		"c1-pro",
		"c1-ultra",
		"c1-premium",
		"graviton4",
		"graviton5",
		"google-axion",
		"google-axion-n4a",
		"azure-cobalt100",
		"nvidia-grace",
		"ampere1a",
		"apple-m5",
	} {
		if !armAliases[alias] {
			t.Errorf("ListValidAliases(aarch64) missing %q", alias)
		}
	}
}

func toAliasSet(aliases []string) map[string]bool {
	result := make(map[string]bool, len(aliases))
	for _, alias := range aliases {
		result[alias] = true
	}
	return result
}
