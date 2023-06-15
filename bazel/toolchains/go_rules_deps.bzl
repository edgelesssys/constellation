"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "bfb0ca2f7b29b87ec81a99aec106f56c1f33bc0e29988ee61d4dbfc3cfa2829b",
        strip_prefix = "rules_go-286e96454a739681639ef3289d92c23060fcb5af",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bfb0ca2f7b29b87ec81a99aec106f56c1f33bc0e29988ee61d4dbfc3cfa2829b",
            "https://github.com/bazelbuild/rules_go/archive/286e96454a739681639ef3289d92c23060fcb5af.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "bazel_gazelle",
        sha256 = "b8b6d75de6e4bf7c41b7737b183523085f56283f6db929b86c5e7e1f09cf59c9",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.31.1/bazel-gazelle-v0.31.1.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8b6d75de6e4bf7c41b7737b183523085f56283f6db929b86c5e7e1f09cf59c9",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.31.1/bazel-gazelle-v0.31.1.tar.gz",
        ],
        type = "tar.gz",
    )
