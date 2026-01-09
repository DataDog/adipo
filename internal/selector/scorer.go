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
type Scorer struct{}

// NewScorer creates a new scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// Score calculates a score for a binary
// Higher score = better match
func (s *Scorer) Score(binary *format.BinaryMetadata) int {
	score := 0

	// Base score from priority field (highest weight)
	score += int(binary.Priority) * 1000

	// Score from version level (higher version = better)
	score += s.versionScore(binary) * 100

	// Score from feature utilization (more features = better optimized)
	score += s.featureScore(binary) * 10

	// Penalty for larger size (prefer smaller if equal features)
	// Subtract 1 point per KB of compressed size
	score -= int(binary.CompressedSize / 1024)

	return score
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
