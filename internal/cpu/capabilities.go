package cpu

import (
	"github.com/corentin-chary/adipo/internal/format"
)

// Capabilities represents detected CPU capabilities
type Capabilities struct {
	Architecture string              // "x86-64" or "aarch64"
	ArchType     format.Architecture // Architecture enum
	Version      format.ArchVersion  // Architecture version
	VersionStr   string              // Version string ("v1", "v2", "v8.0", etc.)
	Features     map[string]bool     // Feature name -> present
	FeatureMask  uint64              // Primary 64-bit feature mask
	ExtMasks     [4]uint64           // Extended feature masks (for future use)
}

// NewCapabilities creates a new Capabilities struct
func NewCapabilities(arch string) *Capabilities {
	return &Capabilities{
		Architecture: arch,
		Features:     make(map[string]bool),
	}
}

// HasFeature checks if a specific feature is present
func (c *Capabilities) HasFeature(feature string) bool {
	return c.Features[feature]
}

// HasAllFeatures checks if all specified features are present
func (c *Capabilities) HasAllFeatures(required uint64) bool {
	return (c.FeatureMask & required) == required
}

// IsCompatibleWith checks if the CPU can run a binary with the given requirements
func (c *Capabilities) IsCompatibleWith(arch format.Architecture, version format.ArchVersion, requiredFeatures uint64) bool {
	// Architecture must match
	if c.ArchType != arch {
		return false
	}

	// Version must be compatible (CPU version >= required version)
	if c.Version < version {
		return false
	}

	// All required features must be present
	if !c.HasAllFeatures(requiredFeatures) {
		return false
	}

	return true
}

// String returns a string representation of the capabilities
func (c *Capabilities) String() string {
	return c.Architecture + "-" + c.VersionStr
}

// GetFeatureList returns a sorted list of feature names
func (c *Capabilities) GetFeatureList() []string {
	features := make([]string, 0, len(c.Features))
	for feature, present := range c.Features {
		if present {
			features = append(features, feature)
		}
	}
	return features
}
