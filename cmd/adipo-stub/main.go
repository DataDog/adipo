package main

import (
	"fmt"
	"os"

	"github.com/DataDog/adipo/internal/compression"
	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/extractor"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/selector"
)

func main() {
	// Check environment variables for configuration (these override header defaults)
	envVerbose := os.Getenv("ADIPO_VERBOSE")
	envDebug := os.Getenv("ADIPO_DEBUG")
	envPreferDisk := os.Getenv("ADIPO_PREFER_DISK")
	envExtractDir := os.Getenv("ADIPO_EXTRACT_DIR")
	envCleanupOnExit := os.Getenv("ADIPO_CLEANUP_ON_EXIT")
	forceSpec := os.Getenv("ADIPO_FORCE")

	// Default debug setting (env var only)
	debug := envDebug == "1"

	// Open ourselves - use os.Executable for cross-platform support
	exePath, err := os.Executable()
	if err != nil {
		fatal("failed to get executable path: %v", err)
	}

	self, err := os.Open(exePath)
	if err != nil {
		fatal("failed to open self: %v", err)
	}
	defer func() { _ = self.Close() }()

	// Parse the fat binary format
	reader, err := format.NewReader(self)
	if err != nil {
		fatal("failed to parse fat binary: %v", err)
	}

	header := reader.Header()
	metadata := reader.Metadata()

	// Read stub settings from header and merge with environment variables
	stubSettings := header.GetStubSettings()
	defaultExtractDir := header.GetDefaultExtractDir()

	// Determine effective settings (env vars override header defaults)
	verbose := false
	if envVerbose != "" {
		verbose = envVerbose == "1"
	} else {
		verbose = (stubSettings & format.StubSettingVerbose) != 0
	}
	if envDebug == "1" {
		verbose = true // debug implies verbose
	}

	preferDisk := false
	if envPreferDisk != "" {
		preferDisk = envPreferDisk == "1"
	}

	extractDir := envExtractDir
	if extractDir == "" {
		extractDir = defaultExtractDir
	}

	var cleanupOnExit bool
	if envCleanupOnExit != "" {
		cleanupOnExit = envCleanupOnExit == "1"
	} else {
		cleanupOnExit = (stubSettings & format.StubSettingCleanupOnExit) != 0
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "adipo stub: found %d binaries\n", header.NumBinaries)
	}

	// Detect CPU capabilities (unless forced)
	var caps *cpu.Capabilities
	if forceSpec != "" {
		if verbose {
			fmt.Fprintf(os.Stderr, "adipo stub: forced specification: %s\n", forceSpec)
		}
		// Parse forced specification and create synthetic capabilities
		caps, err = createForcedCapabilities(forceSpec)
		if err != nil {
			fatal("failed to parse forced specification '%s': %v", forceSpec, err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "adipo stub: forcing CPU: %s\n", caps.String())
			if debug {
				fmt.Fprintf(os.Stderr, "adipo stub: forced features: %v\n", caps.FeatureList())
			}
		}
	} else {
		caps, err = cpu.Detect()
		if err != nil {
			fatal("failed to detect CPU: %v", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "adipo stub: detected CPU: %s\n", caps.String())
			if debug {
				fmt.Fprintf(os.Stderr, "adipo stub: CPU features: %v\n", caps.FeatureList())
			}
		}
	}

	// Select the best binary
	sel := selector.NewSelector(caps, metadata)

	var selectedIndex int
	var selectedBinary *format.BinaryMetadata

	if verbose || debug {
		result, err := sel.SelectBinaryVerbose()
		if err != nil {
			printDetailedSelectionError(caps, metadata)
			os.Exit(1)
		}

		selectedIndex = result.SelectedIndex
		selectedBinary = result.SelectedBinary

		if verbose {
			fmt.Fprintf(os.Stderr, "adipo stub: selected binary %d: %s-%s (score: %d)\n",
				selectedIndex,
				selectedBinary.Architecture.String(),
				selectedBinary.ArchVersion.String(selectedBinary.Architecture),
				result.SelectedScore,
			)
		}

		if debug {
			fmt.Fprintf(os.Stderr, "adipo stub: selection details:\n%s\n", result.String())
		}
	} else {
		var err error
		selectedIndex, selectedBinary, err = sel.SelectBinary()
		if err != nil {
			printDetailedSelectionError(caps, metadata)
			os.Exit(1)
		}
	}

	// Read the compressed binary data
	if verbose {
		fmt.Fprintf(os.Stderr, "adipo stub: reading binary data (compressed: %d bytes)\n",
			selectedBinary.CompressedSize)
	}

	compressedData, err := reader.GetBinaryData(selectedIndex)
	if err != nil {
		fatal("failed to read binary data: %v", err)
	}

	// Decompress the binary
	if verbose {
		fmt.Fprintf(os.Stderr, "adipo stub: decompressing binary (%s, %d -> %d bytes)\n",
			selectedBinary.Compression.String(),
			selectedBinary.CompressedSize,
			selectedBinary.OriginalSize,
		)
	}

	decompressedData, err := compression.Decompress(
		compressedData,
		selectedBinary.Compression,
		selectedBinary.OriginalSize,
	)
	if err != nil {
		fatal("failed to decompress binary: %v", err)
	}

	// Verify checksum
	if debug {
		fmt.Fprintf(os.Stderr, "adipo stub: verifying checksum\n")
	}
	// TODO: Add checksum verification

	// Extract and execute
	opts := &extractor.ExecutionOptions{
		Args:          os.Args[1:], // Pass through arguments (skip argv[0])
		Env:           extractor.GetEnvironment(),
		PreferMemory:  !preferDisk,
		Verbose:       verbose,
		TempDir:       extractDir,
		CleanupOnExit: cleanupOnExit,
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "adipo stub: extracting and executing binary\n")
	}

	err = extractor.ExtractAndExecute(decompressedData, "binary", opts)
	if err != nil {
		fatal("failed to execute: %v", err)
	}

	// This line should never be reached
	fatal("exec returned unexpectedly")
}

