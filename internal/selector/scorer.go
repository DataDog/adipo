// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package selector

import (
	"sort"

	"github.com/DataDog/adipo/internal/format"
)

// BinaryScore represents a scored binary
type BinaryScore struct {
	Binary *format.BinaryMetadata
	Index  int
	Score  int
}

// Scorer calculates scores for binaries
type Scorer struct {
	detectedCPUAlias string // Detected CPU alias (e.g., "zen3", "apple-m1")
}

// NewScorer creates a new scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// NewScorerWithCPUAlias creates a scorer with CPU alias for matching
func NewScorerWithCPUAlias(cpuAlias string) *Scorer {
	return &Scorer{
		detectedCPUAlias: cpuAlias,
	}
}

// Score calculates a score for a binary
// Higher score = better match
func (s *Scorer) Score(binary *format.BinaryMetadata) int {
	score := 0

	// Base score from priority field (highest weight)
	score += int(binary.Priority) * 1000

	// CPU alias match bonus (high priority boost for matching CPU hints)
	// This allows preferring binaries tuned for specific CPUs even at same version level
	// Example: prefer zen3-tuned binary over skylake-tuned when running on Zen 3
	score += s.cpuAliasScore(binary) * 500

	// Score from version level (higher version = better)
	score += s.versionScore(binary) * 100

	// Score from feature utilization (more features = better optimized)
	score += s.featureScore(binary) * 10

	// Penalty for larger size (prefer smaller if equal features)
	// Subtract 1 point per MB of compressed size
	score -= int(binary.CompressedSize / (1024 * 1024))

	return score
}

// cpuAliasScore returns 1 if CPU alias matches, 0 otherwise
func (s *Scorer) cpuAliasScore(binary *format.BinaryMetadata) int {
	// No bonus if we don't have a detected CPU alias
	if s.detectedCPUAlias == "" {
		return 0
	}

	// Check if binary has a CPU hint
	binaryHint := binary.GetCPUHint()
	if binaryHint == "" {
		return 0 // Binary has no hint, no bonus
	}

	// Bonus if hints match exactly
	if binaryHint == s.detectedCPUAlias {
		return 1 // Will be multiplied by 500 in Score()
	}

	return 0
}

// versionScore returns a score based on the architecture version
func (s *Scorer) versionScore(binary *format.BinaryMetadata) int {
	return int(binary.ArchVersion)
}

// featureScore counts how many features this binary requires/utilizes
func (s *Scorer) featureScore(binary *format.BinaryMetadata) int {
	count := 0
	mask := binary.RequiredFeatures

	// Count set bits in the feature mask
	for mask > 0 {
		if mask&1 == 1 {
			count++
		}
		mask >>= 1
	}

	// Also count extended features
	for _, extMask := range binary.ExtFeatures {
		for extMask > 0 {
			if extMask&1 == 1 {
				count++
			}
			extMask >>= 1
		}
	}

	return count
}

// RankBinaries scores and ranks a list of binaries
// Returns a sorted list (highest score first)
func (s *Scorer) RankBinaries(binaries []*format.BinaryMetadata) []BinaryScore {
	scores := make([]BinaryScore, len(binaries))

	for i, binary := range binaries {
		scores[i] = BinaryScore{
			Binary: binary,
			Index:  i,
			Score:  s.Score(binary),
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
}
