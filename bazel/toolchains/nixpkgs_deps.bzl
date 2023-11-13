"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def nixpkgs_deps():
    http_archive(
        name = "io_tweag_rules_nixpkgs",
        sha256 = "cf84628af3e4698acb200c005c4acf1dddaf5e7b9f839eeca78d983db2e874fb",
        strip_prefix = "rules_nixpkgs-2c767691d12b66a92f231bccb06bcf9f7477b962",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cf84628af3e4698acb200c005c4acf1dddaf5e7b9f839eeca78d983db2e874fb",
            "https://github.com/tweag/rules_nixpkgs/archive/2c767691d12b66a92f231bccb06bcf9f7477b962.tar.gz",
        ],
        type = "tar.gz",
    )
