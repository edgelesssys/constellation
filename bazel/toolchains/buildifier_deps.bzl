"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        strip_prefix = "buildtools-6.1.2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/977a0bd4593c8d4c8f45e056d181c35e48aa01ad4f8090bdb84f78dca42f47dc",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v6.1.2.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "977a0bd4593c8d4c8f45e056d181c35e48aa01ad4f8090bdb84f78dca42f47dc",
    )
