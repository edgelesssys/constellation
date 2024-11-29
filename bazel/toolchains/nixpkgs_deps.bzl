"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "1ce13c13a2f354fd37016d9fb333efeddcb308e89db9b3a8f45eafce57746f49",
        strip_prefix = "rules_nixpkgs-668609f0b3627751651cb325166d0e95062be3f7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ce13c13a2f354fd37016d9fb333efeddcb308e89db9b3a8f45eafce57746f49",
            "https://github.com/tweag/rules_nixpkgs/archive/668609f0b3627751651cb325166d0e95062be3f7.tar.gz",
        ],
        type = "tar.gz",
    )
