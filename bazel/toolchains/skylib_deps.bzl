"""bazel skylib"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def skylib_deps():
    http_archive(
        name = "bazel_skylib",
        sha256 = "cd55a062e763b9349921f0f5db8c3933288dc8ba4f76dd9416aac68acee3cb94",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.5.0/bazel-skylib-1.5.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/66ffd9315665bfaafc96b52278f57c7e2dd09f5ede279ea6d39b2be471e7e3aa",
            "https://cdn.confidential.cloud/constellation/cas/sha256/cd55a062e763b9349921f0f5db8c3933288dc8ba4f76dd9416aac68acee3cb94",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.5.0/bazel-skylib-1.5.0.tar.gz",
        ],
        type = "tar.gz",
    )
