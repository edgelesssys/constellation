"""
This module contains rules and macros for building and testing Go code.
"""

load("@io_bazel_rules_go//go:def.bzl", _go_test = "go_test")

def go_test(ld = None, count = 3, **kwargs):
    """go_test is a wrapper for go_test that uses default settings for Constellation.

    It adds the following:
    - Sets test count to 3.
    - Sets race detector to on by default (except Mac OS)
    - Optionally sets the interpreter path to ld.

    Args:
          ld: path to interpreter to that will be written into the elf header.
          count: number of times each test should be executed. defaults to 3.
          **kwargs: all other arguments are passed to go_test.
    """

    # Sets test count to 3.
    kwargs.setdefault("args", [])
    kwargs["args"].append("--test.count={}".format(count))

    # enable race detector by default
    race_value = select({
        "@platforms//os:macos": "off",
        "//conditions:default": "on",
    })
    pure_value = select({
        "@platforms//os:macos": "on",
        "//conditions:default": "off",
    })
    kwargs.setdefault("race", race_value)
    kwargs.setdefault("pure", pure_value)

    # set gc_linkopts to set the interpreter path to ld.
    kwargs.setdefault("gc_linkopts", [])
    if ld:
        kwargs["gc_linkopts"] += ["-I", ld]

    _go_test(**kwargs)

def _vars_script(env, ld, cmd):
    ret = ["#!/bin/sh"]
    for k, v in env.items():
        ret.append('export {}="{}"'.format(k, v))
    ret.append('exec {} {} "$@"'.format(ld, cmd))
    return "\n".join(ret) + "\n"

def _ld_binary_impl(ctx):
    source_info = ctx.attr.src[DefaultInfo]
    ld_info = ctx.attr.ld[DefaultInfo]

    executable = None
    if source_info.files_to_run and source_info.files_to_run.executable:
        command = _vars_script(ctx.attr.env, ld_info.files_to_run.executable.short_path, source_info.files_to_run.executable.short_path)
        executable = ctx.actions.declare_file("{}_under_ld".format(ctx.file.src.basename))
        ctx.actions.write(
            output = executable,
            content = command,
            is_executable = True,
        )

    runfiles = ctx.runfiles(files = ctx.files.src)
    runfiles = runfiles.merge(source_info.data_runfiles)
    runfiles = runfiles.merge(ctx.runfiles(files = ctx.files.ld))
    runfiles = runfiles.merge(ld_info.data_runfiles)
    return [DefaultInfo(
        executable = executable,
        files = depset([executable]),
        runfiles = runfiles,
    )]

_attrs = {
    "env": attr.string_dict(
        doc = "Environment variables for the test",
    ),
    "ld": attr.label(
        executable = True,
        cfg = "exec",
        doc = "ld wrapper executable",
    ),
    "src": attr.label(
        allow_single_file = True,
        mandatory = True,
        doc = "Target to build.",
    ),
}

ld_test = rule(
    implementation = _ld_binary_impl,
    attrs = _attrs,
    executable = True,
    test = True,
)

def go_ld_test(**kwargs):
    """go_ld_test is a wrapper for go_test that uses the specified ld to run the test binary under.

    Args:
          **kwargs: all arguments are passed to go_test.
    """

    # Sets test count to 3.
    kwargs.setdefault("args", [])
    kwargs["args"].append("--test.count=3")

    # Disable test wrapper
    kwargs.setdefault("env", {})
    kwargs["env"]["GO_TEST_WRAP"] = "0"

    ld_test(
        **kwargs
    )
