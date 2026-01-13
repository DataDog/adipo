// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package hwcaps

import (
	"fmt"
	"os"
	"strings"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

// TemplateEvaluator evaluates path templates at runtime
type TemplateEvaluator struct {
	arch    format.Architecture
	version format.ArchVersion
	caps    *cpu.Capabilities
}

// NewTemplateEvaluator creates evaluator for current CPU
func NewTemplateEvaluator(arch format.Architecture, version format.ArchVersion) (*TemplateEvaluator, error) {
	caps, err := cpu.Detect()
	if err != nil {
		return nil, err
	}

	return &TemplateEvaluator{
		arch:    arch,
		version: version,
		caps:    caps,
	}, nil
}

// EvaluateTemplates expands templates and returns existing paths in priority order
func (e *TemplateEvaluator) EvaluateTemplates(templates []string) []string {
	var validPaths []string
	seen := make(map[string]bool)

	// Get all compatible versions (includes fallback)
	versions := e.getVersionFallbackChain()

	// Expand templates in order and collect existing paths
	// Priority: newer versions first, then template order within each version
	for _, ver := range versions {
		for _, template := range templates {
			path := e.expandTemplate(template, ver)

			// Only add if path exists and not already seen
			if !seen[path] && e.pathExists(path) {
				validPaths = append(validPaths, path)
				seen[path] = true
			}
		}
	}

	return validPaths
}

// getVersionFallbackChain returns version chain with fallbacks
// Example: v9.4 → [v9.4, v9.0, v9, v8.9, v8.8, ..., v8.0]
func (e *TemplateEvaluator) getVersionFallbackChain() []format.ArchVersion {
	var versions []format.ArchVersion

	switch e.arch {
	case format.ArchX86_64:
		// x86-64: v3 → [v3, v2, v1]
		current := e.version
		for v := current; v >= format.X86_64_V1; v-- {
			versions = append(versions, v)
		}

	case format.ArchARM64:
		// ARM64: v9.4 → [v9.4, v9.0, v9, v8.9, ..., v8.0]
		versions = e.getARMFallbackChain()
	}

	return versions
}

// getARMFallbackChain handles ARM64 version fallback logic
func (e *TemplateEvaluator) getARMFallbackChain() []format.ArchVersion {
	var versions []format.ArchVersion

	// Use the canonical ARM64 version ordering from the format package
	// Find current version in the list and include it plus all older versions
	foundCurrent := false
	for _, v := range format.ARM64VersionFallbackOrder {
		if v == e.version {
			foundCurrent = true
		}
		if foundCurrent {
			versions = append(versions, v)
		}
	}

	return versions
}

// expandTemplate replaces variables with values
func (e *TemplateEvaluator) expandTemplate(template string, version format.ArchVersion) string {
	replacements := map[string]string{
		"{{.Arch}}":        e.getArchName(),                  // "x86-64", "aarch64"
		"{{.ArchTriple}}":  e.getArchTriple(),                // "x86_64", "aarch64"
		"{{.Version}}":     e.getVersionStr(version),         // "v3", "v8.2"
		"{{.ArchVersion}}": e.getArchVersionStr(version), // "x86-64-v3", "aarch64-v8.2"
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// Helper functions

func (e *TemplateEvaluator) getArchName() string {
	switch e.arch {
	case format.ArchX86_64:
		return "x86-64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

func (e *TemplateEvaluator) getArchTriple() string {
	switch e.arch {
	case format.ArchX86_64:
		return "x86_64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

func (e *TemplateEvaluator) getVersionStr(version format.ArchVersion) string {
	switch e.arch {
	case format.ArchX86_64:
		return fmt.Sprintf("v%d", version)
	case format.ArchARM64:
		return version.String(e.arch) // "v8.2", "v9.0"
	default:
		return "unknown"
	}
}

func (e *TemplateEvaluator) getArchVersionStr(version format.ArchVersion) string {
	return fmt.Sprintf("%s-%s", e.getArchName(), e.getVersionStr(version))
}

func (e *TemplateEvaluator) pathExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