func createForcedCapabilities(spec string) (*cpu.Capabilities, error) {
	// Parse the architecture specification
	archSpec, err := format.ParseArchSpec(spec)
	if err != nil {
		return nil, fmt.Errorf("invalid architecture specification: %w", err)
	}

	// Create synthetic capabilities that claim to support this architecture/version
	caps := cpu.NewCapabilities(archSpec.Architecture.String())
	caps.ArchType = archSpec.Architecture
	caps.Version = archSpec.ArchVersion
	caps.VersionStr = archSpec.ArchVersion.String(archSpec.Architecture)

	// Set feature mask to include all features for this version
	// This ensures compatibility checks pass for the forced selection
	caps.FeatureMask = archSpec.RequiredFeatures

	// Set all extended feature masks to all-bits-on to pass any extended feature checks
	// (Forced selection is mainly for basic arch/version, not fine-grained feature matching)
	for i := 0; i < len(caps.ExtMasks); i++ {
		caps.ExtMasks[i] = ^uint64(0) // All bits set
	}

	// Add feature names to the feature map for display
	if len(archSpec.FeatureNames) > 0 {
		for _, name := range archSpec.FeatureNames {
			caps.Features[name] = struct{}{}
		}
	} else {
		// If no specific features, add a marker for debugging output
		caps.Features["forced"] = struct{}{}
	}

	return caps, nil
}

func printDetailedSelectionError(caps *cpu.Capabilities, binaries []*format.BinaryMetadata) {
	fmt.Fprintf(os.Stderr, "\nadipo stub: ERROR - No compatible binary found\n\n")

	// Show detected CPU
	fmt.Fprintf(os.Stderr, "Detected CPU:\n")
	fmt.Fprintf(os.Stderr, "  Architecture: %s\n", caps.String())
	fmt.Fprintf(os.Stderr, "  Features:     %v\n", caps.FeatureList())
	fmt.Fprintf(os.Stderr, "  Feature mask: 0x%x\n\n", caps.FeatureMask)

	// Show available binaries
	fmt.Fprintf(os.Stderr, "Available binaries in this fat binary:\n")
	matcher := selector.NewMatcher(caps)
	for i, bin := range binaries {
		archStr := bin.Architecture.String()
		versionStr := bin.ArchVersion.String(bin.Architecture)
		compatible := matcher.IsCompatible(bin)

		statusIcon := "✗"
		if compatible {
			statusIcon = "✓"
		}

		fmt.Fprintf(os.Stderr, "  %s Binary %d: %s-%s (features: 0x%x)\n",
			statusIcon, i, archStr, versionStr, bin.RequiredFeatures)

		// Show why incompatible
		if !compatible {
			reasons := getIncompatibilityReasons(caps, bin)
			for _, reason := range reasons {
				fmt.Fprintf(os.Stderr, "      → %s\n", reason)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\nSuggestions:\n")
	fmt.Fprintf(os.Stderr, "  • Run 'ADIPO_VERBOSE=1 %s' to see more details\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  • Use 'adipo inspect %s' to see all bundled binaries\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  • Use 'adipo extract -t <index> %s' to extract a specific binary\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  • Use 'ADIPO_FORCE=<archspec> %s' to force selection (e.g., ADIPO_FORCE=x86-64-v1)\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  • Rebuild the fat binary with variants for your CPU architecture\n")
	fmt.Fprintf(os.Stderr, "\n")
}

func getIncompatibilityReasons(caps *cpu.Capabilities, bin *format.BinaryMetadata) []string {
	reasons := []string{}

	// Check architecture mismatch
	if caps.ArchType != bin.Architecture {
		reasons = append(reasons, fmt.Sprintf("Architecture mismatch (need %s, have %s)",
			bin.Architecture.String(), caps.ArchType.String()))
		return reasons // If arch doesn't match, no point checking features
	}

	// Check version/feature requirements
	if bin.Architecture == format.ArchX86_64 {
		// Check if CPU version is less than required
		if caps.Version < bin.ArchVersion {
			reasons = append(reasons, fmt.Sprintf("CPU is %s but binary requires %s or higher",
				caps.VersionStr, bin.ArchVersion.String(bin.Architecture)))
		}
	} else if bin.Architecture == format.ArchARM64 {
		// For ARM64, check if CPU version is less than required
		if caps.Version < bin.ArchVersion {
			reasons = append(reasons, fmt.Sprintf("CPU is %s but binary requires %s or higher",
				caps.VersionStr, bin.ArchVersion.String(bin.Architecture)))
		}
	}

	// Check missing features
	missingFeatures := bin.RequiredFeatures & ^caps.FeatureMask
	if missingFeatures != 0 {
		reasons = append(reasons, fmt.Sprintf("Missing required features: 0x%x", missingFeatures))
	}

	// Check extended features if any
	for regID := 0; regID < len(bin.ExtFeatures); regID++ {
		requiredMask := bin.ExtFeatures[regID]
		if requiredMask == 0 {
			continue // No requirements for this register
		}
		haveMask := caps.ExtMasks[regID]
		missingMask := requiredMask & ^haveMask
		if missingMask != 0 {
			reasons = append(reasons, fmt.Sprintf("Missing extended features (reg %d): 0x%x", regID, missingMask))
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "Unknown incompatibility")
	}

	return reasons
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "adipo stub: "+format+"\n", args...)
	os.Exit(1)
}
