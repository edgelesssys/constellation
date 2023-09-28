"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "52325c7e982179c97530228a98ffd964530f746de91cf3dc77a134f9fbe7c53a",
        strip_prefix = "rules_go-c2406b27f77e186d0fb3548d8526e6c14a3236d4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52325c7e982179c97530228a98ffd964530f746de91cf3dc77a134f9fbe7c53a",
            "https://github.com/bazelbuild/rules_go/archive/c2406b27f77e186d0fb3548d8526e6c14a3236d4.zip",
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
