"""hermetic cc toolchain (bazel-zig-cc) build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def hermetic_cc_deps():
    """Loads the dependencies for hermetic_cc_toolchain."""

    http_archive(
        name = "hermetic_cc_toolchain",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3bc6ec127622fdceb4129cb06b6f7ab098c4d539124dde96a6318e7c32a53f7a",
            "https://github.com/uber/hermetic_cc_toolchain/releases/download/v3.0.1/hermetic_cc_toolchain-v3.0.1.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "3bc6ec127622fdceb4129cb06b6f7ab098c4d539124dde96a6318e7c32a53f7a",
    )
