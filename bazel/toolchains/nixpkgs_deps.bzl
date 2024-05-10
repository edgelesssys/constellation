"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "2a555348d7f8593fca2bf3fc6ce53c5d62929de81b6c292e23f16c557c0ae45a",
        strip_prefix = "rules_nixpkgs-0.11.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2a555348d7f8593fca2bf3fc6ce53c5d62929de81b6c292e23f16c557c0ae45a",
            "https://github.com/tweag/rules_nixpkgs/releases/download/v0.11.1/rules_nixpkgs-0.11.1.tar.gz",
        ],
        type = "tar.gz",
    )
