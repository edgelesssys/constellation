"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "56d8c5a5c91e1af73eca71a6fab2ced959b67c86d12ba37feedb0a2dfea441a6",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.37.0/rules_go-v0.37.0.zip",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.37.0/rules_go-v0.37.0.zip",
        ],
    )
    http_archive(
        name = "bazel_gazelle",
        strip_prefix = "bazel-gazelle-af021617bef884e1e40afb01b692092a471d9a48",
        sha256 = "cf4f8a87304f417840e5b854eabf4f296e6fd808f7e963ee4d1fbb0c2c190f8a",
        urls = [
            # Depending on main until the next release, needed change from https://github.com/bazelbuild/bazel-gazelle/pull/1432
            # so that "go:embed all:" directives work.
            "https://github.com/bazelbuild/bazel-gazelle/archive/af021617bef884e1e40afb01b692092a471d9a48.tar.gz",
        ],
    )
