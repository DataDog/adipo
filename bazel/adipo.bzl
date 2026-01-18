"""Bazel rules for creating fat binaries with adipo"""

def _adipo_fat_binary_impl(ctx):
    """Implementation of the adipo_fat_binary rule."""

    # Get the adipo tool
    adipo_tool = ctx.executable._adipo

    # Prepare the output fat binary
    out = ctx.actions.declare_file(ctx.label.name)

    # Build the command arguments
    args = ctx.actions.args()
    args.add("create")
    args.add("-o", out.path)

    if ctx.attr.compression:
        args.add("--compress", ctx.attr.compression)

    # Add library path options
    if ctx.attr.lib_path_template:
        args.add("--lib-path-template", ctx.attr.lib_path_template)

    if ctx.attr.enable_lib_path:
        args.add("--enable-lib-path")

    # Add stub binary if provided
    inputs = []
    if ctx.attr.stub:
        stub_file = ctx.attr.stub[DefaultInfo].files.to_list()[0]
        inputs.append(stub_file)
        args.add("--stub-path", stub_file.path)
    elif ctx.attr.no_stub:
        args.add("--no-stub")

    # Add each binary with its specification
    for binary_target, spec in ctx.attr.binaries.items():
        binary_file = binary_target[DefaultInfo].files.to_list()[0]
        inputs.append(binary_file)
        args.add("--binary", "{}:{}".format(binary_file.path, spec))

    # Run the adipo create command
    ctx.actions.run(
        outputs = [out],
        inputs = inputs,
        executable = adipo_tool,
        arguments = [args],
        mnemonic = "AdipoCreate",
        progress_message = "Creating fat binary %s" % ctx.label.name,
    )

    return [DefaultInfo(
        files = depset([out]),
        executable = out,
    )]

adipo_fat_binary = rule(
    implementation = _adipo_fat_binary_impl,
    attrs = {
        "binaries": attr.label_keyed_string_dict(
            mandatory = True,
            allow_files = True,
            doc = "Dictionary mapping binary targets to their architecture specifications (e.g., 'x86-64-v1', 'x86-64-v2', 'x86-64-v3:zen3'). Optionally include CPU hint after architecture spec: 'ARCH-VERSION:CPU-HINT'",
        ),
        "compression": attr.string(
            default = "zstd",
            doc = "Compression algorithm: zstd (default), lz4, gzip, or none",
            values = ["zstd", "lz4", "gzip", "none"],
        ),
        "lib_path_template": attr.string(
            doc = "Library path template (e.g., '/opt/libs/{{.ArchVersion}}'). Template variables: {{.Arch}}, {{.ArchTriple}}, {{.Version}}, {{.ArchVersion}}, {{.CPUAlias}}",
        ),
        "enable_lib_path": attr.bool(
            default = False,
            doc = "Enable automatic library path configuration using default templates for cross-distribution compatibility",
        ),
        "stub": attr.label(
            allow_single_file = True,
            doc = "Optional stub binary to use for self-extraction. If not provided and no_stub is False, adipo will try auto-discovery.",
        ),
        "no_stub": attr.bool(
            default = False,
            doc = "If True, create fat binary without self-extracting stub (--no-stub flag)",
        ),
        "_adipo": attr.label(
            default = Label("//cmd/adipo"),
            executable = True,
            cfg = "exec",
            doc = "The adipo tool binary",
        ),
    },
    executable = True,
    doc = """
Creates a fat binary containing multiple architecture-optimized versions of a binary.

Example:
    adipo_fat_binary(
        name = "myapp_fat",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v2": "x86-64-v2",
            ":myapp_v3": "x86-64-v3",
        },
        compression = "zstd",
        stub = "//cmd/adipo-stub",  # Optional: specify stub binary
    )

    # Without self-extracting stub:
    adipo_fat_binary(
        name = "myapp_fat_nostub",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v2": "x86-64-v2",
        },
        no_stub = True,
    )

    # With CPU-specific optimization hints:
    adipo_fat_binary(
        name = "myapp_fat_cpuhints",
        binaries = {
            ":myapp_zen": "x86-64-v3:zen3",
            ":myapp_intel": "x86-64-v3:skylake",
            ":myapp_baseline": "x86-64-v1",
        },
        enable_lib_path = True,
        lib_path_template = "/opt/{{.CPUAlias}}/lib",
        # Runtime selects best binary based on detected CPU
        # Library paths prioritize CPU-specific directories when hint matches
    )

    # With automatic library path templates (default templates):
    adipo_fat_binary(
        name = "myapp_fat_autolib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v2": "x86-64-v2",
            ":myapp_v4": "x86-64-v4",
        },
        enable_lib_path = True,
        # Uses default templates for cross-distribution compatibility:
        # /usr/lib/{{.ArchTriple}}-linux-gnu/glibc-hwcaps/{{.Version}}
        # /usr/lib64/glibc-hwcaps/{{.ArchVersion}}
        # /opt/{{.Arch}}/lib
    )

    # With custom template library path:
    adipo_fat_binary(
        name = "myapp_fat_customlib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v3": "x86-64-v3",
        },
        lib_path_template = "/opt/glibc-{{.Version}}/lib",
        # Templates are evaluated at runtime to find existing paths
    )
""",
)

def _adipo_multi_arch_binary_impl(ctx):
    """Implementation of helper rule to build the same binary for multiple architectures."""

    # This is a helper macro that will be expanded at analysis time
    pass

def adipo_multi_arch_binary(
        name,
        binary,
        specs,
        goarch = None,
        goos = None,
        compression = "zstd",
        **kwargs):
    """
    Helper macro to build a single binary target for multiple architecture specifications
    and create a fat binary from them.

    This is useful when you want to build the same source with different compiler flags
    for different CPU micro-architectures.

    Args:
        name: Name of the output fat binary
        binary: The binary target to build for multiple architectures
        specs: List of architecture specifications (e.g., ["x86-64-v1", "x86-64-v2", "x86-64-v3"])
        goarch: Target GOARCH (optional, for Go binaries)
        goos: Target GOOS (optional, for Go binaries)
        compression: Compression algorithm (default: "zstd")
        **kwargs: Additional arguments passed to the adipo_fat_binary rule

    Example:
        adipo_multi_arch_binary(
            name = "myapp_fat",
            binary = ":myapp",
            specs = ["x86-64-v1", "x86-64-v2", "x86-64-v3"],
            goos = "linux",
            goarch = "amd64",
        )
    """

    # For now, this requires the user to create the individual binaries
    # In a more advanced implementation, we could automatically create
    # multiple build configurations

    fail("adipo_multi_arch_binary is a placeholder. Please create individual binary targets with different configurations and use adipo_fat_binary directly.")
