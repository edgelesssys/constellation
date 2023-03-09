"""proto toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_deps():
    http_archive(
        name = "rules_proto",
        sha256 = "17fa03f509b0d1df05c70c174a266ab211d04b9969e41924fd07a81ea171f117",
        strip_prefix = "rules_proto-cda0effe6b5af095a6886c67f90c760b83f08c48",
        urls = [
            "https://github.com/bazelbuild/rules_proto/archive/cda0effe6b5af095a6886c67f90c760b83f08c48.tar.gz",
        ],
    )
