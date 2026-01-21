// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.

package extractor

import (
	"fmt"
	"strings"

	"github.com/DataDog/adipo/internal/format"
)

// ExpandTemplate replaces template variables with architecture-specific values
// Supports: {{.Arch}}, {{.ArchTriple}}, {{.Version}}, {{.ArchVersion}}
func ExpandTemplate(template string, arch format.Architecture, version format.ArchVersion) string {
	replacements := map[string]string{
		"{{.Arch}}":        getArchName(arch),
		"{{.ArchTriple}}":  getArchTriple(arch),
		"{{.Version}}":     getVersionStr(arch, version),
		"{{.ArchVersion}}": getArchVersionStr(arch, version),
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

func getArchName(arch format.Architecture) string {
	switch arch {
	case format.ArchX86_64:
		return "x86-64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

func getArchTriple(arch format.Architecture) string {
	switch arch {
	case format.ArchX86_64:
		return "x86_64"
	case format.ArchARM64:
		return "aarch64"
	default:
		return "unknown"
	}
}

func getVersionStr(arch format.Architecture, version format.ArchVersion) string {
	switch arch {
	case format.ArchX86_64:
		return fmt.Sprintf("v%d", version)
	case format.ArchARM64:
		return version.String(arch) // "v8.2", "v9.0"
	default:
		return "unknown"
	}
}

func getArchVersionStr(arch format.Architecture, version format.ArchVersion) string {
	return fmt.Sprintf("%s-%s", getArchName(arch), getVersionStr(arch, version))
}
