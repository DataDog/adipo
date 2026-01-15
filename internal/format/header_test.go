// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package format

import (
	"bytes"
	"strings"
	"testing"
)

func TestArchVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version ArchVersion
		arch    Architecture
		want    string
	}{
		{"x86-64 v1", X86_64_V1, ArchX86_64, "v1"},
		{"x86-64 v2", X86_64_V2, ArchX86_64, "v2"},
		{"x86-64 v3", X86_64_V3, ArchX86_64, "v3"},
		{"x86-64 v4", X86_64_V4, ArchX86_64, "v4"},
		{"ARM64 v8.0", ARM64_V8_0, ArchARM64, "v8.0"},
		{"ARM64 v8", ARM64_V8, ArchARM64, "v8"},
		{"ARM64 v8.1", ARM64_V8_1, ArchARM64, "v8.1"},
		{"ARM64 v9.0", ARM64_V9_0, ArchARM64, "v9.0"},
		{"ARM64 v9", ARM64_V9, ArchARM64, "v9"},
		{"unknown version", ArchVersion(99), ArchX86_64, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(tt.arch); got != tt.want {
				t.Errorf("ArchVersion.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryFormatString(t *testing.T) {
	tests := []struct {
		format BinaryFormat
		want   string
	}{
		{FormatUnknown, "unknown"},
		{FormatELF, "ELF"},
		{FormatMachO, "Mach-O"},
		{BinaryFormat(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("BinaryFormat.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatHeaderMarshalUnmarshal(t *testing.T) {
	original := &FormatHeader{
		Magic:            MagicMarker,
		Version:          FormatVersion,
		NumBinaries:      2,
		StubSize:         1024,
		StubArchitecture: ArchX86_64,
		StubArchVersion:  X86_64_V3,
		Reserved:         [924]byte{},
	}

	// Set stub settings in Reserved space
	original.SetStubSettings(StubSettingVerbose | StubSettingCleanupOnExit)
	if err := original.SetDefaultExtractDir("/tmp/test"); err != nil {
		t.Fatalf("SetDefaultExtractDir() error: %v", err)
	}

	// Marshal
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error: %v", err)
	}

	if len(data) != HeaderSize {
		t.Errorf("MarshalBinary() produced %d bytes, want %d", len(data), HeaderSize)
	}

	// Unmarshal
	decoded := &FormatHeader{}
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error: %v", err)
	}

	// Verify fields
	if decoded.Magic != original.Magic {
		t.Errorf("Magic = %v, want %v", decoded.Magic, original.Magic)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version = %d, want %d", decoded.Version, original.Version)
	}
	if decoded.NumBinaries != original.NumBinaries {
		t.Errorf("NumBinaries = %d, want %d", decoded.NumBinaries, original.NumBinaries)
	}
	if decoded.StubSize != original.StubSize {
		t.Errorf("StubSize = %d, want %d", decoded.StubSize, original.StubSize)
	}
	if decoded.StubArchitecture != original.StubArchitecture {
		t.Errorf("StubArchitecture = %v, want %v", decoded.StubArchitecture, original.StubArchitecture)
	}
	if decoded.StubArchVersion != original.StubArchVersion {
		t.Errorf("StubArchVersion = %v, want %v", decoded.StubArchVersion, original.StubArchVersion)
	}

	// Verify stub settings and extract dir
	if decoded.GetStubSettings() != original.GetStubSettings() {
		t.Errorf("StubSettings = %d, want %d", decoded.GetStubSettings(), original.GetStubSettings())
	}
	if decoded.GetDefaultExtractDir() != "/tmp/test" {
		t.Errorf("DefaultExtractDir = %q, want %q", decoded.GetDefaultExtractDir(), "/tmp/test")
	}
}

func TestFormatHeaderUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"too short", make([]byte, HeaderSize-1)},
		{"empty", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &FormatHeader{}
			err := header.UnmarshalBinary(tt.data)
			if err == nil {
				t.Error("UnmarshalBinary() expected error for invalid data")
			}
		})
	}
}

func TestStubSettings(t *testing.T) {
	header := &FormatHeader{}

	// Set settings
	settings := StubSettingVerbose | StubSettingCleanupOnExit
	header.SetStubSettings(settings)

	// Get settings
	got := header.GetStubSettings()
	if got != settings {
		t.Errorf("GetStubSettings() = %d, want %d", got, settings)
	}
}

func TestDefaultExtractDir(t *testing.T) {
	header := &FormatHeader{}

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{"short path", "/tmp", false},
		{"max length path", strings.Repeat("a", 511), false},
		{"too long path", strings.Repeat("a", 512), true},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := header.SetDefaultExtractDir(tt.dir)

			if tt.wantErr {
				if err == nil {
					t.Error("SetDefaultExtractDir() expected error for too long path")
				}
				return
			}

			if err != nil {
				t.Errorf("SetDefaultExtractDir() unexpected error: %v", err)
				return
			}

			got := header.GetDefaultExtractDir()
			if got != tt.dir {
				t.Errorf("GetDefaultExtractDir() = %q, want %q", got, tt.dir)
			}
		})
	}
}

