"""Bazel rules for devbuild"""

load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")

def _cli_edition_impl(ctx):
    cli_edition = ctx.attr._cli_edition[BuildSettingInfo].value
    if cli_edition == None or cli_edition == "":
        fail("--cli_edition is not set in .bazeloverwriterc")
    if cli_edition not in ["oss", "enterprise"]:
        fail("--cli_edition must be 'oss' or 'enterprise' in .bazeloverwriterc")

    output = ctx.actions.declare_file(ctx.label.name + ".txt")
    ctx.actions.write(output = output, content = cli_edition)
    return [DefaultInfo(files = depset([output]))]

cli_edition = rule(
    implementation = _cli_edition_impl,
    attrs = {
        "_cli_edition": attr.label(default = Label("//bazel/settings:cli_edition")),
    },
)
