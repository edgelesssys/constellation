"""hermetic cc toolchain (bazel-zig-cc) build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def hermetic_cc_deps():
    """Loads the dependencies for hermetic_cc_toolchain."""

    http_archive(
        name = "hermetic_cc_toolchain",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b8107de0d017fe32e6434086a9568f97c60a111b49dc34fc7001e139c30fdea",
            "https://github.com/uber/hermetic_cc_toolchain/releases/download/v2.2.1/hermetic_cc_toolchain-v2.2.1.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "3b8107de0d017fe32e6434086a9568f97c60a111b49dc34fc7001e139c30fdea",
    )
