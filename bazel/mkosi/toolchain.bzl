"""toolchain to provide an mkosi binary."""

TOOLCHAIN_TYPE = "@constellation//bazel/mkosi:toolchain_type"
MAKE_VARIABLES = "@constellation//bazel/mkosi:make_variables"

MkosiInfo = provider(
    doc = """Information needed to invoke mkosi.""",
    fields = {
        "label": "Label of a target providing a mkosi binary",
        "name": "The name of the toolchain",
        "path": "Path to the mkosi binary",
        "valid": "Is this toolchain valid and usable?",
    },
)

def _mkosi_toolchain_impl(ctx):
    if ctx.attr.label and ctx.attr.path:
        fail("mkosi_toolchain: label and path are mutually exclusive")
    valid = bool(ctx.attr.label) or bool(ctx.attr.path)

    mkosi_info = MkosiInfo(
        name = str(ctx.label),
        valid = valid,
        label = ctx.attr.label,
        path = ctx.attr.path,
    )
    toolchain_info = platform_common.ToolchainInfo(
        mkosi = mkosi_info,
    )
    return [toolchain_info]

mkosi_toolchain = rule(
    implementation = _mkosi_toolchain_impl,
    attrs = {
        "label": attr.label(
            doc = "Label of a target providing a mkosi binary. Mutually exclusive with path.",
            executable = True,
            cfg = "exec",
            allow_single_file = True,
            default = None,
        ),
        "path": attr.string(
            doc = "Path to the mkosi binary. Mutually exclusive with label.",
        ),
    },
)

def _mkosi_make_variables_impl(ctx):
    info = ctx.toolchains[TOOLCHAIN_TYPE].mkosi
    variables = {}
    if info.valid:
        binary_path = info.label[DefaultInfo].files_to_run.executable.path if info.label else info.path
        variables["MKOSI"] = binary_path
    return [platform_common.TemplateVariableInfo(variables)]

mkosi_make_variables = rule(
    doc = "Make variables for mkosi.",
    implementation = _mkosi_make_variables_impl,
    toolchains = [TOOLCHAIN_TYPE],
)

def _is_mkosi_available_impl(ctx):
    toolchain = ctx.toolchains[TOOLCHAIN_TYPE]
    available = toolchain and toolchain.mkosi.valid
    return [config_common.FeatureFlagInfo(
        value = ("1" if available else "0"),
    )]

is_mkosi_available = rule(
    implementation = _is_mkosi_available_impl,
    attrs = {},
    toolchains = [TOOLCHAIN_TYPE],
)
