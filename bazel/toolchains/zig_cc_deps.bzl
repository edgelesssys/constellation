"""bazel-zig-cc build rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

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

    # TODO(malt3): Update to a release version once the next release is out.
    # Upgraded to work around https://github.com/uber/bazel-zig-cc/issues/22
    # See also https://github.com/uber/bazel-zig-cc/pull/23
    http_archive(
        name = "bazel-zig-cc",
        sha256 = "9e7592d5714829b769473ec7a7c04de456cd266a117481d2e8f350432d03b7f8",
        strip_prefix = "bazel-zig-cc-6d2ee8cad0b1170883dd13ffddc7b355ee232f07",
        urls = ["https://github.com/uber/bazel-zig-cc/archive/6d2ee8cad0b1170883dd13ffddc7b355ee232f07.tar.gz"],
    )
