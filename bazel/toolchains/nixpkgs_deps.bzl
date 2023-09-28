"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "6a7c2a85b9d21c0ba617229597f1bda3c871b0ec81090d89f75f3acdef32cd16",
        strip_prefix = "rules_nixpkgs-bb2f6996e1e6933cecaf06e60799ca7c90ef3928",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6a7c2a85b9d21c0ba617229597f1bda3c871b0ec81090d89f75f3acdef32cd16",
            "https://github.com/tweag/rules_nixpkgs/archive/bb2f6996e1e6933cecaf06e60799ca7c90ef3928.tar.gz",
        ],
        type = "tar.gz",
    )
