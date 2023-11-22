"""A rule to build a single executable for a specific platform."""

def _platform_transition_impl(settings, attr):
    _ignore = settings  # @unused
    return {
        "//command_line_option:platforms": "{}".format(attr.platform),
    }

_platform_transition = transition(
    implementation = _platform_transition_impl,
    inputs = [],
    outputs = [
        "//command_line_option:platforms",
    ],
)

def _platform_binary_impl(ctx):
    out = ctx.actions.declare_file("{}_{}".format(ctx.file.target_file.basename, ctx.attr.platform))
    ctx.actions.symlink(output = out, target_file = ctx.file.target_file)
    runfiles = ctx.runfiles(files = ctx.files.target_file)
    runfiles = runfiles.merge(ctx.attr.target_file[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.attr.target_file[DefaultInfo].data_runfiles)

    return [
        DefaultInfo(
            executable = out,
            files = depset([out]),
            runfiles = runfiles,
            # runfiles = ctx.attr.target_file[DefaultInfo].default_runfiles,
        ),
    ]

_attrs = {
    "platform": attr.string(
        doc = "The platform to build the target for.",
    ),
    "target_file": attr.label(
        allow_single_file = True,
        mandatory = True,
        doc = "Target to build.",
    ),
    "_allowlist_function_transition": attr.label(
        default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
    ),
}

# wrap a single exectable and build it for the specified platform.
platform_binary = rule(
    implementation = _platform_binary_impl,
    cfg = _platform_transition,
    attrs = _attrs,
    executable = True,
)
