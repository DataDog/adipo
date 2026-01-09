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

	cleanupOnExit := true
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
		// TODO: Parse forced specification
		// For now, still detect but we could override
		caps, err = cpu.Detect()
		if err != nil {
			fatal("failed to detect CPU: %v", err)
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
			fatal("no compatible binary found: %v", err)
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
			fatal("no compatible binary found: %v", err)
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

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "adipo stub: "+format+"\n", args...)
	os.Exit(1)
}
