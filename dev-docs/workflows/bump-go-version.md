# Bump Go version
`govulncheck` from the bazel `check` target will fail if our code is vulnerable, which is often the case when a patch version was released with security fixes.

## Steps

Replace "1.xx.x" with the new version in [WORKSPACE.bazel](/WORKSPACE.bazel):

```starlark
go_register_toolchains(version = "1.xx.x")
```
