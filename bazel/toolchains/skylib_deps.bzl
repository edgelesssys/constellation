"""bazel skylib"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def skylib_deps():
    http_archive(
        name = "bazel_skylib",
        sha256 = "66ffd9315665bfaafc96b52278f57c7e2dd09f5ede279ea6d39b2be471e7e3aa",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-1.4.2.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/66ffd9315665bfaafc96b52278f57c7e2dd09f5ede279ea6d39b2be471e7e3aa",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-1.4.2.tar.gz",
        ],
        type = "tar.gz",
    )
