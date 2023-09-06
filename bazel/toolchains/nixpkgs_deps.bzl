"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "1eb0ef4f8388a9b53f99587a6282250852a03eab3160313e584ada7005a7ab53",
        strip_prefix = "rules_nixpkgs-ee5c44a5bc470d13699960f493d42ed4c4d806af",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1eb0ef4f8388a9b53f99587a6282250852a03eab3160313e584ada7005a7ab53",
            "https://github.com/tweag/rules_nixpkgs/archive/ee5c44a5bc470d13699960f493d42ed4c4d806af.tar.gz",
        ],
        type = "tar.gz",
    )
