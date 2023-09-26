"""hermetic cc toolchain (bazel-zig-cc) build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def hermetic_cc_deps():
    """Loads the dependencies for hermetic_cc_toolchain."""

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

    http_archive(
        name = "hermetic_cc_toolchain",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/28fc71b9b3191c312ee83faa1dc65b38eb70c3a57740368f7e7c7a49bedf3106",
            "https://github.com/uber/hermetic_cc_toolchain/releases/download/v2.1.2/hermetic_cc_toolchain-v2.1.2.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "28fc71b9b3191c312ee83faa1dc65b38eb70c3a57740368f7e7c7a49bedf3106",
    )
