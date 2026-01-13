package main

import (
	"fmt"
	"os"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/extractor"
	"github.com/DataDog/adipo/internal/hwcaps"
	"github.com/DataDog/adipo/internal/runner"
	"github.com/spf13/cobra"
)

var flags struct {
	libPathTemplate string
	includeStandard bool
	includeOpt      bool
	dryRun          bool
	verbose         bool
	scanDirs        []string
}

var rootCmd = &cobra.Command{
	Use:   "hwcaps-exec [flags] program [args...]",
	Short: "Execute a program with LD_LIBRARY_PATH configured based on CPU capabilities",
	Long: `Execute a program with LD_LIBRARY_PATH/DYLD_LIBRARY_PATH configured based on CPU capabilities.

This command replicates glibc hwcaps functionality for platforms without native support.
It detects CPU capabilities, scans for compatible library directories, and executes the
specified program with the appropriate library paths.

Directory Scanning (enabled by default):
  - Standard glibc-hwcaps paths: /usr/lib64/glibc-hwcaps/<arch-version>
  - Custom /opt paths: /opt/<arch>/lib
  - User-defined templates with {{.Arch}}, {{.Version}}, {{.ArchVersion}}
  - Additional directories via --scan-dir flag`,
	Example: `  # Auto-detect CPU and use standard paths
  hwcaps-exec myprogram arg1 arg2

  # Preview what would be executed (dry run)
  hwcaps-exec --dry-run myprogram

  # Use custom template
  hwcaps-exec --lib-path-template "/custom/{{.ArchVersion}}/lib" myprogram

  # Add additional directories
  hwcaps-exec --scan-dir /opt/mylibs myprogram

  # Verbose mode shows scanning process
  hwcaps-exec --verbose myprogram`,
	Args:              cobra.MinimumNArgs(1),
	RunE:              runHwcapsExec,
	DisableFlagParsing: false,
	SilenceUsage:      true,
	SilenceErrors:     true,
}

func init() {
	rootCmd.Flags().StringVar(&flags.libPathTemplate, "lib-path-template", "",
		"Template with {{.Arch}}, {{.Version}}, {{.ArchVersion}} variables")
	rootCmd.Flags().BoolVar(&flags.includeStandard, "include-standard-hwcaps", true,
		"Scan standard glibc-hwcaps directories")
	rootCmd.Flags().BoolVar(&flags.includeOpt, "include-opt-pattern", true,
		"Scan /opt/<arch>/lib directories")
	rootCmd.Flags().BoolVar(&flags.dryRun, "dry-run", false,
		"Show LD_LIBRARY_PATH without executing")
	rootCmd.Flags().BoolVar(&flags.verbose, "verbose", false,
		"Show detailed scanning process")
	rootCmd.Flags().StringSliceVar(&flags.scanDirs, "scan-dir", []string{},
		"Additional directories to scan (can be repeated)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "hwcaps-exec: %v\n", err)
		os.Exit(1)
	}
}

func runHwcapsExec(cmd *cobra.Command, args []string) error {
	program := args[0]
	programArgs := args[1:]

	// 1. Detect CPU capabilities
	if flags.verbose {
		fmt.Fprintf(os.Stderr, "Detecting CPU capabilities...\n")
	}

	caps, err := cpu.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect CPU: %w", err)
	}

	if flags.verbose {
		fmt.Fprintf(os.Stderr, "CPU: %s %s\n", caps.Architecture, caps.VersionStr)
	}

	// 2. Build scan configuration
	config := &hwcaps.ScanConfig{
		Capabilities:          caps,
		Templates:             []string{},
		ScanDirs:              flags.scanDirs,
		IncludeStandardHwcaps: flags.includeStandard,
		IncludeOptPattern:     flags.includeOpt,
	}

	if flags.libPathTemplate != "" {
		config.Templates = append(config.Templates, flags.libPathTemplate)
	}

	// 3. Scan directories
	if flags.verbose {
		fmt.Fprintf(os.Stderr, "\nScanning for library directories...\n")
	}

	results := hwcaps.ScanDirectories(config)

	// 4. Select compatible paths
	selected := hwcaps.SelectCompatiblePaths(results)
	libraryPath := hwcaps.BuildLibraryPath(selected)

	// 5. Verbose output
	if flags.verbose {
		printScanResults(results, selected, libraryPath)
	}

	// 6. Get library path environment variable name
	libEnvVar := runner.GetLibraryPathEnvVar()

	// 7. Dry run mode
	if flags.dryRun {
		if libraryPath != "" {
			fmt.Printf("%s=%s\n", libEnvVar, libraryPath)
		} else {
			fmt.Printf("%s= (no compatible libraries found)\n", libEnvVar)
		}
		fmt.Printf("[would execute: %s", program)
		for _, arg := range programArgs {
			fmt.Printf(" %s", arg)
		}
		fmt.Printf("]\n")
		return nil
	}

	// 8. Prepare environment
	env := extractor.GetEnvironment()
	overrides := make(map[string]string)

	if libraryPath != "" {
		overrides[libEnvVar] = runner.PrependLibraryPath(env, libEnvVar, libraryPath)
		if flags.verbose {
			fmt.Fprintf(os.Stderr, "\nSetting %s=%s\n", libEnvVar, overrides[libEnvVar])
		}
	} else if flags.verbose {
		fmt.Fprintf(os.Stderr, "\nNo compatible libraries found, executing with system defaults\n")
	}

	modifiedEnv := extractor.SetupEnvironment(env, overrides)

	// 9. Execute program
	if flags.verbose {
		fmt.Fprintf(os.Stderr, "Executing: %s", program)
		for _, arg := range programArgs {
			fmt.Fprintf(os.Stderr, " %s", arg)
		}
		fmt.Fprintf(os.Stderr, "\n\n")
	}

	if err := extractor.Execute(program, programArgs, modifiedEnv); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

func printScanResults(results []hwcaps.ScanResult, selected []hwcaps.ScanResult, libraryPath string) {
	fmt.Fprintf(os.Stderr, "\nScanned directories:\n")

	for _, result := range results {
		status := "✗ missing"
		if result.Exists {
			if result.IsCompatible {
				status = "✓ compatible"
			} else {
				status = "✗ incompatible"
			}
		}

		fmt.Fprintf(os.Stderr, "  [%s] %s (source: %s, priority: %d)\n",
			status, result.Path, result.Source, result.Priority)
	}

	fmt.Fprintf(os.Stderr, "\nSelected paths (%d):\n", len(selected))
	if len(selected) == 0 {
		fmt.Fprintf(os.Stderr, "  (none - executing with system defaults)\n")
	} else {
		for _, result := range selected {
			fmt.Fprintf(os.Stderr, "  %s\n", result.Path)
		}
	}

	libEnvVar := runner.GetLibraryPathEnvVar()
	fmt.Fprintf(os.Stderr, "\nFinal %s:\n", libEnvVar)
	if libraryPath != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", libraryPath)
	} else {
		fmt.Fprintf(os.Stderr, "  (not set)\n")
	}
}
