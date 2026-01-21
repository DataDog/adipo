// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package hwcaps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

func TestExpandTemplate(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		arch         format.Architecture
		version      format.ArchVersion
		wantContains []string
		description  string
	}{
		{
			name:     "x86-64 template with ArchVersion",
			template: "/opt/libs/{{.ArchVersion}}",
			arch:     format.ArchX86_64,
			version:  3,
			wantContains: []string{
				"/opt/libs/x86-64-v3",
				"/opt/libs/x86-64-v2",
				"/opt/libs/x86-64-v1",
			},
			description: "Should expand to all versions up to v3",
		},
		{
			name:     "ARM64 template with separate Arch and Version",
			template: "/custom/{{.Arch}}/{{.Version}}/lib",
			arch:     format.ArchARM64,
			version:  82, // v8.2
			wantContains: []string{
				"/custom/aarch64/v8.2/lib",
				"/custom/aarch64/v8.1/lib",
				"/custom/aarch64/v8.0/lib",
			},
			description: "Should expand Arch and Version separately",
		},
		{
			name:     "x86-64 all variables",
			template: "/test/{{.Arch}}_{{.Version}}_{{.ArchVersion}}",
			arch:     format.ArchX86_64,
			version:  2,
			wantContains: []string{
				"/test/x86-64_v2_x86-64-v2",
				"/test/x86-64_v1_x86-64-v1",
			},
			description: "Should expand all three variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := expandTemplate(tt.template, tt.arch, tt.version)

			for _, want := range tt.wantContains {
				found := false
				for _, path := range paths {
					if path == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expandTemplate() missing expected path %s, got %v\nDescription: %s",
						want, paths, tt.description)
				}
			}
		})
	}
}

func TestCheckDirectoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "hwcaps-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name   string
		setup  func() string
		want   bool
		reason string
	}{
		{
			name: "existing directory",
			setup: func() string {
				return tmpDir
			},
			want:   true,
			reason: "Should return true for existing directory",
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return filepath.Join(tmpDir, "does-not-exist")
			},
			want:   false,
			reason: "Should return false for non-existent directory",
		},
		{
			name: "file not directory",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "testfile")
				_ = os.WriteFile(filePath, []byte("test"), 0644)
				return filePath
			},
			want:   false,
			reason: "Should return false for file (not directory)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			got := checkDirectoryExists(path)
			if got != tt.want {
				t.Errorf("checkDirectoryExists(%s) = %v, want %v\nReason: %s",
					path, got, tt.want, tt.reason)
			}
		})
	}
}

func TestParseArchVersionFromPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantArch    format.Architecture
		wantVersion format.ArchVersion
		wantErr     bool
		description string
	}{
		{
			name:        "x86-64-v3 in path",
			path:        "/usr/lib64/glibc-hwcaps/x86-64-v3",
			wantArch:    format.ArchX86_64,
			wantVersion: format.X86_64_V3,
			wantErr:     false,
			description: "Should parse x86-64-v3 correctly",
		},
		{
			name:        "aarch64-v8.1 in path",
			path:        "/usr/lib64/glibc-hwcaps/aarch64-v8.1",
			wantArch:    format.ArchARM64,
			wantVersion: format.ARM64_V8_1,
			wantErr:     false,
			description: "Should parse aarch64-v8.1 correctly",
		},
		{
			name:        "aarch64-v9.0 in path",
			path:        "/custom/libs/aarch64-v9.0",
			wantArch:    format.ArchARM64,
			wantVersion: format.ARM64_V9_0,
			wantErr:     false,
			description: "Should parse aarch64-v9.0 correctly",
		},
		{
			name:        "no version in path",
			path:        "/opt/x86-64/lib",
			wantArch:    format.ArchX86_64,
			wantVersion: 0,
			wantErr:     true,
			description: "Should error when no version found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArch, gotVersion, err := parseArchVersionFromPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArchVersionFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotArch != tt.wantArch {
					t.Errorf("parseArchVersionFromPath() arch = %v, want %v", gotArch, tt.wantArch)
				}
				if gotVersion != tt.wantVersion {
					t.Errorf("parseArchVersionFromPath() version = %v, want %v", gotVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestCheckCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		caps        *cpu.Capabilities
		arch        format.Architecture
		version     format.ArchVersion
		want        bool
		description string
	}{
		{
			name: "compatible same version",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
			},
			arch:        format.ArchX86_64,
			version:     format.X86_64_V3,
			want:        true,
			description: "Should be compatible with same version",
		},
		{
			name: "compatible higher CPU version",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V4,
			},
			arch:        format.ArchX86_64,
			version:     format.X86_64_V2,
			want:        true,
			description: "Should be compatible when CPU version is higher",
		},
		{
			name: "incompatible lower CPU version",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V2,
			},
			arch:        format.ArchX86_64,
			version:     format.X86_64_V3,
			want:        false,
			description: "Should be incompatible when CPU version is lower",
		},
		{
			name: "incompatible different architecture",
			caps: &cpu.Capabilities{
				ArchType: format.ArchX86_64,
				Version:  format.X86_64_V3,
			},
			arch:        format.ArchARM64,
			version:     format.ARM64_V8_0,
			want:        false,
			description: "Should be incompatible with different architecture",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkCompatibility(tt.caps, tt.arch, tt.version)
			if got != tt.want {
				t.Errorf("checkCompatibility() = %v, want %v\nDescription: %s",
					got, tt.want, tt.description)
			}
		})
	}
}

