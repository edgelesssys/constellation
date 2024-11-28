"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "1ce13c13a2f354fd37016d9fb333efeddcb308e89db9b3a8f45eafce57746f49",
        strip_prefix = "rules_nixpkgs-680f407bd4e523ec74a507503ffb619666fb51c3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ce13c13a2f354fd37016d9fb333efeddcb308e89db9b3a8f45eafce57746f49",
            "https://github.com/tweag/rules_nixpkgs/archive/680f407bd4e523ec74a507503ffb619666fb51c3.tar.gz",
        ],
        type = "tar.gz",
    )
