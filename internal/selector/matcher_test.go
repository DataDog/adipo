// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package selector

import (
	"testing"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

func TestIsArchitectureMatch(t *testing.T) {
	tests := []struct {
		name     string
		cpuArch  format.Architecture
		binArch  format.Architecture
		wantMatch bool
	}{
		{"same x86-64", format.ArchX86_64, format.ArchX86_64, true},
		{"same ARM64", format.ArchARM64, format.ArchARM64, true},
		{"different arch", format.ArchX86_64, format.ArchARM64, false},
		{"unknown cpu", format.ArchUnknown, format.ArchX86_64, false},
		{"unknown binary", format.ArchX86_64, format.ArchUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := &cpu.Capabilities{ArchType: tt.cpuArch}
			bin := &format.BinaryMetadata{Architecture: tt.binArch}
			m := NewMatcher(caps)

			if got := m.IsArchitectureMatch(bin); got != tt.wantMatch {
				t.Errorf("IsArchitectureMatch() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestIsVersionCompatible(t *testing.T) {
	tests := []struct {
		name       string
		cpuVer     format.ArchVersion
		binVer     format.ArchVersion
		compatible bool
	}{
		// x86-64: higher CPU can run lower binary
		{"x86-64 v3 runs v1", format.X86_64_V3, format.X86_64_V1, true},
		{"x86-64 v3 runs v2", format.X86_64_V3, format.X86_64_V2, true},
		{"x86-64 v3 runs v3", format.X86_64_V3, format.X86_64_V3, true},
		{"x86-64 v1 cannot run v3", format.X86_64_V1, format.X86_64_V3, false},
		{"x86-64 v2 cannot run v4", format.X86_64_V2, format.X86_64_V4, false},

		// ARM64: higher CPU can run lower binary
		{"arm64 v8.2 runs v8.0", format.ARM64_V8_2, format.ARM64_V8_0, true},
		{"arm64 v9.0 runs v8.0", format.ARM64_V9_0, format.ARM64_V8_0, true},
		{"arm64 v8.0 cannot run v8.2", format.ARM64_V8_0, format.ARM64_V8_2, false},
		{"arm64 v8.1 cannot run v9.0", format.ARM64_V8_1, format.ARM64_V9_0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine arch from version constant
			var archType format.Architecture
			if tt.cpuVer >= format.X86_64_V1 && tt.cpuVer <= format.X86_64_V4 {
				archType = format.ArchX86_64
			} else {
				archType = format.ArchARM64
			}

			caps := &cpu.Capabilities{ArchType: archType, Version: tt.cpuVer}
			bin := &format.BinaryMetadata{Architecture: archType, ArchVersion: tt.binVer}
			m := NewMatcher(caps)

			if got := m.IsVersionCompatible(bin); got != tt.compatible {
				t.Errorf("IsVersionCompatible() = %v, want %v", got, tt.compatible)
			}
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name       string
		caps       *cpu.Capabilities
		bin        *format.BinaryMetadata
		compatible bool
	}{
		{
			name: "compatible v3 cpu with v2 binary",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
			},
			bin: &format.BinaryMetadata{
				Architecture: format.ArchX86_64,
				ArchVersion:  format.X86_64_V2,
			},
			compatible: true,
		},
		{
			name: "incompatible architecture mismatch",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
			},
			bin: &format.BinaryMetadata{
				Architecture: format.ArchARM64,
				ArchVersion:  format.ARM64_V8_0,
			},
			compatible: false,
		},
		{
			name: "incompatible version too new",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V2,
			},
			bin: &format.BinaryMetadata{
				Architecture: format.ArchX86_64,
				ArchVersion:  format.X86_64_V4,
			},
			compatible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.caps)
			if got := m.IsCompatible(tt.bin); got != tt.compatible {
				t.Errorf("IsCompatible() = %v, want %v", got, tt.compatible)
			}
		})
	}
}

func TestFilterCompatible(t *testing.T) {
	caps := &cpu.Capabilities{
		ArchType: format.ArchX86_64,
		Version:  format.X86_64_V3,
	}

	binaries := []*format.BinaryMetadata{
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V1}, // compatible
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V2}, // compatible
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V3}, // compatible
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V4}, // incompatible
		{Architecture: format.ArchARM64, ArchVersion: format.ARM64_V8_0}, // incompatible
	}

	m := NewMatcher(caps)
	compatible := m.FilterCompatible(binaries)

	if len(compatible) != 3 {
		t.Errorf("FilterCompatible() returned %d binaries, want 3", len(compatible))
	}

	// Verify the incompatible ones were filtered out
	for _, bin := range compatible {
		if bin.Architecture == format.ArchARM64 {
			t.Error("ARM64 binary should have been filtered out")
		}
		if bin.ArchVersion == format.X86_64_V4 {
			t.Error("v4 binary should have been filtered out on v3 CPU")
		}
	}
}
