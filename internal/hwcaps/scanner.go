package hwcaps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

// ScanConfig configures directory scanning
type ScanConfig struct {
	Capabilities          *cpu.Capabilities
	Templates             []string
	ScanDirs              []string
	IncludeStandardHwcaps bool
	IncludeOptPattern     bool
}

// ScanResult represents a discovered library directory
type ScanResult struct {
	Path         string
	Architecture format.Architecture
	Version      format.ArchVersion
	Exists       bool
	IsCompatible bool
	Priority     int // Higher = better (based on version)
	Source       string
}

// Source types for priority ordering
const (
	SourceStandardHwcaps = "standard-hwcaps"
	SourceTemplate       = "template"
	SourceOptPattern     = "opt-pattern"
	SourceUserDir        = "user-dir"
)

// ScanDirectories scans for library directories based on configuration
func ScanDirectories(config *ScanConfig) []ScanResult {
	var results []ScanResult
	arch := config.Capabilities.ArchType
	version := config.Capabilities.Version

	// Collect all candidate paths
	var candidates []struct {
		path   string
		source string
	}

	// 1. Standard glibc-hwcaps paths
	if config.IncludeStandardHwcaps {
		standardPaths := generateStandardHwcapsPaths(arch, version)
		for _, path := range standardPaths {
			candidates = append(candidates, struct {
				path   string
				source string
			}{path, SourceStandardHwcaps})
		}
	}

	// 2. Opt pattern paths
	if config.IncludeOptPattern {
		optPaths := generateOptPatternPaths(arch)
		for _, path := range optPaths {
			candidates = append(candidates, struct {
				path   string
				source string
			}{path, SourceOptPattern})
		}
	}

	// 3. Template paths
	for _, template := range config.Templates {
		templatePaths := expandTemplate(template, arch, version)
		for _, path := range templatePaths {
			candidates = append(candidates, struct {
				path   string
				source string
			}{path, SourceTemplate})
		}
	}

	// 4. User-specified directories
	for _, dir := range config.ScanDirs {
		candidates = append(candidates, struct {
			path   string
			source string
		}{dir, SourceUserDir})
	}

	// Process each candidate
	for _, candidate := range candidates {
		result := ScanResult{
			Path:   candidate.path,
			Source: candidate.source,
		}

		// Check if directory exists
		result.Exists = checkDirectoryExists(candidate.path)

		// Parse architecture and version from path
		pathArch, pathVersion, err := parseArchVersionFromPath(candidate.path)
		if err != nil {
			// For user directories and opt patterns, use current CPU capabilities
			if candidate.source == SourceUserDir || candidate.source == SourceOptPattern {
				pathArch = arch
				pathVersion = version
			} else {
				// Skip if we can't parse
				continue
			}
		}

		result.Architecture = pathArch
		result.Version = pathVersion

		// Check compatibility
		result.IsCompatible = checkCompatibility(config.Capabilities, pathArch, pathVersion)

		// Assign priority (higher version = higher priority)
		result.Priority = int(pathVersion)

		// Boost priority for standard hwcaps paths
		if candidate.source == SourceStandardHwcaps {
			result.Priority += 10000
		} else if candidate.source == SourceTemplate {
			result.Priority += 1000
		} else if candidate.source == SourceOptPattern {
			result.Priority += 100
		}

		results = append(results, result)
	}

	return results
}

// generateStandardHwcapsPaths generates standard glibc-hwcaps paths
func generateStandardHwcapsPaths(arch format.Architecture, version format.ArchVersion) []string {
	var paths []string

	baseDir64 := "/usr/lib64/glibc-hwcaps"
	baseDir := "/usr/lib/glibc-hwcaps"

	versions := getAllVersions(arch, version)
	for _, v := range versions {
		archStr := getArchString(arch, v)
		paths = append(paths, filepath.Join(baseDir64, archStr))
		paths = append(paths, filepath.Join(baseDir, archStr))
	}

	return paths
}

// generateOptPatternPaths generates /opt/<arch>/lib paths
func generateOptPatternPaths(arch format.Architecture) []string {
	var paths []string

	baseArch := getBaseArchString(arch)
	paths = append(paths, filepath.Join("/opt", baseArch, "lib"))

	return paths
}

