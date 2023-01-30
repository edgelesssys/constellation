"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "ddbb105d5c3143676e00eaf0a6f4e0789c7d8c1697d13cda8abb2ac82930e9b6",
        strip_prefix = "bazeldnf-92557ee30f69f06838e660f4774600be549c2ba7",
        urls = [
            "https://github.com/rmohr/bazeldnf/archive/92557ee30f69f06838e660f4774600be549c2ba7.tar.gz",
        ],
    )
