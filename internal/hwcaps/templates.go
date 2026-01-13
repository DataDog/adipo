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
		// Debian/Ubuntu multiarch (priority: exact version)
		"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.ArchVersion}}",

		// Debian/Ubuntu multiarch (fuzzy: major version only, for ARM64 v9 → v9.x)
		"/usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}",

		// RedHat/Fedora lib64
		"/usr/lib64/glibc-hwcaps/{{.ArchVersion}}",

		// Generic lib
		"/usr/lib/glibc-hwcaps/{{.ArchVersion}}",

		// Custom /opt installations
		"/opt/{{.Arch}}/lib",
	}
}
