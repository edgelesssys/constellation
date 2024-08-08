"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "1adb04dc0416915fef427757f4272c4f7dacefeceeefc50f683aec7f7e9b787a",
        strip_prefix = "rules_nixpkgs-0.12.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1adb04dc0416915fef427757f4272c4f7dacefeceeefc50f683aec7f7e9b787a",
            "https://github.com/tweag/rules_nixpkgs/releases/download/v0.12.0/rules_nixpkgs-0.12.0.tar.gz",
        ],
        type = "tar.gz",
    )
