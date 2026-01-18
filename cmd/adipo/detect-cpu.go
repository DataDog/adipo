// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/DataDog/adipo/internal/cpu"
	"github.com/spf13/cobra"
)

var detectCPUFlags struct {
	jsonOutput bool
}

var detectCPUCmd = &cobra.Command{
	Use:   "detect-cpu",
	Short: "Detect current CPU capabilities and model",
	Long: `Detect and display information about the current CPU including architecture,
version, model details, and supported features.

This command is useful for:
- Testing CPU detection before building fat binaries
- Understanding what CPU features are available on the current system
- Debugging binary selection issues

Output includes:
- Architecture (x86-64, aarch64)
- Microarchitecture version (v1-v4 for x86-64, v8.x/v9.x for ARM64)
- CPU model information (vendor, family, model, or implementer/part)
- Detected features (instruction set extensions)

Use --json for machine-readable output.`,
	RunE: runDetectCPU,
}

func init() {
	detectCPUCmd.Flags().BoolVar(&detectCPUFlags.jsonOutput, "json", false, "Output in JSON format")
}

func runDetectCPU(cmd *cobra.Command, args []string) error {
	// Detect CPU capabilities
	caps, err := cpu.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect CPU: %w", err)
	}

	if detectCPUFlags.jsonOutput {
		return outputDetectCPUJSON(caps)
	}

	return outputDetectCPUHuman(caps)
}

// CPUInfo represents CPU information for JSON output
type CPUInfo struct {
	Architecture string            `json:"architecture"`
	Version      string            `json:"version"`
	CPUAlias     string            `json:"cpu_alias,omitempty"`
	CPUModel     *CPUModelInfo     `json:"cpu_model,omitempty"`
	Features     []string          `json:"features"`
}

// CPUModelInfo represents CPU model details for JSON output
type CPUModelInfo struct {
	// x86-64 fields
	Vendor    string `json:"vendor,omitempty"`
	ModelName string `json:"model_name,omitempty"`
	Family    int    `json:"family,omitempty"`
	Model     int    `json:"model,omitempty"`

	// ARM64 fields
	Implementer int `json:"implementer,omitempty"`
	PartNum     int `json:"part_num,omitempty"`

	// macOS fields
	BrandString string `json:"brand_string,omitempty"`
}

func outputDetectCPUJSON(caps *cpu.Capabilities) error {
	info := CPUInfo{
		Architecture: caps.Architecture,
		Version:      caps.VersionStr,
		Features:     caps.FeatureList(),
	}

	// Detect CPU alias
	if caps.CPUModel != nil {
		alias := cpu.DetectCPUAlias(caps.CPUModel, caps.ArchType)
		if alias != "" {
			info.CPUAlias = alias
		}
	}

	// Add CPU model information if available
	if caps.CPUModel != nil {
		info.CPUModel = &CPUModelInfo{
			Vendor:      caps.CPUModel.Vendor,
			ModelName:   caps.CPUModel.ModelName,
			Family:      caps.CPUModel.Family,
			Model:       caps.CPUModel.Model,
			Implementer: caps.CPUModel.Implementer,
			PartNum:     caps.CPUModel.PartNum,
			BrandString: caps.CPUModel.BrandString,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

func outputDetectCPUHuman(caps *cpu.Capabilities) error {
	fmt.Printf("Architecture: %s\n", caps.Architecture)
	fmt.Printf("Version: %s\n", caps.VersionStr)

	// Display CPU model information if available
	if caps.CPUModel != nil {
		fmt.Printf("CPU Model: %s\n", caps.CPUModel.String())

		// Detect and display CPU alias
		alias := cpu.DetectCPUAlias(caps.CPUModel, caps.ArchType)
		if alias != "" {
			fmt.Printf("CPU Alias: %s\n", alias)
		}

		// Show detailed fields based on architecture
		switch runtime.GOARCH {
		case "amd64", "386":
			if caps.CPUModel.Vendor != "" {
				fmt.Printf("Vendor: %s\n", caps.CPUModel.Vendor)
			}
			if caps.CPUModel.Family != 0 {
				fmt.Printf("Family: %d\n", caps.CPUModel.Family)
			}
			if caps.CPUModel.Model != 0 {
				fmt.Printf("Model: %d\n", caps.CPUModel.Model)
			}

		case "arm64", "arm":
			if runtime.GOOS == "linux" {
				if caps.CPUModel.Implementer != 0 {
					fmt.Printf("Implementer: 0x%x\n", caps.CPUModel.Implementer)
				}
				if caps.CPUModel.PartNum != 0 {
					fmt.Printf("Part: 0x%x\n", caps.CPUModel.PartNum)
				}
			}
		}
	}

	// Display features
	features := caps.FeatureList()
	sort.Strings(features)
	fmt.Printf("Features: %s\n", formatFeatureList(features))

	return nil
}

// formatFeatureList formats the feature list for display
func formatFeatureList(features []string) string {
	if len(features) == 0 {
		return "(none)"
	}

	// Show first 10 features, then "... and N more"
	const maxShow = 10
	if len(features) <= maxShow {
		return fmt.Sprintf("%s", features)
	}

	shown := features[:maxShow]
	remaining := len(features) - maxShow
	return fmt.Sprintf("%s ... and %d more", shown, remaining)
}
