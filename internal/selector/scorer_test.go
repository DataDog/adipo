package selector

import (
	"testing"

	"github.com/DataDog/adipo/internal/format"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name      string
		bin       *format.BinaryMetadata
		wantScore int
	}{
		{
			name: "x86-64 v3",
			bin: &format.BinaryMetadata{
				Architecture:   format.ArchX86_64,
				ArchVersion:    format.X86_64_V3,
				CompressedSize: 1024, // 1KB penalty = -1
			},
			wantScore: 299, // version 3 * 100 - 1
		},
		{
			name: "x86-64 v1",
			bin: &format.BinaryMetadata{
				Architecture:   format.ArchX86_64,
				ArchVersion:    format.X86_64_V1,
				CompressedSize: 0,
			},
			wantScore: 100, // version 1 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScorer()
			score := s.Score(tt.bin)
			if score != tt.wantScore {
				t.Errorf("Score() = %d, want %d", score, tt.wantScore)
			}
		})
	}
}

func TestRankBinaries(t *testing.T) {
	binaries := []*format.BinaryMetadata{
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V1},
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V3}, // should rank highest
		{Architecture: format.ArchX86_64, ArchVersion: format.X86_64_V2},
	}

	s := NewScorer()
	ranked := s.RankBinaries(binaries)

	if len(ranked) != 3 {
		t.Fatalf("RankBinaries() returned %d results, want 3", len(ranked))
	}

	// Should be sorted by score descending
	if ranked[0].Binary.ArchVersion != format.X86_64_V3 {
		t.Errorf("Best binary should be v3, got %v", ranked[0].Binary.ArchVersion)
	}
	if ranked[1].Binary.ArchVersion != format.X86_64_V2 {
		t.Errorf("Second binary should be v2, got %v", ranked[1].Binary.ArchVersion)
	}
	if ranked[2].Binary.ArchVersion != format.X86_64_V1 {
		t.Errorf("Third binary should be v1, got %v", ranked[2].Binary.ArchVersion)
	}

	// Verify scores are descending
	if ranked[0].Score < ranked[1].Score || ranked[1].Score < ranked[2].Score {
		t.Errorf("Scores not in descending order: %d, %d, %d",
			ranked[0].Score, ranked[1].Score, ranked[2].Score)
	}
}

func TestRankBinaries_EmptyInput(t *testing.T) {
	s := NewScorer()
	ranked := s.RankBinaries([]*format.BinaryMetadata{})

	if len(ranked) != 0 {
		t.Errorf("RankBinaries([]) should return empty slice, got %d items", len(ranked))
	}
}
