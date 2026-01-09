package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/corentin-chary/adipo/internal/compression"
	"github.com/corentin-chary/adipo/internal/elf"
	"github.com/corentin-chary/adipo/internal/format"
	"github.com/corentin-chary/adipo/internal/macho"
	"github.com/corentin-chary/adipo/internal/stub"
	"github.com/spf13/cobra"
)

var createFlags struct {
	output          string
	binaries        []string
	compress        string
	level           int
	verify          bool
	noStub          bool
}

var createCmd = &cobra.Command{
	Use:   "create [flags] [input-binaries...]",
	Short: "Create a fat binary from multiple input binaries",
	Long: `Create a fat binary containing multiple versions of the same executable,
each optimized for different CPU micro-architectures.

Input binaries can be specified in two ways:
1. Positional arguments with --binary flags for explicit specifications
2. Positional arguments only (auto-detection from ELF headers)

Architecture specification format:
  ARCH-VERSION[,FEATURE1,FEATURE2,...]

Examples:
  x86-64-v1, x86-64-v2, x86-64-v3, x86-64-v4
  amd64-v2      (amd64 is an alias for x86-64)
  aarch64-v8.0, aarch64-v8.1, aarch64-v9.0
  arm64-v9.0    (arm64 is an alias for aarch64)
  x86-64-v3,avx2,bmi2
  aarch64-v9.0,sve2

Usage:
  # With explicit specifications
  adipo create -o app.fat --binary app-v1:x86-64-v1 --binary app-v2:x86-64-v2

  # With auto-detection (uses baseline versions)
  adipo create -o app.fat app-v1 app-v2 app-v3

  # Mixed mode
  adipo create -o app.fat app-baseline --binary app-optimized:x86-64-v3,avx2`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createFlags.output, "output", "o", "", "Output fat binary path (required)")
	createCmd.Flags().StringArrayVar(&createFlags.binaries, "binary", []string{}, "Input binary with specification (FILE:SPEC)")
	createCmd.Flags().StringVar(&createFlags.compress, "compress", "zstd", "Compression algorithm (zstd, lz4, gzip, none)")
	createCmd.Flags().IntVar(&createFlags.level, "level", 3, "Compression level")
	createCmd.Flags().BoolVar(&createFlags.verify, "verify", true, "Verify binary format (ELF/Mach-O) and executability")
	createCmd.Flags().BoolVar(&createFlags.noStub, "no-stub", false, "Create fat binary without self-extracting stub (saves space, requires extraction tool)")

	if err := createCmd.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Validate output path
	if createFlags.output == "" {
		return fmt.Errorf("output path is required")
	}

	// Parse compression algorithm
	compAlgo, err := parseCompressionAlgo(createFlags.compress)
	if err != nil {
		return err
	}

	// Collect input binaries
	inputs, err := collectInputBinaries(args)
	if err != nil {
		return err
	}

	if len(inputs) == 0 {
		return fmt.Errorf("no input binaries specified")
	}

	fmt.Printf("Creating fat binary with %d binaries\n", len(inputs))

	// Load stub binary (unless --no-stub)
	var stubData []byte
	var stubArch format.Architecture
	var stubArchVer format.ArchVersion

	if !createFlags.noStub {
		stubData, err = stub.GetStubBinary()
		if err != nil {
			return fmt.Errorf("failed to load stub binary: %w", err)
		}

		fmt.Printf("Stub size: %d bytes\n", len(stubData))

		// Detect stub architecture
		stubArch, stubArchVer, err = detectStubArchitecture(stubData)
		if err != nil {
			return fmt.Errorf("failed to detect stub architecture: %w", err)
		}

		fmt.Printf("Stub architecture: %s-%s\n", stubArch.String(), stubArchVer.String(stubArch))
	} else {
		fmt.Printf("Creating without stub (--no-stub)\n")
	}

	// Process each input binary
	entries := make([]*format.BinaryEntry, 0, len(inputs))

	for i, input := range inputs {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(inputs), input.Path)

		entry, err := processBinary(input, compAlgo, createFlags.level, createFlags.verify)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", input.Path, err)
		}

		fmt.Printf("  Architecture: %s-%s\n",
			entry.Metadata.Architecture.String(),
			entry.Metadata.ArchVersion.String(entry.Metadata.Architecture))
		fmt.Printf("  Original size: %d bytes\n", entry.Metadata.OriginalSize)
		fmt.Printf("  Compressed size: %d bytes (%.1f%%)\n",
			entry.Metadata.CompressedSize,
			float64(entry.Metadata.CompressedSize)/float64(entry.Metadata.OriginalSize)*100)

		entries = append(entries, entry)
	}

	// Create the fat binary
	fmt.Printf("\nWriting fat binary to: %s\n", createFlags.output)

	err = format.WriteToFile(createFlags.output, stubData, entries, stubArch, stubArchVer)
	if err != nil {
		return fmt.Errorf("failed to write fat binary: %w", err)
	}

	// Get final size
	info, err := os.Stat(createFlags.output)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}
	fmt.Printf("Fat binary created successfully (%d bytes)\n", info.Size())

	return nil
}

// InputBinary represents an input binary with optional specification
type InputBinary struct {
	Path string
	Spec *format.ArchSpec
}

