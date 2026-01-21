// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"testing"
)

func TestArchitectureString(t *testing.T) {
	tests := []struct {
		arch     Architecture
		expected string
	}{
		{ArchX86_64, "x86-64"},
		{ArchARM64, "aarch64"},
		{ArchUnknown, "unknown"},
	}

	for _, tt := range tests {
		result := tt.arch.String()
		if result != tt.expected {
			t.Errorf("Architecture.String() = %v, want %v", result, tt.expected)
		}
	}
}

func TestCompressionAlgoString(t *testing.T) {
	tests := []struct {
		algo     CompressionAlgo
		expected string
	}{
		{CompressionNone, "none"},
		{CompressionGzip, "gzip"},
		{CompressionZstd, "zstd"},
		{CompressionLZ4, "lz4"},
	}

	for _, tt := range tests {
		result := tt.algo.String()
		if result != tt.expected {
			t.Errorf("CompressionAlgo.String() = %v, want %v", result, tt.expected)
		}
	}
}

func TestParseArchSpec(t *testing.T) {
	tests := []struct {
		input    string
		wantErr  bool
		wantArch Architecture
		wantVer  ArchVersion
	}{
		{"x86-64-v1", false, ArchX86_64, X86_64_V1},
		{"x86-64-v2", false, ArchX86_64, X86_64_V2},
		{"x86-64-v3", false, ArchX86_64, X86_64_V3},
		{"x86-64-v4", false, ArchX86_64, X86_64_V4},
		{"amd64-v2", false, ArchX86_64, X86_64_V2}, // Alias test
		{"aarch64-v8.0", false, ArchARM64, ARM64_V8_0},
		{"aarch64-v9.0", false, ArchARM64, ARM64_V9_0},
		{"arm64-v8.0", false, ArchARM64, ARM64_V8_0}, // Alias test
		{"invalid", true, ArchUnknown, 0},
		{"", true, ArchUnknown, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			spec, err := ParseArchSpec(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArchSpec(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseArchSpec(%q) unexpected error: %v", tt.input, err)
				return
			}

			if spec.Architecture != tt.wantArch {
				t.Errorf("ParseArchSpec(%q).Architecture = %v, want %v",
					tt.input, spec.Architecture, tt.wantArch)
			}

			if spec.ArchVersion != tt.wantVer {
				t.Errorf("ParseArchSpec(%q).ArchVersion = %v, want %v",
					tt.input, spec.ArchVersion, tt.wantVer)
			}
		})
	}
}
