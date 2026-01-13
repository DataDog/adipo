package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/extractor"
	"github.com/DataDog/adipo/internal/hwcaps"
	"github.com/DataDog/adipo/internal/runner"
)

var (
	libPathTemplate = flag.String("lib-path-template", "", "Template with {{.Arch}}, {{.Version}}, {{.ArchVersion}} variables")
	includeStandard = flag.Bool("include-standard-hwcaps", true, "Scan standard glibc-hwcaps directories")
	includeOpt      = flag.Bool("include-opt-pattern", true, "Scan /opt/<arch>/lib directories")
	dryRun          = flag.Bool("dry-run", false, "Show LD_LIBRARY_PATH without executing")
	verbose         = flag.Bool("verbose", false, "Show detailed scanning process")
)

// Custom flag type for repeated --scan-dir flags
type scanDirsFlag []string

func (s *scanDirsFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *scanDirsFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var scanDirs scanDirsFlag

func main() {
	flag.Var(&scanDirs, "scan-dir", "Additional directories to scan (can be repeated)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: hwcaps-exec [flags] program [args...]\n\n")
		fmt.Fprintf(os.Stderr, "Execute a program with LD_LIBRARY_PATH/DYLD_LIBRARY_PATH configured based on CPU capabilities.\n\n")
		fmt.Fprintf(os.Stderr, "This command replicates glibc hwcaps functionality for platforms without native support.\n")
		fmt.Fprintf(os.Stderr, "It detects CPU capabilities, scans for compatible library directories, and executes the\n")
		fmt.Fprintf(os.Stderr, "specified program with the appropriate library paths.\n\n")
		fmt.Fprintf(os.Stderr, "Directory Scanning (enabled by default):\n")
		fmt.Fprintf(os.Stderr, "  - Standard glibc-hwcaps paths: /usr/lib64/glibc-hwcaps/<arch-version>\n")
		fmt.Fprintf(os.Stderr, "  - Custom /opt paths: /opt/<arch>/lib\n")
		fmt.Fprintf(os.Stderr, "  - User-defined templates with {{.Arch}}, {{.Version}}, {{.ArchVersion}}\n")
		fmt.Fprintf(os.Stderr, "  - Additional directories via --scan-dir flag\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Auto-detect CPU and use standard paths\n")
		fmt.Fprintf(os.Stderr, "  hwcaps-exec myprogram arg1 arg2\n\n")
		fmt.Fprintf(os.Stderr, "  # Preview what would be executed (dry run)\n")
		fmt.Fprintf(os.Stderr, "  hwcaps-exec --dry-run myprogram\n\n")
		fmt.Fprintf(os.Stderr, "  # Use custom template\n")
		fmt.Fprintf(os.Stderr, "  hwcaps-exec --lib-path-template \"/custom/{{.ArchVersion}}/lib\" myprogram\n\n")
		fmt.Fprintf(os.Stderr, "  # Add additional directories\n")
		fmt.Fprintf(os.Stderr, "  hwcaps-exec --scan-dir /opt/mylibs myprogram\n\n")
		fmt.Fprintf(os.Stderr, "  # Verbose mode shows scanning process\n")
		fmt.Fprintf(os.Stderr, "  hwcaps-exec --verbose myprogram\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		fatal("error: program argument required")
	}

	program := flag.Arg(0)
	programArgs := flag.Args()[1:]

	// 1. Detect CPU capabilities
	if *verbose {
		fmt.Fprintf(os.Stderr, "Detecting CPU capabilities...\n")
	}

	caps, err := cpu.Detect()
	if err != nil {
		fatal("failed to detect CPU: %v", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "CPU: %s %s\n", caps.Architecture, caps.VersionStr)
	}

	// 2. Build scan configuration
	config := &hwcaps.ScanConfig{
		Capabilities:          caps,
		Templates:             []string{},
		ScanDirs:              []string(scanDirs),
		IncludeStandardHwcaps: *includeStandard,
		IncludeOptPattern:     *includeOpt,
	}

	if *libPathTemplate != "" {
		config.Templates = append(config.Templates, *libPathTemplate)
	}

	// 3. Scan directories
	if *verbose {
		fmt.Fprintf(os.Stderr, "\nScanning for library directories...\n")
	}

	results := hwcaps.ScanDirectories(config)

	// 4. Select compatible paths
	selected := hwcaps.SelectCompatiblePaths(results)
	libraryPath := hwcaps.BuildLibraryPath(selected)

	// 5. Verbose output
	if *verbose {
		printScanResults(results, selected, libraryPath)
	}

	// 6. Get library path environment variable name
	libEnvVar := runner.GetLibraryPathEnvVar()

	// 7. Dry run mode
	if *dryRun {
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
		os.Exit(0)
	}

	// 8. Prepare environment
	env := extractor.GetEnvironment()
	overrides := make(map[string]string)

	if libraryPath != "" {
		overrides[libEnvVar] = runner.PrependLibraryPath(env, libEnvVar, libraryPath)
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nSetting %s=%s\n", libEnvVar, overrides[libEnvVar])
		}
	} else if *verbose {
		fmt.Fprintf(os.Stderr, "\nNo compatible libraries found, executing with system defaults\n")
	}

	modifiedEnv := extractor.SetupEnvironment(env, overrides)

	// 9. Execute program
	if *verbose {
		fmt.Fprintf(os.Stderr, "Executing: %s", program)
		for _, arg := range programArgs {
			fmt.Fprintf(os.Stderr, " %s", arg)
		}
		fmt.Fprintf(os.Stderr, "\n\n")
	}

	if err := extractor.Execute(program, programArgs, modifiedEnv); err != nil {
		fatal("execution failed: %v", err)
	}
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

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "hwcaps-exec: "+format+"\n", args...)
	os.Exit(1)
}
