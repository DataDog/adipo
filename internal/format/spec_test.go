// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"testing"
)

func TestParseArchSpecWithFeatures(t *testing.T) {
	spec, err := ParseArchSpec("x86-64-v3,avx2,fma")
	if err != nil {
		t.Fatalf("ParseArchSpec() error: %v", err)
	}

	if spec.Architecture != ArchX86_64 {
		t.Errorf("Architecture = %v, want %v", spec.Architecture, ArchX86_64)
	}

	if spec.ArchVersion != X86_64_V3 {
		t.Errorf("ArchVersion = %v, want %v", spec.ArchVersion, X86_64_V3)
	}

	if spec.RequiredFeatures == 0 {
		t.Error("RequiredFeatures should be non-zero with features specified")
	}

	if len(spec.FeatureNames) != 2 {
		t.Errorf("FeatureNames length = %d, want 2", len(spec.FeatureNames))
	}
}

func TestArchSpecString(t *testing.T) {
	tests := []struct {
		name string
		spec *ArchSpec
		want string
	}{
		{
			name: "x86-64 v3",
			spec: &ArchSpec{
				Architecture: ArchX86_64,
				ArchVersion:  X86_64_V3,
			},
			want: "x86-64-v3",
		},
		{
			name: "arm64 v8.0",
			spec: &ArchSpec{
				Architecture: ArchARM64,
				ArchVersion:  ARM64_V8_0,
			},
			want: "aarch64-v8.0",
		},
		{
			name: "with features",
			spec: &ArchSpec{
				Architecture: ArchX86_64,
				ArchVersion:  X86_64_V3,
				FeatureNames: []string{"avx2", "fma"},
			},
			want: "x86-64-v3,avx2,fma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.String()
			if got != tt.want {
				t.Errorf("ArchSpec.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeArchAlias(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"amd64-v1", "x86-64-v1"},
		{"amd64_v1", "x86-64-v1"},
		{"x86-64-v1", "x86-64-v1"},
		{"aarch64-v8.0", "aarch64-v8.0"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeArchAlias(tt.input)
			if got != tt.want {
				t.Errorf("normalizeArchAlias(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseX86Version(t *testing.T) {
	tests := []struct {
		version string
		want    ArchVersion
		wantErr bool
	}{
		{"v1", X86_64_V1, false},
		{"v2", X86_64_V2, false},
		{"v3", X86_64_V3, false},
		{"v4", X86_64_V4, false},
		{"v5", 0, true},
		{"v0", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := parseX86Version(tt.version)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseX86Version(%q) expected error", tt.version)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseX86Version(%q) unexpected error: %v", tt.version, err)
			}

			if got != tt.want {
				t.Errorf("parseX86Version(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestParseARMVersion(t *testing.T) {
	tests := []struct {
		version string
		want    ArchVersion
		wantErr bool
	}{
		{"v8.0", ARM64_V8_0, false},
		{"v8.1", ARM64_V8_1, false},
		{"v8.2", ARM64_V8_2, false},
		{"v8.3", ARM64_V8_3, false},
		{"v9.0", ARM64_V9_0, false},
		{"v9.5", ARM64_V9_5, false},
		{"v7.0", 0, true},
		{"v10.0", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := parseARMVersion(tt.version)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseARMVersion(%q) expected error", tt.version)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseARMVersion(%q) unexpected error: %v", tt.version, err)
			}

			if got != tt.want {
				t.Errorf("parseARMVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
