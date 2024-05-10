"""bazel skylib"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def skylib_deps():
    http_archive(
        name = "bazel_skylib",
        sha256 = "9f38886a40548c6e96c106b752f242130ee11aaa068a56ba7e56f4511f33e4f2",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.6.1/bazel-skylib-1.6.1.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f38886a40548c6e96c106b752f242130ee11aaa068a56ba7e56f4511f33e4f2",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.6.1/bazel-skylib-1.6.1.tar.gz",
        ],
        type = "tar.gz",
    )
