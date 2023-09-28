""" A repository rule use either nixpkgs or download a go toolchain / SDK """

def _has_nix(ctx):
    return ctx.os.environ.get("BAZEL_NIX_HOST_PLATFORM", "0") == "1"

def _gen_imports_impl(ctx):
    ctx.file("BUILD", "")

    imports_for_nix = """
load("@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl", "nixpkgs_cc_configure")
load("@io_tweag_rules_nixpkgs//nixpkgs:toolchains/go.bzl", "nixpkgs_go_configure")

def go_toolchain():
    nixpkgs_go_configure(
        repository = "@nixpkgs",
        attribute_path = "go_1_21",
    )

def cc_toolchain():
    nixpkgs_cc_configure(repository = "@nixpkgs")
    native.register_toolchains(
        "@zig_sdk//libc_aware/toolchain:linux_amd64_gnu.2.23",
        "@zig_sdk//libc_aware/toolchain:linux_arm64_gnu.2.23",
        "@zig_sdk//libc_aware/toolchain:linux_amd64_musl",
        "@zig_sdk//libc_aware/toolchain:linux_arm64_musl",
        "@zig_sdk//toolchain:linux_amd64_gnu.2.23",
        "@zig_sdk//toolchain:linux_arm64_gnu.2.23",
        "@zig_sdk//toolchain:linux_amd64_musl",
        "@zig_sdk//toolchain:linux_arm64_musl",
        "@zig_sdk//toolchain:darwin_amd64",
        "@zig_sdk//toolchain:darwin_arm64",
        "@zig_sdk//toolchain:windows_amd64",
    )
    """
    imports_for_non_nix = """
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains")

def go_toolchain():
    go_register_toolchains(version = "1.21.1")

def cc_toolchain():
    native.register_toolchains(
        "@zig_sdk//libc_aware/toolchain:linux_amd64_gnu.2.23",
        "@zig_sdk//libc_aware/toolchain:linux_arm64_gnu.2.23",
        "@zig_sdk//libc_aware/toolchain:linux_amd64_musl",
        "@zig_sdk//libc_aware/toolchain:linux_arm64_musl",
        "@zig_sdk//toolchain:linux_amd64_gnu.2.23",
        "@zig_sdk//toolchain:linux_arm64_gnu.2.23",
        "@zig_sdk//toolchain:linux_amd64_musl",
        "@zig_sdk//toolchain:linux_arm64_musl",
        "@zig_sdk//toolchain:darwin_amd64",
        "@zig_sdk//toolchain:darwin_arm64",
        "@zig_sdk//toolchain:windows_amd64",
    )
    """

    if _has_nix(ctx):
        ctx.file("imports.bzl", imports_for_nix)
    else:
        ctx.file("imports.bzl", imports_for_non_nix)

_gen_imports = repository_rule(
    implementation = _gen_imports_impl,
)

def gen_imports():
    _gen_imports(
        name = "nixos_support",
    )
