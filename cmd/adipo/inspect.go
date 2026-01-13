package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/selector"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var inspectFlags struct {
	formatType string
	features   bool
	verify     bool
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [flags] fat-binary",
	Short: "Inspect a fat binary and show its contents",
	Long: `Inspect a fat binary and display information about the embedded binaries,
including architecture, version, features, sizes, and compression ratios.`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

func init() {
	inspectCmd.Flags().StringVar(&inspectFlags.formatType, "format", "table", "Output format (table, json, yaml)")
	inspectCmd.Flags().BoolVar(&inspectFlags.features, "features", false, "Show detailed CPU features")
	inspectCmd.Flags().BoolVar(&inspectFlags.verify, "verify", false, "Verify checksums")
}

func runInspect(cmd *cobra.Command, args []string) error {
	fatBinary := args[0]

	// Open the fat binary
	reader, err := format.OpenFile(fatBinary)
	if err != nil {
		return fmt.Errorf("failed to open fat binary: %w", err)
	}
	defer func() { _ = reader.Close() }()

	header := reader.Header()
	metadata := reader.Metadata()

	// Verify checksum if requested
	if inspectFlags.verify {
		fmt.Println("Verifying checksums...")
		if err := reader.VerifyChecksum(); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("Checksum verification passed")
	}

	// Format output based on requested format
	switch inspectFlags.formatType {
	case "json":
		return outputJSON(fatBinary, header, metadata)
	case "yaml":
		return outputYAML(fatBinary, header, metadata)
	case "table":
		return outputTable(fatBinary, header, metadata)
	default:
		return fmt.Errorf("unknown format: %s", inspectFlags.formatType)
	}
}

