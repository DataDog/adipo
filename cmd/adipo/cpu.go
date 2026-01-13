// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package main

import (
	"fmt"
	"sort"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/DataDog/adipo/internal/format"
	"github.com/spf13/cobra"
)

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Show current CPU architecture and features",
	Long: `Detects and displays the current CPU architecture, version level,
and available CPU features. This helps understand which binaries
from a fat binary would be compatible with this system.`,
	RunE: runCPU,
}

func init() {
	rootCmd.AddCommand(cpuCmd)
}

func runCPU(cmd *cobra.Command, args []string) error {
	// Detect CPU capabilities
	caps, err := cpu.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect CPU: %w", err)
	}

	// Print architecture info
	fmt.Printf("CPU Architecture: %s\n", caps.Architecture)
	fmt.Printf("Architecture Type: %s\n", caps.ArchType)
	fmt.Printf("Version Level: %s\n", caps.Version.String(caps.ArchType))
	fmt.Printf("Feature Mask: 0x%016x\n\n", caps.FeatureMask)

	// Print features in a nice format
	fmt.Println("Detected CPU Features:")

	// Collect and sort feature names
	var featureNames []string
	for name := range caps.Features {
		featureNames = append(featureNames, name)
	}
	sort.Strings(featureNames)

	// Print features based on architecture
	if len(featureNames) == 0 {
		fmt.Println("  (baseline features only)")
	} else {
		// Group features by category for better readability
		printFeaturesByCategory(caps.ArchType, featureNames)
	}

	return nil
}

func printFeaturesByCategory(arch format.Architecture, featureNames []string) {
	// For x86-64, group by categories
	if arch == format.ArchX86_64 {
		categories := map[string][]string{
			"Vector Extensions": {},
			"Other Extensions":  {},
		}

		vectorFeatures := map[string]bool{
			"sse3": true, "ssse3": true, "sse4.1": true, "sse4.2": true,
			"avx": true, "avx2": true, "avx512f": true, "avx512dq": true,
			"avx512cd": true, "avx512bw": true, "avx512vl": true,
		}

		for _, name := range featureNames {
			if vectorFeatures[name] {
				categories["Vector Extensions"] = append(categories["Vector Extensions"], name)
			} else {
				categories["Other Extensions"] = append(categories["Other Extensions"], name)
			}
		}

		// Print categorized features
		for _, category := range []string{"Vector Extensions", "Other Extensions"} {
			if len(categories[category]) > 0 {
				fmt.Printf("\n  %s:\n", category)
				for _, name := range categories[category] {
					fmt.Printf("    - %s\n", name)
				}
			}
		}
	} else {
		// For ARM64, just list all features
		for _, name := range featureNames {
			fmt.Printf("  - %s\n", name)
		}
	}
}
