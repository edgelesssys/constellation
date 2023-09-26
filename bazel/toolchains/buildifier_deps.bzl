"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        strip_prefix = "buildtools-6.3.3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42968f9134ba2c75c03bb271bd7bb062afb7da449f9b913c96e5be4ce890030a",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v6.3.3.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "42968f9134ba2c75c03bb271bd7bb062afb7da449f9b913c96e5be4ce890030a",
    )
