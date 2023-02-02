"""proto toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_deps():
    http_archive(
        name = "rules_proto",
        sha256 = "c9cc7f7be05e50ecd64f2b0dc2b9fd6eeb182c9cc55daf87014d605c31548818",
        strip_prefix = "rules_proto-3f1ab99b718e3e7dd86ebdc49c580aa6a126b1cd",
        urls = [
            "https://github.com/bazelbuild/rules_proto/archive/3f1ab99b718e3e7dd86ebdc49c580aa6a126b1cd.tar.gz",
        ],
    )
