// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// TODO: Future self-stub mode support could be added here.
	// When invoked with a non-standard name (not 'adipo' or 'adipo-*'),
	// the binary could detect it's embedded in a fat binary and act as a stub.
	// Challenges to solve:
	// - adipo binary contains "ADIPOFAT" magic marker constant that interferes with detection
	// - Need robust magic marker search that skips false positives in code/data sections
	// Alternative: Consider robust stub embedding or separate stub binaries (current approach)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "adipo",
	Short: "Create and run fat binaries based on CPU micro-architecture",
	Long: `adipo creates and runs fat binaries containing multiple versions of the same executable,
optimized for different CPU micro-architectures (x86-64 v1/v2/v3/v4, ARM64 v8/v9).

At runtime, adipo automatically selects and executes the best binary for the current CPU.`,
	Version: getVersionString(),
}

func getVersionString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(detectCPUCmd)
}
