// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DataDog/adipo/internal/features"
	"github.com/DataDog/adipo/internal/format"
)

// InstructionStatistics tracks statistics about instructions analyzed
type InstructionStatistics struct {
	TotalInstructions  uint64
	FeatureGroupCounts map[string]uint64 // feature name -> count
}

// NewInstructionStatistics creates a new statistics tracker
func NewInstructionStatistics() *InstructionStatistics {
	return &InstructionStatistics{
		FeatureGroupCounts: make(map[string]uint64),
	}
}

// RecordInstruction records an instruction with its detected features
func (s *InstructionStatistics) RecordInstruction(insnFeatures uint64, arch format.Architecture) {
	s.TotalInstructions++

	// Skip baseline instructions (no special features)
	if insnFeatures == 0 {
		return
	}

	// Get feature names for this architecture
	var featureNames map[uint64]string
	switch arch {
	case format.ArchX86_64:
		featureNames = features.X86FeatureNames
	case format.ArchARM64:
		featureNames = features.ARMFeatureNames
	default:
		return
	}

	// Count each feature present in this instruction
	for bit, name := range featureNames {
		if insnFeatures&bit != 0 {
			s.FeatureGroupCounts[name]++
		}
	}
}

// Format returns a formatted string representation of the statistics
func (s *InstructionStatistics) Format() string {
	var sb strings.Builder

	sb.WriteString("Instruction Statistics:\n")
	sb.WriteString(fmt.Sprintf("  Total instructions analyzed: %s\n", formatNumber(s.TotalInstructions)))

	if len(s.FeatureGroupCounts) == 0 {
		sb.WriteString("  No feature-specific instructions detected\n")
		return sb.String()
	}

	sb.WriteString("\n  Feature Group Counts:\n")

	// Sort features by count (descending)
	type featureCount struct {
		name  string
		count uint64
	}
	var sorted []featureCount
	for name, count := range s.FeatureGroupCounts {
		sorted = append(sorted, featureCount{name, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	// Display top features
	for _, fc := range sorted {
		percentage := float64(fc.count) / float64(s.TotalInstructions) * 100
		sb.WriteString(fmt.Sprintf("    %-20s %s (%.1f%%)\n",
			fc.name+" instructions:",
			formatNumber(fc.count),
			percentage))
	}

	return sb.String()
}

// formatNumber formats a number with thousands separators
func formatNumber(n uint64) string {
	// Convert to string
	s := fmt.Sprintf("%d", n)

	// Add commas for thousands
	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	return result.String()
}
