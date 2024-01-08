"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "d4a8c10121ec7494402a0ae8c1a896ced20d4bef4485b107e37f5331716c3626",
        strip_prefix = "rules_nixpkgs-244ae504d3f25534f6d3877ede4ee50e744a5234",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4a8c10121ec7494402a0ae8c1a896ced20d4bef4485b107e37f5331716c3626",
            "https://github.com/tweag/rules_nixpkgs/archive/244ae504d3f25534f6d3877ede4ee50e744a5234.tar.gz",
        ],
        type = "tar.gz",
    )
