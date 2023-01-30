"""bazel-zig-cc build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

BAZEL_ZIG_CC_VERSION = "v1.0.0"

def zig_cc_deps():
    """Loads the dependencies for bazel-zig-cc."""

    http_archive(
        name = "bazel_skylib",
        sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        ],
    )

    http_archive(
        name = "bazel-zig-cc",
        sha256 = "1f4a1d1e0f6b3e5aa6e1c225fcb23c032f8849441de97b9a38d6ea37362d28e2",
        strip_prefix = "bazel-zig-cc-{}".format(BAZEL_ZIG_CC_VERSION),
        urls = ["https://git.sr.ht/~motiejus/bazel-zig-cc/archive/{}.tar.gz".format(BAZEL_ZIG_CC_VERSION)],
    )
