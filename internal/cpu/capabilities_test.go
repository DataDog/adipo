// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package cpu

import (
	"strings"
	"testing"

	"github.com/DataDog/adipo/internal/format"
)

func TestNewCapabilities(t *testing.T) {
	caps := NewCapabilities("x86-64")

	if caps.Architecture != "x86-64" {
		t.Errorf("Architecture = %q, want %q", caps.Architecture, "x86-64")
	}

	if caps.Features == nil {
		t.Error("Features map should be initialized")
	}
}

func TestHasFeature(t *testing.T) {
	caps := NewCapabilities("x86-64")
	caps.Features["sse4_2"] = struct{}{}
	caps.Features["avx2"] = struct{}{}

	tests := []struct {
		name    string
		feature string
		want    bool
	}{
		{"existing feature", "sse4_2", true},
		{"another existing feature", "avx2", true},
		{"missing feature", "avx512f", false},
		{"empty feature", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := caps.HasFeature(tt.feature); got != tt.want {
				t.Errorf("HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}
}

func TestHasAllFeatures(t *testing.T) {
	tests := []struct {
		name         string
		featureMask  uint64
		requiredMask uint64
		want         bool
	}{
		{"all features present", 0b1111, 0b1010, true},
		{"exact match", 0b1010, 0b1010, true},
		{"missing some features", 0b1010, 0b1111, false},
		{"no features required", 0b1010, 0b0000, true},
		{"no features present", 0b0000, 0b1010, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := &Capabilities{FeatureMask: tt.featureMask}
			if got := caps.HasAllFeatures(tt.requiredMask); got != tt.want {
				t.Errorf("HasAllFeatures(0b%b) with mask 0b%b = %v, want %v",
					tt.requiredMask, tt.featureMask, got, tt.want)
			}
		})
	}
}

func TestIsCompatibleWith(t *testing.T) {
	tests := []struct {
		name       string
		caps       *Capabilities
		arch       format.Architecture
		version    format.ArchVersion
		features   uint64
		compatible bool
	}{
		{
			name: "compatible x86-64 v3 with v2 binary",
			caps: &Capabilities{
				ArchType:    format.ArchX86_64,
				Version:     format.X86_64_V3,
				FeatureMask: 0b1111,
			},
			arch:       format.ArchX86_64,
			version:    format.X86_64_V2,
			features:   0b1010,
			compatible: true,
		},
		{
			name: "incompatible architecture",
			caps: &Capabilities{
				ArchType:    format.ArchX86_64,
				Version:     format.X86_64_V3,
				FeatureMask: 0b1111,
			},
			arch:       format.ArchARM64,
			version:    format.ARM64_V8_0,
			features:   0b1010,
			compatible: false,
		},
		{
			name: "incompatible version (too new)",
			caps: &Capabilities{
				ArchType:    format.ArchX86_64,
				Version:     format.X86_64_V2,
				FeatureMask: 0b1111,
			},
			arch:       format.ArchX86_64,
			version:    format.X86_64_V4,
			features:   0b0000,
			compatible: false,
		},
		{
			name: "missing required features",
			caps: &Capabilities{
				ArchType:    format.ArchX86_64,
				Version:     format.X86_64_V3,
				FeatureMask: 0b1010,
			},
			arch:       format.ArchX86_64,
			version:    format.X86_64_V2,
			features:   0b1111,
			compatible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.caps.IsCompatibleWith(tt.arch, tt.version, tt.features)
			if got != tt.compatible {
				t.Errorf("IsCompatibleWith() = %v, want %v", got, tt.compatible)
			}
		})
	}
}

func TestString(t *testing.T) {
	caps := &Capabilities{
		Architecture: "x86-64",
		VersionStr:   "v3",
		ArchType:     format.ArchX86_64,
		Version:      format.X86_64_V3,
	}

	str := caps.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Should contain architecture and version info
	if !strings.Contains(str, "x86-64") {
		t.Errorf("String() = %q, should contain 'x86-64'", str)
	}
	if !strings.Contains(str, "v3") {
		t.Errorf("String() = %q, should contain 'v3'", str)
	}
}

func TestFeatureList(t *testing.T) {
	caps := NewCapabilities("x86-64")
	caps.Features["sse4_2"] = struct{}{}
	caps.Features["avx2"] = struct{}{}
	caps.Features["avx512f"] = struct{}{}

	features := caps.FeatureList()

	if len(features) != 3 {
		t.Errorf("FeatureList() returned %d features, want 3", len(features))
	}

	// Check that all features are present (order may vary)
	featureSet := make(map[string]bool)
	for _, f := range features {
		featureSet[f] = true
	}

	expectedFeatures := []string{"sse4_2", "avx2", "avx512f"}
	for _, expected := range expectedFeatures {
		if !featureSet[expected] {
			t.Errorf("FeatureList() missing feature %q", expected)
		}
	}
}

func TestFeatureList_Empty(t *testing.T) {
	caps := NewCapabilities("x86-64")
	features := caps.FeatureList()

	if len(features) != 0 {
		t.Errorf("FeatureList() for empty features = %v, want empty slice", features)
	}
}
