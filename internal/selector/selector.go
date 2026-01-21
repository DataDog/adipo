// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package selector

import (
	"fmt"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
)

// Selector selects the best binary for the current CPU
type Selector struct {
	caps     *cpu.Capabilities
	binaries []*format.BinaryMetadata
	matcher  *Matcher
	scorer   *Scorer
}

// NewSelector creates a new binary selector
func NewSelector(caps *cpu.Capabilities, binaries []*format.BinaryMetadata) *Selector {
	// Detect CPU alias for scoring (if available)
	cpuAlias := ""
	if caps.CPUModel != nil {
		cpuAlias = cpu.DetectCPUAlias(caps.CPUModel, caps.ArchType)
	}

	return &Selector{
		caps:     caps,
		binaries: binaries,
		matcher:  NewMatcher(caps),
		scorer:   NewScorerWithCPUAlias(cpuAlias),
	}
}

// SelectBinary selects the best compatible binary for the current CPU
// Returns the index of the selected binary or an error if none are compatible
func (s *Selector) SelectBinary() (int, *format.BinaryMetadata, error) {
	// Filter to only compatible binaries
	compatible := s.matcher.FilterCompatible(s.binaries)
	if len(compatible) == 0 {
		return -1, nil, format.ErrNoCompatibleBinary
	}

	// Rank compatible binaries by score
	ranked := s.scorer.RankBinaries(compatible)
	if len(ranked) == 0 {
		return -1, nil, format.ErrNoCompatibleBinary
	}

	// Select the highest scored binary
	best := ranked[0]

	// Find the original index
	originalIndex := -1
	for i, binary := range s.binaries {
		if binary == best.Binary {
			originalIndex = i
			break
		}
	}

	return originalIndex, best.Binary, nil
}

// SelectBinaryVerbose selects the best binary and returns detailed information
func (s *Selector) SelectBinaryVerbose() (*SelectionResult, error) {
	// Filter to only compatible binaries
	compatible := s.matcher.FilterCompatible(s.binaries)
	if len(compatible) == 0 {
		return nil, format.ErrNoCompatibleBinary
	}

	// Rank all binaries (both compatible and incompatible) for information
	allRanked := s.scorer.RankBinaries(s.binaries)
	compatibleRanked := s.scorer.RankBinaries(compatible)

	// Select the best compatible binary
	best := compatibleRanked[0]

	// Find the original index
	originalIndex := -1
	for i, binary := range s.binaries {
		if binary == best.Binary {
			originalIndex = i
			break
		}
	}

	result := &SelectionResult{
		SelectedIndex:       originalIndex,
		SelectedBinary:      best.Binary,
		SelectedScore:       best.Score,
		TotalBinaries:       len(s.binaries),
		CompatibleBinaries:  len(compatible),
		AllScores:           allRanked,
		CompatibleScores:    compatibleRanked,
		CPUCapabilities:     s.caps,
	}

	return result, nil
}

// SelectionResult contains detailed information about the selection process
type SelectionResult struct {
	SelectedIndex      int
	SelectedBinary     *format.BinaryMetadata
	SelectedScore      int
	TotalBinaries      int
	CompatibleBinaries int
	AllScores          []BinaryScore
	CompatibleScores   []BinaryScore
	CPUCapabilities    *cpu.Capabilities
}

// String returns a human-readable description of the selection
func (r *SelectionResult) String() string {
	return fmt.Sprintf(
		"Selected binary %d (score: %d) from %d compatible binaries out of %d total\n"+
			"CPU: %s\n"+
			"Binary: %s-%s (features: 0x%x)",
		r.SelectedIndex,
		r.SelectedScore,
		r.CompatibleBinaries,
		r.TotalBinaries,
		r.CPUCapabilities.String(),
		r.SelectedBinary.Architecture.String(),
		r.SelectedBinary.ArchVersion.String(r.SelectedBinary.Architecture),
		r.SelectedBinary.RequiredFeatures,
	)
}
