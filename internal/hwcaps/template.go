package hwcaps

import (
	"fmt"
	"os"
	"sort"
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

// PathCandidate represents a potential library path with its score
type PathCandidate struct {
	Path     string
	Score    int
	Version  format.ArchVersion
	Template string
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

// EvaluateTemplates expands templates, scores them, and returns ranked usable paths
func (e *TemplateEvaluator) EvaluateTemplates(templates []string) []string {
	var candidates []PathCandidate

	// Get all compatible versions (includes fallback)
	versions := e.getVersionFallbackChain()

	// Expand all templates and create candidates
	for templateIdx, template := range templates {
		for versionIdx, ver := range versions {
			path := e.expandTemplate(template, ver)

			// Only add if path exists
			if e.pathExists(path) {
				score := e.calculateScore(templateIdx, versionIdx, template, ver)
				candidates = append(candidates, PathCandidate{
					Path:     path,
					Score:    score,
					Version:  ver,
					Template: template,
				})
			}
		}
	}

	// Sort by score (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Extract paths in priority order
	var validPaths []string
	for _, candidate := range candidates {
		validPaths = append(validPaths, candidate.Path)
	}

	return validPaths
}

// calculateScore assigns priority score to a path candidate
// Higher score = higher priority in LD_LIBRARY_PATH
func (e *TemplateEvaluator) calculateScore(templateIdx, versionIdx int, template string, version format.ArchVersion) int {
	score := 0

	// Base score: prefer templates earlier in list (user/system priority)
	// Template 0: +1000, Template 1: +900, etc.
	score += (10 - templateIdx) * 100

	// Version match score: prefer exact version match, then close versions
	// Exact version: +100, next version: +90, etc.
	score += (10 - versionIdx) * 10

	// Path pattern bonuses
	if strings.Contains(template, "{{.ArchTriple}}-linux-gnu") {
		score += 50 // Prefer Debian multiarch (more specific)
	}
	if strings.Contains(template, "/usr/lib64") {
		score += 30 // Prefer lib64 over lib
	}
	if strings.Contains(template, "/opt/") {
		score -= 20 // Deprioritize /opt paths
	}

	// Exact version match bonus
	if version == e.version {
		score += 200
	}

	return score
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

	// Add current version (e.g., v9.4)
	versions = append(versions, e.version)

	// Add .0 variant if not already .0 (v9.4 → v9.0)
	if e.version%10 != 0 {
		baseVersion := (e.version / 10) * 10
		versions = append(versions, baseVersion)
	}

	// Add all previous versions down to v8.0
	current := e.version
	for v := current - 1; v >= format.ARM64_V8_0; v-- {
		versions = append(versions, v)
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
