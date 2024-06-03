# Bump Go version

`govulncheck` from the bazel `check` target will fail if our code is vulnerable, which is often the case when a patch version was released with security fixes.

## Steps

Replace "1.xx.x" with the new version in [MODULE.bazel](/MODULE.bazel):

```starlark
go_sdk = use_extension("@io_bazel_rules_go//go:extensions.bzl", "go_sdk")
go_sdk.download(
    name = "go_sdk",
    patches = ["//3rdparty/bazel/org_golang:go_tls_max_handshake_size.patch"],
    version = "1.xx.x", <--- Replace this one
              ~~~~~~~~
)

```
