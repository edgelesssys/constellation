"""proto toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_deps():
    http_archive(
        name = "rules_proto",
        sha256 = "17fa03f509b0d1df05c70c174a266ab211d04b9969e41924fd07a81ea171f117",
        strip_prefix = "rules_proto-f889a1b532fdca5f5051691f023a6a9f37ce494f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17fa03f509b0d1df05c70c174a266ab211d04b9969e41924fd07a81ea171f117",
            "https://github.com/bazelbuild/rules_proto/archive/f889a1b532fdca5f5051691f023a6a9f37ce494f.tar.gz",
        ],
        type = "tar.gz",
    )
