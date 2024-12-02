"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "0f7ac344873873d89f8286b4971401dca4a1b249421c2f7c7b56a1befe4501cb",
        strip_prefix = "rules_nixpkgs-4e169eefd4e9cf38a3569e0b463e93f0c7a3c181",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f7ac344873873d89f8286b4971401dca4a1b249421c2f7c7b56a1befe4501cb",
            "https://github.com/tweag/rules_nixpkgs/archive/4e169eefd4e9cf38a3569e0b463e93f0c7a3c181.tar.gz",
        ],
        type = "tar.gz",
    )
