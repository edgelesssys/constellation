"""aspect bazel library"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aspect_bazel_lib():
    http_archive(
        name = "aspect_bazel_lib",
        sha256 = "87ab4ec479ebeb00d286266aca2068caeef1bb0b1765e8f71c7b6cfee6af4226",
        strip_prefix = "bazel-lib-2.7.3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/87ab4ec479ebeb00d286266aca2068caeef1bb0b1765e8f71c7b6cfee6af4226",
            "https://github.com/aspect-build/bazel-lib/releases/download/v2.7.3/bazel-lib-v2.7.3.tar.gz",
        ],
        type = "tar.gz",
    )
