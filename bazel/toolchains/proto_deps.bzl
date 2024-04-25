"""proto toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_deps():
    http_archive(
        name = "rules_proto",
        sha256 = "17fa03f509b0d1df05c70c174a266ab211d04b9969e41924fd07a81ea171f117",
        strip_prefix = "rules_proto-d205d37866925569d99b4d6cdcba172326ecf812",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17fa03f509b0d1df05c70c174a266ab211d04b9969e41924fd07a81ea171f117",
            "https://github.com/bazelbuild/rules_proto/archive/d205d37866925569d99b4d6cdcba172326ecf812.tar.gz",
        ],
        type = "tar.gz",
    )
