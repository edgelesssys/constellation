"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "278b7ff5a826f3dc10f04feaf0b70d48b68748ccd512d7f98bf442077f043fe3",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.41.0/rules_go-v0.41.0.zip",
            "https://cdn.confidential.cloud/constellation/cas/sha256/51dc53293afe317d2696d4d6433a4c33feedb7748a9e352072e2ec3c0dafd2c6",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.41.0/rules_go-v0.41.0.zip",
        ],
        type = "zip",
    )
    http_archive(
        name = "bazel_gazelle",
        sha256 = "29218f8e0cebe583643cbf93cae6f971be8a2484cdcfa1e45057658df8d54002",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.32.0/bazel-gazelle-v0.32.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8b6d75de6e4bf7c41b7737b183523085f56283f6db929b86c5e7e1f09cf59c9",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.32.0/bazel-gazelle-v0.32.0.tar.gz",
        ],
        type = "tar.gz",
    )
