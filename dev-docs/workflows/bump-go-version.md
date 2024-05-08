# Bump Go version

`govulncheck` from the bazel `check` target will fail if our code is vulnerable, which is often the case when a patch version was released with security fixes.

## Steps

Replace "1.xx.x" with the new version in [WORKSPACE.bazel](/WORKSPACE.bazel):

```starlark
load("@io_bazel_rules_go//go:deps.bzl", "go_download_sdk", "go_register_toolchains", "go_rules_dependencies")

go_download_sdk(
    name = "go_sdk",
    patches = ["//3rdparty/bazel/org_golang:go_tls_max_handshake_size.patch"],
    version = "1.xx.x", <--- Replace this one
              ~~~~~~~~
)

```
