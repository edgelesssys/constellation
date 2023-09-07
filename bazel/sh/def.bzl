"""Bazel rules for CI and dev tooling"""

load("@bazel_skylib//lib:shell.bzl", "shell")

def _sh_template_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")

    substitutions = {
        "@@BASE_LIB@@": ctx.file._base_lib.path,
    }
    for k, v in ctx.attr.substitutions.items():
        sub = ctx.expand_location(v, ctx.attr.data)
        sub = ctx.expand_make_variables("substitutions", sub, {})
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
        runfiles = ctx.runfiles(files = ctx.files.data + [ctx.file._base_lib]),
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
        "_base_lib": attr.label(
            default = Label("@constellation//bazel/sh:base_lib"),
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
    data = kwargs.pop("data", [])
    substitutions = kwargs.pop("substitutions", [])
    template = kwargs.pop("template", [])
    toolchains = kwargs.pop("toolchains", [])

    _sh_template(
        name = script_name,
        tags = tags,
        data = data,
        substitutions = substitutions,
        template = template,
        toolchains = toolchains,
    )

    native.sh_binary(
        name = name,
        srcs = [script_name],
        data = [script_name] + data,
        **kwargs
    )

def sh_test_template(name, **kwargs):
    """Build a sh_test from a template

    Args:
        name: name
        **kwargs: **kwargs
    """
    script_name = name + "-script"

    tags = kwargs.get("tags", [])
    data = kwargs.pop("data", [])
    substitutions = kwargs.pop("substitutions", [])
    template = kwargs.pop("template", [])

    _sh_template(
        name = script_name,
        tags = tags,
        data = data,
        substitutions = substitutions,
        template = template,
    )

    native.sh_test(
        name = name,
        srcs = [script_name],
        data = [script_name] + data,
        **kwargs
    )

def repo_command(name, **kwargs):
    """Build a sh_binary that executes a single command.

    Args:
        name: name
        **kwargs: **kwargs
    """
    cmd = kwargs.pop("command")
    args = shell.array_literal(kwargs.pop("args", []))

    substitutions = {
        "@@ARGS@@": args,
        "@@CMD@@": "$(rootpath %s)" % cmd,
    }

    data = kwargs.pop("data", [])
    data.append(cmd)

    sh_template(
        name = name,
        data = data,
        substitutions = substitutions,
        template = "//bazel/sh:repo_command.sh.in",
        **kwargs
    )

def noop_warn(name, **kwargs):
    """Build a sh_binary that warns about a step beeing replaced by a no-op.

    Args:
        name: name
        **kwargs: **kwargs
    """
    warning = kwargs.pop("warning", "The binary that should have been executed is likely not available on your platform.")
    warning = "\\033[0;33mWARNING:\\033[0m This step is a no-op. %s" % warning
    substitutions = {
        "@@WARNING@@": warning,
    }

    sh_template(
        name = name,
        substitutions = substitutions,
        template = "//bazel/sh:noop_warn.sh.in",
        **kwargs
    )
