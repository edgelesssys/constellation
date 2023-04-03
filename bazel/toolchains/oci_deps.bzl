"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    http_archive(
        name = "rules_oci",
        sha256 = "4a738bdbeacb0e1df070209dddfa7b55fed9bbc553b905cf3d2dd25115e0b598",
        strip_prefix = "rules_oci-0.3.8",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a738bdbeacb0e1df070209dddfa7b55fed9bbc553b905cf3d2dd25115e0b598",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v0.3.8/rules_oci-v0.3.8.tar.gz",
        ],
    )
