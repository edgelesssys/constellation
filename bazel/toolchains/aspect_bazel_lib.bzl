"""aspect bazel library"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aspect_bazel_lib():
    http_archive(
        name = "aspect_bazel_lib",
        sha256 = "c858cc637db5370f6fd752478d1153955b4b4cbec7ffe95eb4a47a48499a79c3",
        strip_prefix = "bazel-lib-2.0.3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c858cc637db5370f6fd752478d1153955b4b4cbec7ffe95eb4a47a48499a79c3",
            "https://github.com/aspect-build/bazel-lib/releases/download/v2.0.3/bazel-lib-v2.0.3.tar.gz",
        ],
        type = "tar.gz",
    )
