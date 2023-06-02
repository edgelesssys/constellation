"""hermetic cc toolchain (bazel-zig-cc) build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def hermetic_cc_deps():
    """Loads the dependencies for hermetic_cc_toolchain."""

    http_archive(
        name = "bazel_skylib",
        sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
            "https://cdn.confidential.cloud/constellation/cas/sha256/74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        ],
        type = "tar.gz",
    )

    http_archive(
        name = "hermetic_cc_toolchain",
        sha256 = "43a1b398f08109c4f03b9ba2b3914bd43d1fec0425f71b71f802bf3f78cee0c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/43a1b398f08109c4f03b9ba2b3914bd43d1fec0425f71b71f802bf3f78cee0c2",
            "https://github.com/uber/hermetic_cc_toolchain/releases/download/v2.0.0/hermetic_cc_toolchain-v2.0.0.tar.gz",
        ],
        type = "tar.gz",
    )