func TestBinaryMetadataMarshalUnmarshal(t *testing.T) {
	original := &BinaryMetadata{
		Architecture:     ArchX86_64,
		ArchVersion:      X86_64_V3,
		Format:           FormatELF,
		Compression:      CompressionZstd,
		RequiredFeatures: 0x12345678,
		Priority:         10,
		OriginalSize:     1024 * 1024,
		CompressedSize:   512 * 1024,
		Checksum:         [32]byte{1, 2, 3, 4, 5},
	}

	// Marshal
	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error: %v", err)
	}

	if len(data) != MetadataEntrySize {
		t.Errorf("MarshalBinary() produced %d bytes, want %d", len(data), MetadataEntrySize)
	}

	// Unmarshal
	decoded := &BinaryMetadata{}
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error: %v", err)
	}

	// Verify all fields
	if decoded.Architecture != original.Architecture {
		t.Errorf("Architecture = %v, want %v", decoded.Architecture, original.Architecture)
	}
	if decoded.ArchVersion != original.ArchVersion {
		t.Errorf("ArchVersion = %v, want %v", decoded.ArchVersion, original.ArchVersion)
	}
	if decoded.Format != original.Format {
		t.Errorf("Format = %v, want %v", decoded.Format, original.Format)
	}
	if decoded.Compression != original.Compression {
		t.Errorf("Compression = %v, want %v", decoded.Compression, original.Compression)
	}
	if decoded.RequiredFeatures != original.RequiredFeatures {
		t.Errorf("RequiredFeatures = 0x%x, want 0x%x", decoded.RequiredFeatures, original.RequiredFeatures)
	}
	if decoded.Priority != original.Priority {
		t.Errorf("Priority = %d, want %d", decoded.Priority, original.Priority)
	}
	if decoded.OriginalSize != original.OriginalSize {
		t.Errorf("OriginalSize = %d, want %d", decoded.OriginalSize, original.OriginalSize)
	}
	if decoded.CompressedSize != original.CompressedSize {
		t.Errorf("CompressedSize = %d, want %d", decoded.CompressedSize, original.CompressedSize)
	}
	if !bytes.Equal(decoded.Checksum[:], original.Checksum[:]) {
		t.Errorf("Checksum mismatch")
	}
}

func TestBinaryMetadataUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"too short", make([]byte, MetadataEntrySize-1)},
		{"empty", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := &BinaryMetadata{}
			err := meta.UnmarshalBinary(tt.data)
			if err == nil {
				t.Error("UnmarshalBinary() expected error for invalid data")
			}
		})
	}
}

func TestLibraryPathTemplates(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
		wantErr   bool
	}{
		{
			name:      "empty templates",
			templates: []string{},
			wantErr:   true, // SetLibraryPathTemplates rejects empty arrays
		},
		{
			name:      "single template",
			templates: []string{"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}"},
			wantErr:   false,
		},
		{
			name: "multiple templates",
			templates: []string{
				"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
				"/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
				"/opt/{{.Arch}}/lib",
			},
			wantErr: false,
		},
		{
			name: "templates that exceed available space",
			templates: []string{
				strings.Repeat("a", 200),
				strings.Repeat("b", 200),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &BinaryMetadata{
				Architecture: ArchX86_64,
				ArchVersion:  X86_64_V3,
			}

			err := metadata.SetLibraryPathTemplates(tt.templates)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLibraryPathTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := metadata.GetLibraryPathTemplates()
				if len(got) != len(tt.templates) {
					t.Errorf("GetLibraryPathTemplates() length = %v, want %v", len(got), len(tt.templates))
					return
				}
				for i, template := range tt.templates {
					if got[i] != template {
						t.Errorf("GetLibraryPathTemplates()[%d] = %v, want %v", i, got[i], template)
					}
				}
			}
		})
	}
}

func TestLibraryPathTemplatesMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
	}{
		{
			name: "single template",
			templates: []string{
				"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
			},
		},
		{
			name: "multiple templates",
			templates: []string{
				"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",
				"/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",
				"/opt/{{.Arch}}/lib",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original metadata
			original := &BinaryMetadata{
				Architecture: ArchX86_64,
				ArchVersion:  X86_64_V3,
			}

			// Set library path templates
			if err := original.SetLibraryPathTemplates(tt.templates); err != nil {
				t.Fatalf("SetLibraryPathTemplates() error = %v", err)
			}

			// Marshal
			data, err := original.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary() error = %v", err)
			}

			// Unmarshal
			restored := &BinaryMetadata{}
			err = restored.UnmarshalBinary(data)
			if err != nil {
				t.Fatalf("UnmarshalBinary() error = %v", err)
			}

			// Verify templates are preserved
			gotTemplates := restored.GetLibraryPathTemplates()
			if len(gotTemplates) != len(tt.templates) {
				t.Errorf("Templates not preserved: got %d templates, want %d",
					len(gotTemplates), len(tt.templates))
				return
			}

			for i, template := range tt.templates {
				if gotTemplates[i] != template {
					t.Errorf("Template[%d] not preserved: got %v, want %v",
						i, gotTemplates[i], template)
				}
			}

			// Verify metadata version is set correctly
			if restored.MetadataVersion != MetadataVersionV1 {
				t.Errorf("MetadataVersion = %v, want %v",
					restored.MetadataVersion, MetadataVersionV1)
			}
		})
	}
}
