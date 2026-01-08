package format

import (
	"fmt"
	"strings"

	"github.com/corentin-chary/adipo/internal/features"
)

// ArchSpec represents a parsed architecture specification
type ArchSpec struct {
	Architecture     Architecture
	ArchVersion      ArchVersion
	RequiredFeatures uint64
	FeatureNames     []string
}

// ParseArchSpec parses an architecture specification string
// Format: ARCH-VERSION[,FEATURE1,FEATURE2,...]
// Examples:
//   x86-64-v2
//   amd64-v3,avx2       (amd64 is an alias for x86-64)
//   aarch64-v8.1,crc
//   arm64-v9.0,sve2     (arm64 is an alias for aarch64)
func ParseArchSpec(spec string) (*ArchSpec, error) {
	if spec == "" {
		return nil, fmt.Errorf("empty architecture specification")
	}

	// Split by comma to separate base spec from features
	parts := strings.Split(spec, ",")
	baseSpec := strings.TrimSpace(parts[0])
	featureNames := make([]string, 0)
	for i := 1; i < len(parts); i++ {
		feature := strings.TrimSpace(parts[i])
		if feature != "" {
			featureNames = append(featureNames, feature)
		}
	}

	// Parse base specification (ARCH-VERSION)
	arch, version, err := parseBaseSpec(baseSpec)
	if err != nil {
		return nil, err
	}

	result := &ArchSpec{
		Architecture: arch,
		ArchVersion:  version,
		FeatureNames: featureNames,
	}

	// Parse features and build bitmask
	featureMask, err := parseFeaturesForArch(arch, featureNames)
	if err != nil {
		return nil, err
	}
	result.RequiredFeatures = featureMask

	return result, nil
}

// parseBaseSpec parses the base architecture-version specification
func parseBaseSpec(spec string) (Architecture, ArchVersion, error) {
	// Normalize the spec
	spec = strings.ToLower(spec)

	// Handle architecture aliases
	spec = normalizeArchAlias(spec)

	// Try x86-64 formats
	if strings.HasPrefix(spec, "x86-64-") || strings.HasPrefix(spec, "x86_64-") {
		versionStr := strings.TrimPrefix(spec, "x86-64-")
		versionStr = strings.TrimPrefix(versionStr, "x86_64-")
		version, err := parseX86Version(versionStr)
		if err != nil {
			return ArchUnknown, 0, err
		}
		return ArchX86_64, version, nil
	}

	// Try aarch64/ARM64 formats
	if strings.HasPrefix(spec, "aarch64-") || strings.HasPrefix(spec, "arm64-") {
		versionStr := strings.TrimPrefix(spec, "aarch64-")
		versionStr = strings.TrimPrefix(versionStr, "arm64-")
		version, err := parseARMVersion(versionStr)
		if err != nil {
			return ArchUnknown, 0, err
		}
		return ArchARM64, version, nil
	}

	return ArchUnknown, 0, fmt.Errorf("invalid architecture specification: %s", spec)
}

// normalizeArchAlias normalizes architecture aliases
func normalizeArchAlias(spec string) string {
	// amd64 -> x86-64
	spec = strings.ReplaceAll(spec, "amd64-", "x86-64-")
	spec = strings.ReplaceAll(spec, "amd64_", "x86-64-")

	// arm64 -> aarch64 (keep as-is, both are valid)

	return spec
}

// parseX86Version parses an x86-64 version string
func parseX86Version(version string) (ArchVersion, error) {
	switch version {
	case "v1":
		return X86_64_V1, nil
	case "v2":
		return X86_64_V2, nil
	case "v3":
		return X86_64_V3, nil
	case "v4":
		return X86_64_V4, nil
	default:
		return 0, fmt.Errorf("invalid x86-64 version: %s", version)
	}
}

// parseARMVersion parses an ARM64 version string
func parseARMVersion(version string) (ArchVersion, error) {
	switch version {
	case "v8.0", "v8":
		return ARM64_V8_0, nil
	case "v8.1":
		return ARM64_V8_1, nil
	case "v8.2":
		return ARM64_V8_2, nil
	case "v8.3":
		return ARM64_V8_3, nil
	case "v8.4":
		return ARM64_V8_4, nil
	case "v9.0", "v9":
		return ARM64_V9_0, nil
	case "v9.1":
		return ARM64_V9_1, nil
	default:
		return 0, fmt.Errorf("invalid ARM64 version: %s", version)
	}
}

// parseFeaturesForArch parses feature names for a specific architecture
func parseFeaturesForArch(arch Architecture, featureNames []string) (uint64, error) {
	var mask uint64

	for _, feature := range featureNames {
		feature = strings.ToLower(strings.TrimSpace(feature))

		var bit uint64
		var ok bool

		switch arch {
		case ArchX86_64:
			bit, ok = features.ParseX86FeatureName(feature)
			if !ok {
				return 0, fmt.Errorf("unknown x86-64 feature: %s", feature)
			}
		case ArchARM64:
			bit, ok = features.ParseARMFeatureName(feature)
			if !ok {
				return 0, fmt.Errorf("unknown ARM64 feature: %s", feature)
			}
		default:
			return 0, fmt.Errorf("unsupported architecture for feature parsing")
		}

		mask |= bit
	}

	return mask, nil
}

// String returns a string representation of the arch spec
func (s *ArchSpec) String() string {
	base := fmt.Sprintf("%s-%s",
		s.Architecture.String(),
		s.ArchVersion.String(s.Architecture))

	if len(s.FeatureNames) > 0 {
		return base + "," + strings.Join(s.FeatureNames, ",")
	}

	return base
}
