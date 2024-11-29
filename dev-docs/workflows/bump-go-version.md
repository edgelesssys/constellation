# Bump Go version

`govulncheck` from the bazel `check` target will fail if our code is vulnerable, which is often the case when a patch version was released with security fixes.

## Steps

Replace `"1.xx.x"` with the new version in [MODULE.bazel](/MODULE.bazel):

```starlark
go_sdk = use_extension("@io_bazel_rules_go//go:extensions.bzl", "go_sdk")
go_sdk.download(
    name = "go_sdk",
    patches = ["//3rdparty/bazel/org_golang:go_tls_max_handshake_size.patch"],
    version = "1.xx.x", <--- Replace this one
              ~~~~~~~~
)

```

Replace `go-version: "1.xx.x"` with the new version in all GitHub actions and workflows.
You can use the following command to find replace all instances of `go-version: "1.xx.x"` in the `.github` directory:

```bash
OLD_VERSION="1.xx.x"
NEW_VERSION="1.xx.y"
find .github -type f -exec sed -i "s/go-version: \"${OLD_VERSION}\"/go-version: \"${NEW_VERSION}\"/g" {} \;
sed -i "s/go ${OLD_VERSION}/go ${NEW_VERSION}/g" go.mod
sed -i "s/${OLD_VERSION}/${NEW_VERSION}/g" go.work
```

Or manually:

```yaml
- name: Setup Go environment
  uses: actions/setup-go@v5
  with:
  go-version: "1.xx.x" <--- Replace this one
              ~~~~~~~~
```