// expandTemplate expands a template with variables
func expandTemplate(template string, arch format.Architecture, version format.ArchVersion) []string {
	var paths []string

	baseArch := getBaseArchString(arch)
	archTriple := getArchTripleString(arch)
	versions := getAllVersions(arch, version)

	for _, v := range versions {
		versionStr := v.String(arch)
		archVersion := getArchString(arch, v)

		path := template
		path = strings.ReplaceAll(path, "{{.ArchTriple}}", archTriple)
		path = strings.ReplaceAll(path, "{{.Arch}}", baseArch)
		path = strings.ReplaceAll(path, "{{.Version}}", versionStr)
		path = strings.ReplaceAll(path, "{{.ArchVersion}}", archVersion)

		paths = append(paths, path)
	}

	return paths
}

// checkDirectoryExists checks if a directory exists
func checkDirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// parseArchVersionFromPath attempts to parse architecture and version from path
func parseArchVersionFromPath(path string) (format.Architecture, format.ArchVersion, error) {
	// Extract the last component of the path
	base := filepath.Base(path)

	// Try to parse as arch-version format (e.g., "x86-64-v3", "aarch64-v8.1")
	spec, err := format.ParseArchSpec(base)
	if err == nil {
		return spec.Architecture, spec.ArchVersion, nil
	}

	// Check parent directory for architecture hints
	parent := filepath.Dir(path)
	parentBase := filepath.Base(parent)

	if strings.Contains(parentBase, "x86-64") || strings.Contains(parentBase, "x86_64") {
		return format.ArchX86_64, 0, fmt.Errorf("architecture found but no version")
	}

	if strings.Contains(parentBase, "aarch64") || strings.Contains(parentBase, "arm64") {
		return format.ArchARM64, 0, fmt.Errorf("architecture found but no version")
	}

	return format.ArchUnknown, 0, fmt.Errorf("could not parse architecture from path: %s", path)
}

// checkCompatibility checks if CPU is compatible with the given architecture/version
func checkCompatibility(caps *cpu.Capabilities, arch format.Architecture, version format.ArchVersion) bool {
	if caps.ArchType != arch {
		return false
	}

	return caps.Version >= version
}

// getAllVersions returns all versions from v1/v8.0 up to the current version
func getAllVersions(arch format.Architecture, maxVersion format.ArchVersion) []format.ArchVersion {
	var versions []format.ArchVersion

	switch arch {
	case format.ArchX86_64:
		// x86-64: v1, v2, v3, v4
		allX86Versions := []format.ArchVersion{
			format.X86_64_V1,
			format.X86_64_V2,
			format.X86_64_V3,
			format.X86_64_V4,
		}
		for _, v := range allX86Versions {
			if v <= maxVersion {
				versions = append(versions, v)
			}
		}
	case format.ArchARM64:
		// ARM64: all versions in order
		allARMVersions := []format.ArchVersion{
			format.ARM64_V8_0,
			format.ARM64_V8_1,
			format.ARM64_V8_2,
			format.ARM64_V8_3,
			format.ARM64_V8_4,
			format.ARM64_V8_5,
			format.ARM64_V8_6,
			format.ARM64_V8_7,
			format.ARM64_V8_8,
			format.ARM64_V8_9,
			format.ARM64_V9_0,
			format.ARM64_V9_1,
			format.ARM64_V9_2,
			format.ARM64_V9_3,
			format.ARM64_V9_4,
			format.ARM64_V9_5,
		}
		for _, v := range allARMVersions {
			if v <= maxVersion {
				versions = append(versions, v)
			}
		}
	}

	// Reverse to get highest versions first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return versions
}

// getBaseArchString returns the base architecture string (x86-64 or aarch64)
func getBaseArchString(arch format.Architecture) string {
	switch arch {
	case format.ArchX86_64:
		return "x86-64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

// getArchTripleString returns the architecture triple (for Debian multiarch)
func getArchTripleString(arch format.Architecture) string {
	switch arch {
	case format.ArchX86_64:
		return "x86_64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

// getArchString returns the full architecture-version string (e.g., x86-64-v3)
func getArchString(arch format.Architecture, version format.ArchVersion) string {
	base := getBaseArchString(arch)
	versionStr := version.String(arch)
	return base + "-" + versionStr
}