func outputTable(path string, header *format.FormatHeader, metadata []*format.BinaryMetadata) error {
	fmt.Printf("Fat Binary: %s\n", path)
	fmt.Printf("Format Version: %d\n", header.Version)

	if header.StubSize > 0 {
		fmt.Printf("Stub Size: %d bytes (%.2f MB)\n", header.StubSize, float64(header.StubSize)/(1024*1024))
		fmt.Printf("Stub Architecture: %s-%s\n",
			header.StubArchitecture.String(),
			header.StubArchVersion.String(header.StubArchitecture))
	} else {
		fmt.Printf("Stub: none (extraction required)\n")
	}

	fmt.Printf("Number of Binaries: %d\n", header.NumBinaries)
	fmt.Printf("Default Compression: %s\n", header.CompressionAlgo.String())
	fmt.Println()

	// Detect current CPU to find preferred binary
	var preferredIndex = -1
	caps, err := cpu.Detect()
	if err == nil {
		// Find which binary would be selected for this CPU
		sel := selector.NewSelector(caps, metadata)
		preferredIndex, _, _ = sel.SelectBinary()
	}

	// Calculate total sizes
	var totalOriginal, totalCompressed uint64
	for _, meta := range metadata {
		totalOriginal += meta.OriginalSize
		totalCompressed += meta.CompressedSize
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	_ = table.Append([]string{"Index", "Architecture", "Version", "Features", "Original", "Compressed", "Ratio"})
	_ = table.Append([]string{"-----", "------------", "-------", "--------", "--------", "----------", "-----"})

	for i, meta := range metadata {
		archStr := meta.Architecture.String()
		versionStr := meta.ArchVersion.String(meta.Architecture)

		// Format features
		var featuresStr string
		if inspectFlags.features {
			var featureNames []string
			switch meta.Architecture {
			case format.ArchX86_64:
				featureNames = cpu.FormatX86Features(meta.RequiredFeatures)
			case format.ArchARM64:
				featureNames = cpu.FormatARMFeatures(meta.RequiredFeatures)
			}
			if len(featureNames) > 0 {
				featuresStr = fmt.Sprintf("%d features", len(featureNames))
			} else {
				featuresStr = "(baseline)"
			}
		} else {
			if meta.RequiredFeatures == 0 {
				featuresStr = "(baseline)"
			} else {
				featuresStr = fmt.Sprintf("0x%x", meta.RequiredFeatures)
			}
		}

		originalStr := formatBytes(meta.OriginalSize)
		compressedStr := formatBytes(meta.CompressedSize)
		ratio := float64(meta.CompressedSize) / float64(meta.OriginalSize) * 100

		// Mark preferred binary with *
		indexStr := fmt.Sprintf("%d", i)
		if i == preferredIndex {
			indexStr = fmt.Sprintf("%d *", i)
		}

		_ = table.Append([]string{
			indexStr,
			archStr,
			versionStr,
			featuresStr,
			originalStr,
			compressedStr,
			fmt.Sprintf("%.1f%%", ratio),
		})
	}

	_ = table.Render()

	// Show library path templates if any are set
	hasLibraryPaths := false
	for _, meta := range metadata {
		templates := meta.GetLibraryPathTemplates()
		if len(templates) > 0 {
			hasLibraryPaths = true
			break
		}
	}

	if hasLibraryPaths {
		fmt.Println("\nLibrary Path Templates:")
		for i, meta := range metadata {
			templates := meta.GetLibraryPathTemplates()
			if len(templates) > 0 {
				fmt.Printf("  Binary %d:\n", i)
				for _, template := range templates {
					fmt.Printf("    - %s\n", template)
				}
			}
		}
	}

	// Show preferred binary information
	if preferredIndex >= 0 {
		fmt.Printf("\n* Preferred binary for current CPU (%s %s)\n",
			caps.Architecture,
			caps.Version.String(caps.ArchType))
	} else if caps != nil {
		fmt.Printf("\n⚠ No compatible binary found for current CPU (%s %s)\n",
			caps.Architecture,
			caps.Version.String(caps.ArchType))
	}

	fmt.Printf("\nTotal Original Size: %s\n", formatBytes(totalOriginal))
	fmt.Printf("Total Compressed Size: %s\n", formatBytes(totalCompressed))
	fmt.Printf("Overall Compression Ratio: %.1f%%\n",
		float64(totalCompressed)/float64(totalOriginal)*100)

	return nil
}

func outputJSON(path string, header *format.FormatHeader, metadata []*format.BinaryMetadata) error {
	data := map[string]interface{}{
		"path":               path,
		"version":            header.Version,
		"stub_size":          header.StubSize,
		"stub_architecture":  header.StubArchitecture.String(),
		"stub_arch_version":  header.StubArchVersion.String(header.StubArchitecture),
		"num_binaries":       header.NumBinaries,
		"compression":        header.CompressionAlgo.String(),
		"binaries":           formatMetadataForJSON(metadata),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(path string, header *format.FormatHeader, metadata []*format.BinaryMetadata) error {
	data := map[string]interface{}{
		"path":               path,
		"version":            header.Version,
		"stub_size":          header.StubSize,
		"stub_architecture":  header.StubArchitecture.String(),
		"stub_arch_version":  header.StubArchVersion.String(header.StubArchitecture),
		"num_binaries":       header.NumBinaries,
		"compression":        header.CompressionAlgo.String(),
		"binaries":           formatMetadataForJSON(metadata),
	}

	encoder := yaml.NewEncoder(os.Stdout)
	return encoder.Encode(data)
}

func formatMetadataForJSON(metadata []*format.BinaryMetadata) []map[string]interface{} {
	result := make([]map[string]interface{}, len(metadata))

	for i, meta := range metadata {
		var features []string
		switch meta.Architecture {
		case format.ArchX86_64:
			features = cpu.FormatX86Features(meta.RequiredFeatures)
		case format.ArchARM64:
			features = cpu.FormatARMFeatures(meta.RequiredFeatures)
		}

		libraryPath := meta.GetLibraryPath()

		result[i] = map[string]interface{}{
			"index":             i,
			"architecture":      meta.Architecture.String(),
			"version":           meta.ArchVersion.String(meta.Architecture),
			"features":          features,
			"features_mask":     fmt.Sprintf("0x%x", meta.RequiredFeatures),
			"original_size":     meta.OriginalSize,
			"compressed_size":   meta.CompressedSize,
			"compression":       meta.Compression.String(),
			"compression_ratio": float64(meta.CompressedSize) / float64(meta.OriginalSize) * 100,
			"priority":          meta.Priority,
			"library_path":      libraryPath,
			"metadata_version":  meta.MetadataVersion,
		}
	}

	return result
}

func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
}
