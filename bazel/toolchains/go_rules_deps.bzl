"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        strip_prefix = "rules_go-ea3cc4f0778ba4bb35a682affc8e278551187fad",
        sha256 = "9f0c386d233e7160cb752527c34654620cef1920a53617a2f1cca8d8edee5e8a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f0c386d233e7160cb752527c34654620cef1920a53617a2f1cca8d8edee5e8a",
            "https://github.com/bazelbuild/rules_go/archive/ea3cc4f0778ba4bb35a682affc8e278551187fad.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "bazel_gazelle",
        strip_prefix = "bazel-gazelle-97a754c6e45848828b27152fa64ca5dd3003d832",
        sha256 = "2591fe5c9ff639317c5144665f2b97f3e45dac7ebb0b9357f8ddb3533b60a16f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2591fe5c9ff639317c5144665f2b97f3e45dac7ebb0b9357f8ddb3533b60a16f",
            "https://github.com/bazelbuild/bazel-gazelle/archive/97a754c6e45848828b27152fa64ca5dd3003d832.tar.gz",
        ],
        type = "tar.gz",
    )
