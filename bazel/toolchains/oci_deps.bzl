"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    http_archive(
        name = "rules_oci",
        sha256 = "4f119dc9e08319a3262c04b334bda54ba0484ca34f8ead706dd2397fc11816f7",
        strip_prefix = "rules_oci-0.3.3",
        url = "https://github.com/bazel-contrib/rules_oci/releases/download/v0.3.3/rules_oci-v0.3.3.tar.gz",
    )
