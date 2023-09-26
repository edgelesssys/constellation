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
            "https://cdn.confidential.cloud/constellation/cas/sha256/278b7ff5a826f3dc10f04feaf0b70d48b68748ccd512d7f98bf442077f043fe3",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.41.0/rules_go-v0.41.0.zip",
        ],
        type = "zip",
    )
    http_archive(
        name = "bazel_gazelle",
        sha256 = "d3fa66a39028e97d76f9e2db8f1b0c11c099e8e01bf363a923074784e451f809",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.33.0/bazel-gazelle-v0.33.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/d3fa66a39028e97d76f9e2db8f1b0c11c099e8e01bf363a923074784e451f809",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.33.0/bazel-gazelle-v0.33.0.tar.gz",
        ],
        type = "tar.gz",
    )
