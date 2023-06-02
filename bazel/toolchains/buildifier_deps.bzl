"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        sha256 = "ae34c344514e08c23e90da0e2d6cb700fcd28e80c02e23e4d5715dddcb42f7b3",
        strip_prefix = "buildtools-4.2.5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ae34c344514e08c23e90da0e2d6cb700fcd28e80c02e23e4d5715dddcb42f7b3",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/4.2.5.tar.gz",
        ],
        type = "tar.gz",
    )
