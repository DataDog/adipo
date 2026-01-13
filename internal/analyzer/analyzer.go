// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"fmt"
	"os"
	"strings"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/elf"
	"github.com/DataDog/adipo/internal/features"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/macho"
)

// Config holds configuration for the analyzer
type Config struct {
	ObjdumpPath       string // Path to objdump binary (empty = auto-detect)
	CollectStatistics bool   // Whether to collect instruction statistics
	MaxInstructions   int    // Maximum instructions to analyze (0 = unlimited)
}

// Analyzer analyzes binaries to detect CPU feature usage
type Analyzer struct {
	config       Config
	disassembler *Disassembler
}

// AnalysisResult contains the results of binary analysis
type AnalysisResult struct {
	Architecture     format.Architecture
	DetectedVersion  format.ArchVersion
	DetectedFeatures uint64
	FeatureNames     []string
	Statistics       *InstructionStatistics

	// Compatibility with current CPU
	CurrentCPU       *cpu.Capabilities
	CanRun           bool
	CompatibilityMsg string
}

// New creates a new analyzer with the given configuration
func New(config Config) (*Analyzer, error) {
	// Default max instructions if not specified
	if config.MaxInstructions == 0 {
		config.MaxInstructions = 100000 // Reasonable default to avoid analyzing huge binaries
	}

	// Create disassembler
	disassembler, err := NewDisassembler(config.ObjdumpPath)
	if err != nil {
		return nil, err
	}

	return &Analyzer{
		config:       config,
		disassembler: disassembler,
	}, nil
}

// Analyze analyzes a binary and returns detected features
func (a *Analyzer) Analyze(binaryPath string) (*AnalysisResult, error) {
	// 1. Detect architecture from binary format
	arch, err := detectArchitecture(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect architecture: %w", err)
	}

	// 2. Disassemble binary
	instructions, err := a.disassembler.Disassemble(binaryPath, a.config.MaxInstructions)
	if err != nil {
		return nil, fmt.Errorf("failed to disassemble: %w", err)
	}

	if len(instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in binary")
	}

	// 3. Map instructions to features
	var detectedFeatures uint64
	var stats *InstructionStatistics
	if a.config.CollectStatistics {
		stats = NewInstructionStatistics()
	}

	for _, insn := range instructions {
		var insnFeatures uint64

		switch arch {
		case format.ArchX86_64:
			insnFeatures = MapX86InstructionToFeatures(insn)
		case format.ArchARM64:
			insnFeatures = MapARMInstructionToFeatures(insn)
		}

		detectedFeatures |= insnFeatures

		if stats != nil {
			stats.RecordInstruction(insnFeatures, arch)
		}
	}

	// 4. Determine microarchitecture level from features
	version := determineVersionFromFeatures(arch, detectedFeatures)

	// 5. Get feature names
	featureNames := formatFeatureNames(arch, detectedFeatures)

	// 6. Create result
	result := &AnalysisResult{
		Architecture:     arch,
		DetectedVersion:  version,
		DetectedFeatures: detectedFeatures,
		FeatureNames:     featureNames,
		Statistics:       stats,
	}

	// 7. Assess compatibility with current CPU
	result.AssessCompatibility()

	return result, nil
}

// detectArchitecture detects the CPU architecture of a binary
func detectArchitecture(binaryPath string) (format.Architecture, error) {
	// Try to detect format
	binaryFormat, err := format.DetectFormat(binaryPath)
	if err != nil {
		return format.ArchUnknown, err
	}

	// Analyze based on format
	switch binaryFormat {
	case format.FormatELF:
		info, err := elf.Analyze(binaryPath)
		if err != nil {
			return format.ArchUnknown, fmt.Errorf("failed to analyze ELF: %w", err)
		}
		return info.Architecture, nil

	case format.FormatMachO:
		info, err := macho.Analyze(binaryPath)
		if err != nil {
			return format.ArchUnknown, fmt.Errorf("failed to analyze Mach-O: %w", err)
		}
		return info.Architecture, nil

	default:
		return format.ArchUnknown, fmt.Errorf("unsupported binary format: %s", binaryFormat)
	}
}

