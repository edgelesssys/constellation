"""bazel-zig-cc build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def zig_cc_deps():
    """Loads the dependencies for bazel-zig-cc."""

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

    # TODO(malt3): Update to a release version once the next release is out.
    # Upgraded to work around https://github.com/uber/bazel-zig-cc/issues/22
    # See also https://github.com/uber/bazel-zig-cc/pull/23
    http_archive(
        name = "bazel-zig-cc",
        sha256 = "bea372f7f9bd8541f7b0a152c76c7b9396201c36a0ed229b36c48301815c3141",
        strip_prefix = "bazel-zig-cc-f3e4542bd62f4aef794a3d184140a9d30b8fadb8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bea372f7f9bd8541f7b0a152c76c7b9396201c36a0ed229b36c48301815c3141",
            "https://github.com/uber/bazel-zig-cc/archive/f3e4542bd62f4aef794a3d184140a9d30b8fadb8.tar.gz",
        ],
        type = "tar.gz",
    )
