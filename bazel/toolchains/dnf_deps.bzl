"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "c6aecb167e41e923aeaa629443dabb7dc37462d96db928c3e91e2b019160e710",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c6aecb167e41e923aeaa629443dabb7dc37462d96db928c3e91e2b019160e710",
            "https://github.com/rmohr/bazeldnf/releases/download/v0.5.7/bazeldnf-v0.5.7.tar.gz",
        ],
        type = "tar.gz",
    )
