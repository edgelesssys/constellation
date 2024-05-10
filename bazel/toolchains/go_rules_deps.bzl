"""Go toolchain dependencies for Bazel.

Defines hermetic go toolchains and rules to build and test go code.
Gazelle is a build file generator for Bazel projects written in Go.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def go_deps():
    http_archive(
        name = "io_bazel_rules_go",
        sha256 = "f74c98d6df55217a36859c74b460e774abc0410a47cc100d822be34d5f990f16",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.47.1/rules_go-v0.47.1.zip",
            "https://cdn.confidential.cloud/constellation/cas/sha256/f74c98d6df55217a36859c74b460e774abc0410a47cc100d822be34d5f990f16",
            "https://github.com/bazelbuild/rules_go/releases/download/v0.47.1/rules_go-v0.47.1.zip",
        ],
        remote_patches = {
            # Move timeout handling back to bzltestutil
            # remove after https://github.com/bazelbuild/rules_go/pull/3939 is merged
            "https://github.com/bazelbuild/rules_go/commit/cc911bfec4f52d93d1c47cc92a3bc03ec8f9cb33.patch": "sha256-Z1jNoEagzSghHrf1SiLLMLGpFq/IBvOjZMxWaIk1O3M=",
        },
        remote_patch_strip = 1,
        type = "zip",
    )
    http_archive(
        name = "bazel_gazelle",
        sha256 = "75df288c4b31c81eb50f51e2e14f4763cb7548daae126817247064637fd9ea62",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.36.0/bazel-gazelle-v0.36.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/75df288c4b31c81eb50f51e2e14f4763cb7548daae126817247064637fd9ea62",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.36.0/bazel-gazelle-v0.36.0.tar.gz",
        ],
        type = "tar.gz",
    )
