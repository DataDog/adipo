package selector

import (
	"github.com/corentin-chary/adipo/internal/cpu"
	"github.com/corentin-chary/adipo/internal/format"
)

// Matcher checks binary compatibility with CPU capabilities
type Matcher struct {
	caps *cpu.Capabilities
}

// NewMatcher creates a new compatibility matcher
func NewMatcher(caps *cpu.Capabilities) *Matcher {
	return &Matcher{caps: caps}
}

// IsCompatible checks if a binary is compatible with the CPU
func (m *Matcher) IsCompatible(meta *format.BinaryMetadata) bool {
	return m.caps.IsCompatibleWith(meta.Architecture, meta.ArchVersion, meta.RequiredFeatures)
}

// IsArchitectureMatch checks if the architecture matches
func (m *Matcher) IsArchitectureMatch(meta *format.BinaryMetadata) bool {
	return m.caps.ArchType == meta.Architecture
}

// IsVersionCompatible checks if the CPU version is compatible
func (m *Matcher) IsVersionCompatible(meta *format.BinaryMetadata) bool {
	// CPU version must be >= required version
	return m.caps.Version >= meta.ArchVersion
}

// HasAllFeatures checks if the CPU has all required features
func (m *Matcher) HasAllFeatures(meta *format.BinaryMetadata) bool {
	return m.caps.HasAllFeatures(meta.RequiredFeatures)
}

// FilterCompatible filters a list of binaries to only compatible ones
func (m *Matcher) FilterCompatible(binaries []*format.BinaryMetadata) []*format.BinaryMetadata {
	compatible := make([]*format.BinaryMetadata, 0)
	for _, binary := range binaries {
		if m.IsCompatible(binary) {
			compatible = append(compatible, binary)
		}
	}
	return compatible
}
