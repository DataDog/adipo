package main

import (
	"fmt"
	"os"

	"github.com/corentin-chary/adipo/internal/compression"
	"github.com/corentin-chary/adipo/internal/cpu"
	"github.com/corentin-chary/adipo/internal/extractor"
	"github.com/corentin-chary/adipo/internal/format"
	"github.com/corentin-chary/adipo/internal/selector"
	"github.com/spf13/cobra"
)

var runFlags struct {
	force       string
	verbose     bool
	dryRun      bool
	preferDisk  bool
}

var runCmd = &cobra.Command{
	Use:   "run [flags] fat-binary [args...]",
	Short: "Run a fat binary",
	Long: `Run a fat binary by detecting the current CPU and executing the best matching binary.

Note: Fat binaries are self-executing, so you can also just run them directly:
  ./app.fat [args...]

This command is useful for debugging or forcing a specific binary version.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVar(&runFlags.force, "force", "", "Force specific architecture (e.g., x86-64-v2)")
	runCmd.Flags().BoolVarP(&runFlags.verbose, "verbose", "v", false, "Verbose output")
	runCmd.Flags().BoolVar(&runFlags.dryRun, "dry-run", false, "Show what would be executed without running")
	runCmd.Flags().BoolVar(&runFlags.preferDisk, "prefer-disk", false, "Prefer disk extraction over memory")
}

func runRun(cmd *cobra.Command, args []string) error {
	fatBinary := args[0]
	binaryArgs := args[1:]

	// Open the fat binary
	reader, err := format.OpenFile(fatBinary)
	if err != nil {
		return fmt.Errorf("failed to open fat binary: %w", err)
	}

	metadata := reader.Metadata()

	if runFlags.verbose {
		fmt.Fprintf(os.Stderr, "Found %d binaries in fat binary\n", len(metadata))
	}

	// Detect CPU (or use forced spec)
	var caps *cpu.Capabilities
	if runFlags.force != "" {
		if runFlags.verbose {
			fmt.Fprintf(os.Stderr, "Forced specification: %s\n", runFlags.force)
		}
		// TODO: Parse forced spec and create custom capabilities
		// For now, still detect
		caps, err = cpu.Detect()
		if err != nil {
			return fmt.Errorf("failed to detect CPU: %w", err)
		}
	} else {
		caps, err = cpu.Detect()
		if err != nil {
			return fmt.Errorf("failed to detect CPU: %w", err)
		}

		if runFlags.verbose {
			fmt.Fprintf(os.Stderr, "Detected CPU: %s\n", caps.String())
		}
	}

	// Select binary
	sel := selector.NewSelector(caps, metadata)
	result, err := sel.SelectBinaryVerbose()
	if err != nil {
		return fmt.Errorf("no compatible binary found: %w", err)
	}

	if runFlags.verbose {
		fmt.Fprintf(os.Stderr, "Selected binary %d: %s-%s (score: %d)\n",
			result.SelectedIndex,
			result.SelectedBinary.Architecture.String(),
			result.SelectedBinary.ArchVersion.String(result.SelectedBinary.Architecture),
			result.SelectedScore,
		)
	}

	// Dry run - just show what would be executed
	if runFlags.dryRun {
		fmt.Printf("Would execute: binary %d (%s-%s)\n",
			result.SelectedIndex,
			result.SelectedBinary.Architecture.String(),
			result.SelectedBinary.ArchVersion.String(result.SelectedBinary.Architecture),
		)
		fmt.Printf("Arguments: %v\n", binaryArgs)
		return nil
	}

	// Read and decompress the binary
	if runFlags.verbose {
		fmt.Fprintf(os.Stderr, "Reading binary data...\n")
	}

	compressedData, err := reader.GetBinaryData(result.SelectedIndex)
	if err != nil {
		return fmt.Errorf("failed to read binary data: %w", err)
	}

	if runFlags.verbose {
		fmt.Fprintf(os.Stderr, "Decompressing (%s)...\n", result.SelectedBinary.Compression.String())
	}

	decompressedData, err := compression.Decompress(
		compressedData,
		result.SelectedBinary.Compression,
		result.SelectedBinary.OriginalSize,
	)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	// Execute
	opts := &extractor.ExecutionOptions{
		Args:         binaryArgs,
		Env:          extractor.GetEnvironment(),
		PreferMemory: !runFlags.preferDisk,
		Verbose:      runFlags.verbose,
	}

	if runFlags.verbose {
		fmt.Fprintf(os.Stderr, "Executing binary...\n")
	}

	err = extractor.ExtractAndExecute(decompressedData, "binary", opts)
	if err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	// This line should never be reached
	return fmt.Errorf("exec returned unexpectedly")
}
