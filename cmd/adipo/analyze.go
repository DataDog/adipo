// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DataDog/adipo/internal/analyzer"
	"github.com/DataDog/adipo/internal/compression"
	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
	"github.com/DataDog/adipo/internal/selector"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var analyzeFlags struct {
	short       bool
	statistics  bool
	formatType  string
	target      string
	objdumpPath string
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [flags] binary",
	Short: "Analyze a binary's instruction set requirements",
	Long: `Analyzes a binary to detect which CPU features are actually used
by disassembling and examining the instruction set. Works on both
regular binaries and fat binaries.

For fat binaries, use --target to select which binary to analyze:
  --target auto        Auto-select based on current CPU (default)
  --target 0           Analyze first binary
  --target x86-64-v3   Analyze binary matching specification

Examples:
  adipo analyze myapp
  adipo analyze --short myapp
  adipo analyze --statistics myapp
  adipo analyze --target x86-64-v3 myapp.fat`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().BoolVar(&analyzeFlags.short, "short", false,
		"Output in GCC march format only")
	analyzeCmd.Flags().BoolVar(&analyzeFlags.statistics, "statistics", false,
		"Show instruction counts per feature group")
	analyzeCmd.Flags().StringVar(&analyzeFlags.formatType, "format", "table",
		"Output format (table, json, yaml)")
	analyzeCmd.Flags().StringVar(&analyzeFlags.target, "target", "auto",
		"For fat binaries: auto, index, or arch spec")
	analyzeCmd.Flags().StringVar(&analyzeFlags.objdumpPath, "objdump", "",
		"Path to objdump binary (auto-detect if empty)")

	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	binaryPath := args[0]

	// Check if fat binary
	isFat, err := analyzer.IsFatBinary(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to check binary type: %w", err)
	}

	var analysisTarget string
	var cleanup func()

	if isFat {
		// Handle fat binary: extract target binary to temp file
		target, cleanupFn, err := extractTargetBinary(binaryPath, analyzeFlags.target)
		if err != nil {
			return err
		}
		analysisTarget = target
		cleanup = cleanupFn
		defer cleanup()
	} else {
		analysisTarget = binaryPath
	}

	// Create analyzer
	config := analyzer.Config{
		ObjdumpPath:       analyzeFlags.objdumpPath,
		CollectStatistics: analyzeFlags.statistics,
	}

	a, err := analyzer.New(config)
	if err != nil {
		return err
	}

	// Analyze
	result, err := a.Analyze(analysisTarget)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Output
	if analyzeFlags.short {
		return outputAnalysisShort(result)
	}

	switch analyzeFlags.formatType {
	case "table":
		return outputAnalysisTable(result)
	case "json":
		return outputAnalysisJSON(result)
	case "yaml":
		return outputAnalysisYAML(result)
	default:
		return fmt.Errorf("unknown format: %s", analyzeFlags.formatType)
	}
}

// extractTargetBinary extracts a binary from a fat binary to a temp file for analysis
func extractTargetBinary(fatBinaryPath string, target string) (string, func(), error) {
	// Open the fat binary
	reader, err := format.OpenFile(fatBinaryPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open fat binary: %w", err)
	}
	defer func() { _ = reader.Close() }()

	metadata := reader.Metadata()

	// Select target binary
	index, err := selectAnalysisTarget(target, metadata)
	if err != nil {
		return "", nil, err
	}

	fmt.Fprintf(os.Stderr, "Analyzing binary %d: %s-%s\n",
		index,
		metadata[index].Architecture.String(),
		metadata[index].ArchVersion.String(metadata[index].Architecture))

	// Read compressed data
	compressedData, err := reader.GetBinaryData(index)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read binary data: %w", err)
	}

	// Decompress
	var binaryData []byte
	if metadata[index].Compression == format.CompressionNone {
		binaryData = compressedData
	} else {
		binaryData, err = compression.Decompress(compressedData, metadata[index].Compression, metadata[index].OriginalSize)
		if err != nil {
			return "", nil, fmt.Errorf("failed to decompress binary: %w", err)
		}
	}

	// Write to temp file
	tempFile, err := os.CreateTemp("", "adipo-analyze-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tempPath := tempFile.Name()

	if _, err := tempFile.Write(binaryData); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Make it executable
	if err := os.Chmod(tempPath, 0755); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, fmt.Errorf("failed to make temp file executable: %w", err)
	}

	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	return tempPath, cleanup, nil
}

// selectAnalysisTarget selects which binary to analyze from a fat binary
func selectAnalysisTarget(target string, metadata []*format.BinaryMetadata) (int, error) {
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

// outputAnalysisShort outputs just the GCC march string
func outputAnalysisShort(result *analyzer.AnalysisResult) error {
	gen := analyzer.NewMarchGenerator(result.Architecture)
	march := gen.Generate(result.DetectedVersion, result.DetectedFeatures)
	fmt.Println(march)
	return nil
}

// outputAnalysisTable outputs in table format
func outputAnalysisTable(result *analyzer.AnalysisResult) error {
	fmt.Printf("Architecture: %s\n", result.Architecture.String())
	fmt.Printf("Detected Microarchitecture Level: %s\n",
		result.DetectedVersion.String(result.Architecture))

	if len(result.FeatureNames) > 0 {
		fmt.Printf("Detected Features: %s\n", strings.Join(result.FeatureNames, ", "))
	} else {
		fmt.Println("Detected Features: none (baseline)")
	}
	fmt.Println()

	if result.Statistics != nil {
		fmt.Print(result.Statistics.Format())
		fmt.Println()
	}

	fmt.Printf("Compatibility: %s\n", result.CompatibilityMsg)
	fmt.Println("\nNote: This analysis is based on static disassembly. The binary may")
	fmt.Println("contain runtime CPU detection and branching, which could allow it to")
	fmt.Println("run on older CPUs than detected here.")

	return nil
}

// outputAnalysisJSON outputs in JSON format
func outputAnalysisJSON(result *analyzer.AnalysisResult) error {
	data := map[string]interface{}{
		"architecture":     result.Architecture.String(),
		"detected_version": result.DetectedVersion.String(result.Architecture),
		"detected_features": result.FeatureNames,
		"feature_mask":     fmt.Sprintf("0x%x", result.DetectedFeatures),
		"compatibility": map[string]interface{}{
			"can_run": result.CanRun,
			"message": result.CompatibilityMsg,
		},
	}

	if result.Statistics != nil {
		data["statistics"] = map[string]interface{}{
			"total_instructions":    result.Statistics.TotalInstructions,
			"feature_group_counts": result.Statistics.FeatureGroupCounts,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputAnalysisYAML outputs in YAML format
func outputAnalysisYAML(result *analyzer.AnalysisResult) error {
	data := map[string]interface{}{
		"architecture":     result.Architecture.String(),
		"detected_version": result.DetectedVersion.String(result.Architecture),
		"detected_features": result.FeatureNames,
		"feature_mask":     fmt.Sprintf("0x%x", result.DetectedFeatures),
		"compatibility": map[string]interface{}{
			"can_run": result.CanRun,
			"message": result.CompatibilityMsg,
		},
	}

	if result.Statistics != nil {
		data["statistics"] = map[string]interface{}{
			"total_instructions":    result.Statistics.TotalInstructions,
			"feature_group_counts": result.Statistics.FeatureGroupCounts,
		}
	}

	encoder := yaml.NewEncoder(os.Stdout)
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(data)
}
