"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "97c43034d55af16061adcae26e69588bc2e164cbc289a140a3bb5b2125f0109d",
        strip_prefix = "bazeldnf-32db3eee870531104529711782da1aa4eca7dacf",
        urls = [
            "https://github.com/rmohr/bazeldnf/archive/32db3eee870531104529711782da1aa4eca7dacf.tar.gz",
        ],
    )
