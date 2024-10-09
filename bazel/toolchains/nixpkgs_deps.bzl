"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "f2c927815c18c088f02ff81caf9903f9c0b2596ac6e6bd40534bc299af9dc0d7",
        strip_prefix = "rules_nixpkgs-705ee3b26cf49e990cddbbe6f60510fa46d50904",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f2c927815c18c088f02ff81caf9903f9c0b2596ac6e6bd40534bc299af9dc0d7",
            "https://github.com/tweag/rules_nixpkgs/archive/705ee3b26cf49e990cddbbe6f60510fa46d50904.tar.gz",
        ],
        type = "tar.gz",
    )
