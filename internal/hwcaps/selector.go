// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package hwcaps

import (
	"sort"
	"strings"
)

// SelectCompatiblePaths filters and sorts compatible paths
func SelectCompatiblePaths(results []ScanResult) []ScanResult {
	var selected []ScanResult

	// Filter to only compatible directories that exist
	for _, result := range results {
		if result.Exists && result.IsCompatible {
			selected = append(selected, result)
		}
	}

	// Sort by priority (highest first)
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Priority > selected[j].Priority
	})

	return selected
}

// BuildLibraryPath constructs colon-separated LD_LIBRARY_PATH from selected results
func BuildLibraryPath(selected []ScanResult) string {
	if len(selected) == 0 {
		return ""
	}

	var paths []string
	seen := make(map[string]bool)

	for _, result := range selected {
		// Avoid duplicates
		if !seen[result.Path] {
			paths = append(paths, result.Path)
			seen[result.Path] = true
		}
	}

	return strings.Join(paths, ":")
}

// SortByPriority sorts results by priority in descending order
func SortByPriority(results []ScanResult) []ScanResult {
	sorted := make([]ScanResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	return sorted
}