func TestSelectCompatiblePaths(t *testing.T) {
	results := []ScanResult{
		{Path: "/path1", Exists: true, IsCompatible: true, Priority: 100},
		{Path: "/path2", Exists: false, IsCompatible: true, Priority: 90},
		{Path: "/path3", Exists: true, IsCompatible: false, Priority: 80},
		{Path: "/path4", Exists: true, IsCompatible: true, Priority: 110},
		{Path: "/path5", Exists: true, IsCompatible: true, Priority: 95},
	}

	selected := SelectCompatiblePaths(results)

	// Should only include existing and compatible
	if len(selected) != 3 {
		t.Errorf("SelectCompatiblePaths() returned %d results, want 3", len(selected))
	}

	// Should be sorted by priority (highest first)
	if len(selected) >= 3 {
		if selected[0].Path != "/path4" {
			t.Errorf("First selected path = %s, want /path4", selected[0].Path)
		}
		if selected[1].Path != "/path1" {
			t.Errorf("Second selected path = %s, want /path1", selected[1].Path)
		}
		if selected[2].Path != "/path5" {
			t.Errorf("Third selected path = %s, want /path5", selected[2].Path)
		}
	}
}

func TestBuildLibraryPath(t *testing.T) {
	tests := []struct {
		name     string
		selected []ScanResult
		want     string
	}{
		{
			name: "multiple paths",
			selected: []ScanResult{
				{Path: "/path1"},
				{Path: "/path2"},
				{Path: "/path3"},
			},
			want: "/path1:/path2:/path3",
		},
		{
			name: "single path",
			selected: []ScanResult{
				{Path: "/only/path"},
			},
			want: "/only/path",
		},
		{
			name:     "empty selection",
			selected: []ScanResult{},
			want:     "",
		},
		{
			name: "duplicate paths removed",
			selected: []ScanResult{
				{Path: "/path1"},
				{Path: "/path2"},
				{Path: "/path1"},
			},
			want: "/path1:/path2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLibraryPath(tt.selected)
			if got != tt.want {
				t.Errorf("BuildLibraryPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanDirectories(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "hwcaps-scan-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create some test directories
	testDir1 := filepath.Join(tmpDir, "test-v3")
	testDir2 := filepath.Join(tmpDir, "test-v2")
	if err := os.MkdirAll(testDir1, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	if err := os.MkdirAll(testDir2, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	caps := &cpu.Capabilities{
		ArchType: format.ArchX86_64,
		Version:  format.X86_64_V3,
	}

	config := &ScanConfig{
		Capabilities:          caps,
		Templates:             []string{},
		ScanDirs:              []string{testDir1, testDir2},
		IncludeStandardHwcaps: false,
		IncludeOptPattern:     false,
	}

	results := ScanDirectories(config)

	// Should find both user directories
	if len(results) != 2 {
		t.Errorf("ScanDirectories() found %d results, want 2", len(results))
	}

	// Both should exist
	for _, result := range results {
		if !result.Exists {
			t.Errorf("Expected directory %s to exist", result.Path)
		}
	}
}

func TestGetAllVersions(t *testing.T) {
	tests := []struct {
		name       string
		arch       format.Architecture
		maxVersion format.ArchVersion
		wantCount  int
		wantFirst  format.ArchVersion
		wantLast   format.ArchVersion
	}{
		{
			name:       "x86-64 v3",
			arch:       format.ArchX86_64,
			maxVersion: format.X86_64_V3,
			wantCount:  3,
			wantFirst:  format.X86_64_V3, // Reversed, so highest first
			wantLast:   format.X86_64_V1,
		},
		{
			name:       "x86-64 v4",
			arch:       format.ArchX86_64,
			maxVersion: format.X86_64_V4,
			wantCount:  4,
			wantFirst:  format.X86_64_V4,
			wantLast:   format.X86_64_V1,
		},
		{
			name:       "ARM64 v8.2",
			arch:       format.ArchARM64,
			maxVersion: format.ARM64_V8_2,
			wantCount:  3, // v8.0, v8.1, v8.2
			wantFirst:  format.ARM64_V8_2,
			wantLast:   format.ARM64_V8_0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := getAllVersions(tt.arch, tt.maxVersion)

			if len(versions) != tt.wantCount {
				t.Errorf("getAllVersions() returned %d versions, want %d", len(versions), tt.wantCount)
			}

			if len(versions) > 0 {
				if versions[0] != tt.wantFirst {
					t.Errorf("First version = %v, want %v", versions[0], tt.wantFirst)
				}
				if versions[len(versions)-1] != tt.wantLast {
					t.Errorf("Last version = %v, want %v", versions[len(versions)-1], tt.wantLast)
				}
			}
		})
	}
}

func TestGetArchString(t *testing.T) {
	tests := []struct {
		arch    format.Architecture
		version format.ArchVersion
		want    string
	}{
		{format.ArchX86_64, format.X86_64_V3, "x86-64-v3"},
		{format.ArchX86_64, format.X86_64_V1, "x86-64-v1"},
		{format.ArchARM64, format.ARM64_V8_0, "aarch64-v8.0"},
		{format.ArchARM64, format.ARM64_V9_0, "aarch64-v9.0"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := getArchString(tt.arch, tt.version)
			if got != tt.want {
				t.Errorf("getArchString(%v, %v) = %v, want %v",
					tt.arch, tt.version, got, tt.want)
			}
		})
	}
}

func TestPathPriority(t *testing.T) {
	caps := &cpu.Capabilities{
		ArchType: format.ArchX86_64,
		Version:  format.X86_64_V3,
	}

	// Create config with all scanning modes enabled
	config := &ScanConfig{
		Capabilities:          caps,
		Templates:             []string{"/template/{{.ArchVersion}}"},
		ScanDirs:              []string{"/user/dir"},
		IncludeStandardHwcaps: true,
		IncludeOptPattern:     true,
	}

	results := ScanDirectories(config)

	// Check that standard hwcaps paths have highest priority
	var standardPriority, templatePriority, optPriority, userPriority int
	for _, result := range results {
		if result.Source == SourceStandardHwcaps && strings.Contains(result.Path, "x86-64-v3") {
			standardPriority = result.Priority
		}
		if result.Source == SourceTemplate && strings.Contains(result.Path, "x86-64-v3") {
			templatePriority = result.Priority
		}
		if result.Source == SourceOptPattern {
			optPriority = result.Priority
		}
		if result.Source == SourceUserDir {
			userPriority = result.Priority
		}
	}

	// Verify priority ordering (for same version)
	if standardPriority <= templatePriority {
		t.Errorf("Standard hwcaps priority (%d) should be higher than template (%d)",
			standardPriority, templatePriority)
	}
	if templatePriority <= optPriority {
		t.Errorf("Template priority (%d) should be higher than opt pattern (%d)",
			templatePriority, optPriority)
	}
	if optPriority <= userPriority {
		t.Errorf("Opt pattern priority (%d) should be higher than user dir (%d)",
			optPriority, userPriority)
	}
}