// determineVersionFromFeatures determines the microarchitecture level from detected features
func determineVersionFromFeatures(arch format.Architecture, detectedFeatures uint64) format.ArchVersion {
	switch arch {
	case format.ArchX86_64:
		// Check from highest to lowest
		if (detectedFeatures & features.X86_64_V4_Features) == features.X86_64_V4_Features {
			return format.X86_64_V4
		}
		if (detectedFeatures & features.X86_64_V3_Features) == features.X86_64_V3_Features {
			return format.X86_64_V3
		}
		if (detectedFeatures & features.X86_64_V2_Features) == features.X86_64_V2_Features {
			return format.X86_64_V2
		}
		return format.X86_64_V1

	case format.ArchARM64:
		// ARM version detection is more complex, check key features
		// Note: This is a simplified heuristic based on key features
		if detectedFeatures&features.ARM_SVE2 != 0 {
			return format.ARM64_V9_0 // SVE2 implies v9.0+
		}
		if detectedFeatures&features.ARM_SVE != 0 {
			return format.ARM64_V8_2 // SVE implies v8.2+
		}
		if detectedFeatures&features.ARM_ATOMICS != 0 {
			return format.ARM64_V8_1 // LSE implies v8.1+
		}
		return format.ARM64_V8_0
	}

	return 0 // Unknown
}

// formatFeatureNames converts a feature bitmask to a list of feature names
func formatFeatureNames(arch format.Architecture, detectedFeatures uint64) []string {
	switch arch {
	case format.ArchX86_64:
		return features.FormatX86Features(detectedFeatures)
	case format.ArchARM64:
		return features.FormatARMFeatures(detectedFeatures)
	default:
		return []string{}
	}
}

// AssessCompatibility checks if the analyzed binary can run on the current CPU
func (r *AnalysisResult) AssessCompatibility() {
	caps, err := cpu.Detect()
	if err != nil {
		r.CompatibilityMsg = "Unable to detect current CPU capabilities"
		r.CanRun = false
		return
	}

	r.CurrentCPU = caps
	r.CanRun = caps.IsCompatibleWith(
		r.Architecture,
		r.DetectedVersion,
		r.DetectedFeatures,
	)

	if r.CanRun {
		r.CompatibilityMsg = fmt.Sprintf(
			"Can likely run on current CPU (%s %s)",
			caps.Architecture,
			caps.VersionStr,
		)
	} else {
		// Find missing features
		missing := r.DetectedFeatures & ^caps.FeatureMask
		if missing != 0 {
			missingNames := formatFeatureNames(r.Architecture, missing)
			r.CompatibilityMsg = fmt.Sprintf(
				"Cannot run on current CPU (missing: %s)",
				strings.Join(missingNames, ", "),
			)
		} else {
			// Version mismatch
			r.CompatibilityMsg = fmt.Sprintf(
				"Cannot run on current CPU (requires %s %s, have %s %s)",
				r.Architecture.String(),
				r.DetectedVersion.String(r.Architecture),
				caps.Architecture,
				caps.VersionStr,
			)
		}
	}
}

// IsFatBinary checks if the given file is a fat binary
func IsFatBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	// Read magic marker
	magic := make([]byte, 8)
	if _, err := f.Read(magic); err != nil {
		return false, nil // Not a fat binary if we can't read magic
	}

	// Check if it matches ADIPOFAT magic
	expectedMagic := [8]byte{'A', 'D', 'I', 'P', 'O', 'F', 'A', 'T'}
	for i := 0; i < 8; i++ {
		if magic[i] != expectedMagic[i] {
			return false, nil
		}
	}

	return true, nil
}
