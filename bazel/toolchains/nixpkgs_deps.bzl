"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "30271f7bd380e4e20e4d7132c324946c4fdbc31ebe0bbb6638a0f61a37e74397",
        strip_prefix = "rules_nixpkgs-0.13.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/30271f7bd380e4e20e4d7132c324946c4fdbc31ebe0bbb6638a0f61a37e74397",
            "https://github.com/tweag/rules_nixpkgs/releases/download/v0.13.0/rules_nixpkgs-0.13.0.tar.gz",
        ],
        type = "tar.gz",
    )
