"""
Go toolchain for the host platformS
Inspired by https://github.com/bazel-contrib/rules_go/blob/6e4fdcfeb1a333b54ab39ae3413d4ded46d8958d/go/private/rules/go_bin_for_host.bzl
"""

load("@local_config_platform//:constraints.bzl", "HOST_CONSTRAINTS")

GO_TOOLCHAIN = "@io_bazel_rules_go//go:toolchain"

def _ensure_target_cfg(ctx):
    if "-exec" in ctx.bin_dir.path or "/host/" in ctx.bin_dir.path:
        fail("exec not found")

def _go_bin_for_host_impl(ctx):
    _ensure_target_cfg(ctx)
    sdk = ctx.toolchains[GO_TOOLCHAIN].sdk
    sdk_files = ctx.runfiles([sdk.go] + sdk.headers.to_list() + sdk.libs.to_list() + sdk.srcs.to_list() + sdk.tools.to_list())
    return [
        DefaultInfo(
            files = depset([sdk.go]),
            runfiles = sdk_files,
        ),
    ]

go_bin_for_host = rule(
    implementation = _go_bin_for_host_impl,
    toolchains = [GO_TOOLCHAIN],
    exec_compatible_with = HOST_CONSTRAINTS,
)
