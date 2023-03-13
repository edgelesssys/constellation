"""Bazel rules for CI and dev tooling"""

load("@bazel_skylib//lib:shell.bzl", "shell")

def _sh_template_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")

    substitutions = {}
    for k, v in ctx.attr.substitutions.items():
        sub = ctx.expand_location(v, ctx.attr.data)
        sub = shell.quote(sub)
        substitutions[k] = sub

    ctx.actions.expand_template(
        template = ctx.file.template,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )

    return [DefaultInfo(
        files = depset([out_file]),
        executable = out_file,
    )]

_sh_template = rule(
    implementation = _sh_template_impl,
    attrs = {
        "data": attr.label_list(
            allow_files = True,
        ),
        "substitutions": attr.string_dict(),
        "template": attr.label(
            allow_single_file = True,
        ),
    },
)

def sh_template(name, **kwargs):
    """Build a sh_binary from a template

    Args:
        name: name
        **kwargs: **kwargs
    """
    script_name = name + "-script"

    tags = kwargs.get("tags", [])
    data = kwargs.get("data", [])
    substitutions = kwargs.pop("substitutions", [])
    template = kwargs.pop("template", [])

    _sh_template(
        name = script_name,
        tags = tags,
        data = data,
        substitutions = substitutions,
        template = template,
    )

    native.sh_binary(
        name = name,
        srcs = [script_name],
        **kwargs
    )
