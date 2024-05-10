"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        strip_prefix = "buildtools-7.1.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/60a9025072ae237f325d0e7b661e1685f34922c29883888c2d06f5789462b939",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v7.1.1.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "60a9025072ae237f325d0e7b661e1685f34922c29883888c2d06f5789462b939",
    )
