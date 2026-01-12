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
    if ctx.attr.lib_path:
        args.add("--lib-path", ctx.attr.lib_path)

    if ctx.attr.auto_lib_path:
        args.add("--auto-lib-path", ctx.attr.auto_lib_path)

    if ctx.attr.enable_auto_lib:
        args.add("--enable-auto-lib")

    # Add stub binary if provided
    inputs = []
    if ctx.attr.stub:
        stub_file = ctx.attr.stub[DefaultInfo].files.to_list()[0]
        inputs.append(stub_file)
        args.add("--stub-path", stub_file.path)
    elif ctx.attr.no_stub:
        args.add("--no-stub")

    # Add each binary with its specification and optional library path
    for binary_target, spec in ctx.attr.binaries.items():
        binary_file = binary_target[DefaultInfo].files.to_list()[0]
        inputs.append(binary_file)
        args.add("--binary", "{}:{}".format(binary_file.path, spec))

        # Add per-binary library path if specified
        binary_name = binary_file.basename
        if binary_name in ctx.attr.binary_libs:
            args.add("--binary-lib", "{}:{}".format(binary_name, ctx.attr.binary_libs[binary_name]))

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
            doc = "Dictionary mapping binary targets to their architecture specifications (e.g., 'x86-64-v1', 'x86-64-v2')",
        ),
        "compression": attr.string(
            default = "zstd",
            doc = "Compression algorithm: zstd (default), lz4, gzip, or none",
            values = ["zstd", "lz4", "gzip", "none"],
        ),
        "lib_path": attr.string(
            doc = "Default library path to prepend to LD_LIBRARY_PATH for all binaries (absolute paths only)",
        ),
        "binary_libs": attr.string_dict(
            doc = "Dictionary mapping binary basenames to their library paths (e.g., {'myapp_v3': '/opt/libs/v3'})",
        ),
        "auto_lib_path": attr.string(
            doc = "Auto-generate library paths using template (e.g., '/opt/libs/{{.ArchVersion}}'). Template variables: {{.Arch}}, {{.Version}}, {{.ArchVersion}}",
        ),
        "enable_auto_lib": attr.bool(
            default = False,
            doc = "Enable automatic library path generation using default two-path format: /opt/<arch>/lib:/usr/lib<width>/glibc-hwcaps/<arch-version>",
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

    # With automatic library paths (default two-path format):
    adipo_fat_binary(
        name = "myapp_fat_autolib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v2": "x86-64-v2",
            ":myapp_v4": "x86-64-v4",
        },
        enable_auto_lib = True,
        # Results in:
        # myapp_v1 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v1
        # myapp_v2 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v2
        # myapp_v4 → /opt/x86-64/lib:/usr/lib64/glibc-hwcaps/x86-64-v4
    )

    # With custom template library paths:
    adipo_fat_binary(
        name = "myapp_fat_customlib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v3": "x86-64-v3",
        },
        auto_lib_path = "/opt/glibc-{{.Version}}/lib",
        # Results in:
        # myapp_v1 → /opt/glibc-v1/lib
        # myapp_v3 → /opt/glibc-v3/lib
    )

    # With per-binary library paths:
    adipo_fat_binary(
        name = "myapp_fat_perlib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v3": "x86-64-v3",
        },
        binary_libs = {
            "myapp_v1": "/custom/path/v1",
            "myapp_v3": "/custom/path/v3",
        },
    )

    # With fixed library path for all binaries:
    adipo_fat_binary(
        name = "myapp_fat_fixedlib",
        binaries = {
            ":myapp_v1": "x86-64-v1",
            ":myapp_v2": "x86-64-v2",
        },
        lib_path = "/opt/myapp/lib",
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