// collectInputBinaries collects and parses input binaries from args and flags
func collectInputBinaries(args []string) ([]*InputBinary, error) {
	inputs := make([]*InputBinary, 0)

	// Process --binary flags first
	for _, binSpec := range createFlags.binaries {
		parts := strings.SplitN(binSpec, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --binary format (expected FILE:SPEC): %s", binSpec)
		}

		path := strings.TrimSpace(parts[0])
		specStr := strings.TrimSpace(parts[1])

		spec, err := format.ParseArchSpec(specStr)
		if err != nil {
			return nil, fmt.Errorf("invalid architecture spec for %s: %w", path, err)
		}

		inputs = append(inputs, &InputBinary{
			Path: path,
			Spec: spec,
		})
	}

	// Process positional arguments (auto-detection)
	for _, path := range args {
		inputs = append(inputs, &InputBinary{
			Path: path,
			Spec: nil, // Will be auto-detected
		})
	}

	return inputs, nil
}

// processBinary processes a single input binary
func processBinary(input *InputBinary, compAlgo format.CompressionAlgo, level int, verify bool) (*format.BinaryEntry, error) {
	// Detect binary format
	binaryFormat, err := format.DetectFormat(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect binary format: %w", err)
	}

	// Validate and analyze based on format
	var arch format.Architecture
	var archVersion format.ArchVersion

	switch binaryFormat {
	case format.FormatELF:
		if verify {
			if err := elf.Validate(input.Path); err != nil {
				return nil, fmt.Errorf("ELF validation failed: %w", err)
			}
		}

		elfInfo, err := elf.Analyze(input.Path)
		if err != nil {
			return nil, fmt.Errorf("ELF analysis failed: %w", err)
		}
		arch = elfInfo.Architecture
		archVersion = elfInfo.ArchVersion

	case format.FormatMachO:
		if verify {
			if err := macho.Validate(input.Path); err != nil {
				return nil, fmt.Errorf("Mach-O validation failed: %w", err)
			}
		}

		machoInfo, err := macho.Analyze(input.Path)
		if err != nil {
			return nil, fmt.Errorf("Mach-O analysis failed: %w", err)
		}
		arch = machoInfo.Architecture
		archVersion = machoInfo.ArchVersion

	default:
		return nil, fmt.Errorf("unsupported binary format: %s", binaryFormat.String())
	}

	// Use provided spec or auto-detected info
	var features uint64

	if input.Spec != nil {
		// Use explicit specification
		specArch := input.Spec.Architecture
		archVersion = input.Spec.ArchVersion
		features = input.Spec.RequiredFeatures

		// Validate architecture matches
		if specArch != arch {
			return nil, fmt.Errorf("architecture mismatch: spec says %s but binary is %s",
				specArch.String(), arch.String())
		}
	} else {
		// Use auto-detected info (baseline version)
		features = 0 // Baseline features
	}

	// Read binary data
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Compress binary
	compressed, err := compression.Compress(data, compAlgo, level)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Calculate checksum
	checksum := sha256.Sum256(data)

	// Create metadata
	metadata := &format.BinaryMetadata{
		Architecture:     arch,
		ArchVersion:      archVersion,
		RequiredFeatures: features,
		OriginalSize:     uint64(len(data)),
		CompressedSize:   uint64(len(compressed)),
		Compression:      compAlgo,
		Checksum:         checksum,
		Priority:         uint32(archVersion), // Higher version = higher priority
		Format:           binaryFormat,
	}

	entry := &format.BinaryEntry{
		Data:         compressed,
		Metadata:     metadata,
		OriginalData: data,
	}

	return entry, nil
}

// parseCompressionAlgo parses a compression algorithm name
func parseCompressionAlgo(name string) (format.CompressionAlgo, error) {
	switch strings.ToLower(name) {
	case "none":
		return format.CompressionNone, nil
	case "gzip":
		return format.CompressionGzip, nil
	case "zstd":
		return format.CompressionZstd, nil
	case "lz4":
		return format.CompressionLZ4, nil
	default:
		return format.CompressionNone, fmt.Errorf("unknown compression algorithm: %s", name)
	}
}

// detectStubArchitecture detects the architecture of the stub binary
func detectStubArchitecture(stubData []byte) (format.Architecture, format.ArchVersion, error) {
	// Write stub data to a temporary file for analysis
	tmpFile, err := os.CreateTemp("", "adipo-stub-*")
	if err != nil {
		return format.ArchUnknown, 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.Write(stubData); err != nil {
		return format.ArchUnknown, 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return format.ArchUnknown, 0, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Detect format
	binaryFormat, err := format.DetectFormat(tmpFile.Name())
	if err != nil {
		return format.ArchUnknown, 0, fmt.Errorf("failed to detect stub format: %w", err)
	}

	// Analyze based on format
	switch binaryFormat {
	case format.FormatELF:
		elfInfo, err := elf.Analyze(tmpFile.Name())
		if err != nil {
			return format.ArchUnknown, 0, fmt.Errorf("failed to analyze ELF stub: %w", err)
		}
		return elfInfo.Architecture, elfInfo.ArchVersion, nil

	case format.FormatMachO:
		machoInfo, err := macho.Analyze(tmpFile.Name())
		if err != nil {
			return format.ArchUnknown, 0, fmt.Errorf("failed to analyze Mach-O stub: %w", err)
		}
		return machoInfo.Architecture, machoInfo.ArchVersion, nil

	default:
		return format.ArchUnknown, 0, fmt.Errorf("unsupported stub format: %s", binaryFormat.String())
	}
}
