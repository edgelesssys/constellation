"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-1.4.2",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/be1fce88d05dd0b946f3c874f8af1a473468ea4daba0b69b459a5866416e10d5",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v1.4.2/rules_oci-v1.4.2.tar.gz",
        ],
        sha256 = "be1fce88d05dd0b946f3c874f8af1a473468ea4daba0b69b459a5866416e10d5",
    )
