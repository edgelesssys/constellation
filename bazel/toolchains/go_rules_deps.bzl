"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c8035e8ae248b56040a65ad3f0b7434712e2037e5dfdcebfe97576e620422709",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.44.0/rules_go-v0.44.0.zip",
        ],
        type = "zip",
        sha256 = "c8035e8ae248b56040a65ad3f0b7434712e2037e5dfdcebfe97576e620422709",
    )
    http_archive(
        name = "bazel_gazelle",
        sha256 = "b7387f72efb59f876e4daae42f1d3912d0d45563eac7cb23d1de0b094ab588cf",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.34.0/bazel-gazelle-v0.34.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7387f72efb59f876e4daae42f1d3912d0d45563eac7cb23d1de0b094ab588cf",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.34.0/bazel-gazelle-v0.34.0.tar.gz",
        ],
        type = "tar.gz",
    )
