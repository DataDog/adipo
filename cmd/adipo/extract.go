package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DataDog/adipo/internal/compression"
	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/selector"
	"github.com/spf13/cobra"
)

var extractFlags struct {
	target string
	output string
	all    bool
}

var extractCmd = &cobra.Command{
	Use:   "extract [flags] fat-binary",
	Short: "Extract a binary from a fat binary",
	Long: `Extract a specific binary from a fat binary.

Target can be:
  - "auto": Extract the best binary for the current CPU
  - Index number (e.g., "0", "1", "2")
  - Architecture specification (e.g., "x86-64-v3", "aarch64-v9.0")`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func init() {
	extractCmd.Flags().StringVarP(&extractFlags.target, "target", "t", "auto", "Target specification (auto, index, or arch spec)")
	extractCmd.Flags().StringVarP(&extractFlags.output, "output", "o", "", "Output file path (required)")
	extractCmd.Flags().BoolVar(&extractFlags.all, "all", false, "Extract all binaries (output must be a directory)")

	if err := extractCmd.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
}

func runExtract(cmd *cobra.Command, args []string) error {
	fatBinary := args[0]

	// Open the fat binary
	reader, err := format.OpenFile(fatBinary)
	if err != nil {
		return fmt.Errorf("failed to open fat binary: %w", err)
	}
	defer func() { _ = reader.Close() }()

	metadata := reader.Metadata()

	if extractFlags.all {
		return extractAll(reader, metadata, extractFlags.output)
	}

	// Select target binary
	index, err := selectTarget(extractFlags.target, metadata)
	if err != nil {
		return err
	}

	fmt.Printf("Extracting binary %d: %s-%s\n",
		index,
		metadata[index].Architecture.String(),
		metadata[index].ArchVersion.String(metadata[index].Architecture))

	// Extract the binary
	return extractBinary(reader, index, metadata[index], extractFlags.output)
}

func selectTarget(target string, metadata []*format.BinaryMetadata) (int, error) {
	// Try "auto" - detect CPU and select best
	if target == "auto" {
		caps, err := cpu.Detect()
		if err != nil {
			return -1, fmt.Errorf("failed to detect CPU: %w", err)
		}

		sel := selector.NewSelector(caps, metadata)
		index, _, err := sel.SelectBinary()
		if err != nil {
			return -1, err
		}

		return index, nil
	}

	// Try as index number
	if index, err := strconv.Atoi(target); err == nil {
		if index < 0 || index >= len(metadata) {
			return -1, fmt.Errorf("invalid index: %d (must be 0-%d)", index, len(metadata)-1)
		}
		return index, nil
	}

	// Try as architecture specification
	spec, err := format.ParseArchSpec(target)
	if err != nil {
		return -1, fmt.Errorf("invalid target specification: %w", err)
	}

	// Find matching binary
	for i, meta := range metadata {
		if meta.Architecture == spec.Architecture && meta.ArchVersion == spec.ArchVersion {
			return i, nil
		}
	}

	return -1, fmt.Errorf("no binary found matching: %s", target)
}

func extractBinary(reader *format.Reader, index int, meta *format.BinaryMetadata, outputPath string) error {
	// Read compressed data
	compressedData, err := reader.GetBinaryData(index)
	if err != nil {
		return fmt.Errorf("failed to read binary data: %w", err)
	}

	// Decompress
	fmt.Printf("Decompressing (%s)...\n", meta.Compression.String())
	decompressedData, err := compression.Decompress(compressedData, meta.Compression, meta.OriginalSize)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	// Write to file
	fmt.Printf("Writing to: %s\n", outputPath)
	if err := os.WriteFile(outputPath, decompressedData, 0755); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Extracted successfully (%d bytes)\n", len(decompressedData))

	return nil
}

func extractAll(reader *format.Reader, metadata []*format.BinaryMetadata, outputDir string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Printf("Extracting %d binaries to: %s\n", len(metadata), outputDir)

	for i, meta := range metadata {
		filename := fmt.Sprintf("binary-%d-%s-%s",
			i,
			meta.Architecture.String(),
			meta.ArchVersion.String(meta.Architecture))

		outputPath := filepath.Join(outputDir, filename)

		fmt.Printf("[%d/%d] %s\n", i+1, len(metadata), filename)

		if err := extractBinary(reader, i, meta, outputPath); err != nil {
			return fmt.Errorf("failed to extract binary %d: %w", i, err)
		}
	}

	fmt.Println("All binaries extracted successfully")

	return nil
}
