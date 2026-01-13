// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-2026 Datadog, Inc.


package hwcaps

// GetDefaultTemplates returns standard library path templates
// for cross-distribution compatibility.
// Templates are evaluated at runtime, not build time.
//
// Template variable meanings:
//
//	{{.Arch}}        = "x86-64" or "aarch64"
//	{{.ArchTriple}}  = "x86_64" or "aarch64" (for Debian multiarch)
//	{{.Version}}     = "v3" or "v8.2" or "v9" (version string)
//	{{.ArchVersion}} = "x86-64-v3" or "aarch64-v8.2" (full identifier)
func GetDefaultTemplates() []string {
	return []string{
		// Debian/Ubuntu multiarch with fallback version support
		// Supports both exact (v9.4) and fuzzy (v9) matching for ARM64
		"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",

		// RedHat/Fedora lib64
		"/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",

		// Custom /opt installations
		"/opt/{{.Arch}}/lib",
	}
}
