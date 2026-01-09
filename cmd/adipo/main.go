package main

import (
	"fmt"
	"os"

	"github.com/DataDog/adipo/internal/stub"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
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
	baseVersion := fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	stubSize := stub.StubSize()
	if stubSize > 0 {
		return fmt.Sprintf("%s, stub: embedded (%d bytes)", baseVersion, stubSize)
	}
	return fmt.Sprintf("%s, stub: not embedded", baseVersion)
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(runCmd)
}
