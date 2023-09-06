"""Repository rule to autoconfigure a toolchain using the system mkosi."""

def _write_build(rctx, path):
    if not path:
        path = ""
    rctx.template(
        "BUILD",
        Label("//bazel/mkosi:BUILD.tpl"),
        substitutions = {
            "{GENERATOR}": "@constellation//bazel/mkosi/mkosi_configure.bzl%find_system_mkosi",
            "{MKOSI_PATH}": str(path),
        },
        executable = False,
    )

def _find_system_mkosi_impl(rctx):
    mkosi_path = rctx.which("mkosi")
    if rctx.attr.verbose:
        if mkosi_path:
            print("Found mkosi at '%s'" % mkosi_path)  # buildifier: disable=print
        else:
            print("No system mkosi found.")  # buildifier: disable=print
    _write_build(rctx = rctx, path = mkosi_path)

_find_system_mkosi = repository_rule(
    implementation = _find_system_mkosi_impl,
    doc = """Create a repository that defines an mkosi toolchain based on the system mkosi.""",
    local = True,
    environ = ["PATH"],
    attrs = {
        "verbose": attr.bool(
            doc = "If true, print status messages.",
        ),
    },
)

def find_system_mkosi(name, verbose = False):
    _find_system_mkosi(name = name, verbose = verbose)
    native.register_toolchains(
        "@constellation//bazel/mkosi:mkosi_nix_toolchain",
        "@%s//:mkosi_auto_toolchain" % name,
        "@constellation//bazel/mkosi:mkosi_missing_toolchain",
    )
