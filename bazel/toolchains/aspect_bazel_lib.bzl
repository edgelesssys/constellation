"""aspect bazel library"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aspect_bazel_lib():
    http_archive(
        name = "aspect_bazel_lib",
        sha256 = "979667bb7276ee8fcf2c114c9be9932b9a3052a64a647e0dcaacfb9c0016f0a3",
        strip_prefix = "bazel-lib-2.4.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/979667bb7276ee8fcf2c114c9be9932b9a3052a64a647e0dcaacfb9c0016f0a3",
            "https://github.com/aspect-build/bazel-lib/releases/download/v2.4.1/bazel-lib-v2.4.1.tar.gz",
        ],
        type = "tar.gz",
    )
