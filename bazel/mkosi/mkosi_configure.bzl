"""Repository rule to configure a toolchain using nixpkgs mkosi."""

def register_mkosi(name):
    native.register_toolchains(
        "@constellation//bazel/mkosi:mkosi_nix_toolchain",
        "@constellation//bazel/mkosi:mkosi_missing_toolchain",
    )
