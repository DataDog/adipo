// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package selector

import (
	"strings"
	"testing"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

func TestSelectBinary(t *testing.T) {
	tests := []struct {
		name      string
		caps      *cpu.Capabilities
		metadata  []*format.BinaryMetadata
		wantIndex int
		wantError bool
	}{
		{
			name: "select best matching version",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
			},
			metadata: []*format.BinaryMetadata{
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V1},
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V2},
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V3},
			},
			wantIndex: 2, // v3 is best match
			wantError: false,
		},
		{
			name: "select lower version when exact not available",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V4,
			},
			metadata: []*format.BinaryMetadata{
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V1},
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V3},
			},
			wantIndex: 1, // v3 is highest available
			wantError: false,
		},
		{
			name: "no compatible binaries",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V2,
			},
			metadata: []*format.BinaryMetadata{
				{Architecture: format.ArchARM64, ArchVersion: format.ARM64_V8_0},
				{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V4}, // too new
			},
			wantError: true,
		},
		{
			name:      "empty metadata",
			caps:      &cpu.Capabilities{ArchType: format.ArchX86_64},
			metadata:  []*format.BinaryMetadata{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := NewSelector(tt.caps, tt.metadata)
			index, _, err := sel.SelectBinary()

			if tt.wantError {
				if err == nil {
					t.Error("SelectBinary() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("SelectBinary() unexpected error: %v", err)
				return
			}

			if index != tt.wantIndex {
				t.Errorf("SelectBinary() index = %d, want %d", index, tt.wantIndex)
			}
		})
	}
}

func TestSelectBinaryVerbose(t *testing.T) {
	caps := &cpu.Capabilities{
		ArchType: format.ArchX86_64,
		Version:  format.X86_64_V3,
	}

	metadata := []*format.BinaryMetadata{
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V1},
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V2},
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V3},
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V4}, // incompatible
	}

	sel := NewSelector(caps, metadata)
	result, err := sel.SelectBinaryVerbose()

	if err != nil {
		t.Fatalf("SelectBinaryVerbose() error: %v", err)
	}

	if result.SelectedIndex != 2 {
		t.Errorf("SelectedIndex = %d, want 2", result.SelectedIndex)
	}

	if result.SelectedBinary.ArchVersion != format.X86_64_V3 {
		t.Errorf("Selected wrong binary version: %v", result.SelectedBinary.ArchVersion)
	}

	if result.SelectedScore <= 0 {
		t.Errorf("SelectedScore = %d, want > 0", result.SelectedScore)
	}

	// Should have scores for all binaries
	if len(result.AllScores) != len(metadata) {
		t.Errorf("AllScores length = %d, want %d", len(result.AllScores), len(metadata))
	}

	// Should have 3 compatible binaries (not v4)
	if result.CompatibleBinaries != 3 {
		t.Errorf("CompatibleBinaries = %d, want 3", result.CompatibleBinaries)
	}

	// Check that String() method works
	str := result.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	if !strings.Contains(str, "Selected") {
		t.Error("String() should contain 'Selected'")
	}
}

func TestSelectBinaryVerbose_NoCompatible(t *testing.T) {
	caps := &cpu.Capabilities{
		ArchType: format.ArchX86_64,
		Version:  format.X86_64_V1,
	}

	metadata := []*format.BinaryMetadata{
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V4},
		{Architecture: format.ArchARM64, ArchVersion: format.ARM64_V8_0},
	}

	sel := NewSelector(caps, metadata)
	_, err := sel.SelectBinaryVerbose()

	if err == nil {
		t.Error("SelectBinaryVerbose() expected error for no compatible binaries")
	}
}

func TestSelectBinaryWithCPUAlias(t *testing.T) {
	tests := []struct {
		name      string
		caps      *cpu.Capabilities
		metadata  []*format.BinaryMetadata
		wantIndex int
		wantError bool
	}{
		{
			name: "prefer binary with matching CPU hint over same version without hint",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
				CPUModel: &cpu.CPUModel{
					Vendor: "AuthenticAMD",
					Family: 25, // Zen 3
					Model:  33,
				},
			},
			metadata: func() []*format.BinaryMetadata {
				// Create three v3 binaries with different CPU hints
				binNoHint := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}
				binZen3 := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}
				binSkyLake := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}

				// Set CPU hints
				_ = binZen3.SetCPUHint("zen3")
				_ = binSkyLake.SetCPUHint("skylake")

				return []*format.BinaryMetadata{binNoHint, binZen3, binSkyLake}
			}(),
			wantIndex: 1, // zen3 binary should be preferred
			wantError: false,
		},
		{
			name: "prefer binary with matching CPU hint over wrong hint",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
				CPUModel: &cpu.CPUModel{
					Vendor: "GenuineIntel",
					Family: 6,  // Skylake
					Model:  94,
				},
			},
			metadata: func() []*format.BinaryMetadata {
				binZen3 := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}
				binSkyLake := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}

				_ = binZen3.SetCPUHint("zen3")
				_ = binSkyLake.SetCPUHint("skylake")

				return []*format.BinaryMetadata{binZen3, binSkyLake}
			}(),
			wantIndex: 1, // skylake binary should be preferred
			wantError: false,
		},
		{
			name: "no CPU model detected - fallback to version selection",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
				CPUModel: nil, // No CPU model detected
			},
			metadata: func() []*format.BinaryMetadata {
				binV2 := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V2,
				}
				binV3WithHint := &format.BinaryMetadata{
					Architecture: format.ArchX86_64,
					ArchVersion:  format.X86_64_V3,
				}

				_ = binV3WithHint.SetCPUHint("zen3")

				return []*format.BinaryMetadata{binV2, binV3WithHint}
			}(),
			wantIndex: 1, // v3 still selected (no CPU alias bonus, but higher version)
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := NewSelector(tt.caps, tt.metadata)
			index, binary, err := sel.SelectBinary()

			if tt.wantError {
				if err == nil {
					t.Error("SelectBinary() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("SelectBinary() unexpected error: %v", err)
				return
			}

			if index != tt.wantIndex {
				t.Errorf("SelectBinary() index = %d, want %d", index, tt.wantIndex)

				// Show what was actually selected for debugging
				if binary != nil {
					hint := binary.GetCPUHint()
					t.Logf("Selected binary with hint: %q", hint)
				}
			}
		})
	}
}
