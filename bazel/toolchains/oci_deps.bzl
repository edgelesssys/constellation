"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-0.4.0",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d7b0760ba28554b71941ea0bbfd0a9f089bf250fd4448f9c116e1cb7a63b3933",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v0.4.0/rules_oci-v0.4.0.tar.gz",
        ],
        sha256 = "d7b0760ba28554b71941ea0bbfd0a9f089bf250fd4448f9c116e1cb7a63b3933",
    )
