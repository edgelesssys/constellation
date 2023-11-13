"""hermetic cc toolchain (bazel-zig-cc) build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def hermetic_cc_deps():
    """Loads the dependencies for hermetic_cc_toolchain."""

    http_archive(
        name = "hermetic_cc_toolchain",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a5caccbf6d86d4f60afd45b541a05ca4cc3f5f523aec7d3f7711e584600fb075",
            "https://github.com/uber/hermetic_cc_toolchain/releases/download/v2.1.3/hermetic_cc_toolchain-v2.1.3.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a5caccbf6d86d4f60afd45b541a05ca4cc3f5f523aec7d3f7711e584600fb075",
    )
