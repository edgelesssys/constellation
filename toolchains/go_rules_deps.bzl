"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        strip_prefix = "rules_go-afbd9c4999970bdbe8dca9dda277c2a7b8ce7eea",
        sha256 = "02476ca4e422d7e7c35ec13c667e43c2e32d821c61eb7976b2116b73079ec971",
        urls = [
            "https://github.com/bazelbuild/rules_go/archive/afbd9c4999970bdbe8dca9dda277c2a7b8ce7eea.tar.gz",
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
