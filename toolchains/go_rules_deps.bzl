"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "dd926a88a564a9246713a9c00b35315f54cbd46b31a26d5d8fb264c07045f05d",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.38.1/rules_go-v0.38.1.zip",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.38.1/rules_go-v0.38.1.zip",
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
