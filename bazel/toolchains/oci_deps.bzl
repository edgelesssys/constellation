"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    http_archive(
        name = "contrib_rules_oci",
        sha256 = "4ba3022f6475df4dbd1b21142f7bcb9bfaf2689a74687536837e1827b6aaddbf",
        strip_prefix = "rules_oci-0.2.0",
        url = "https://github.com/bazel-contrib/rules_oci/archive/v0.2.0.tar.gz",
    )
