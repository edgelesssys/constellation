"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "871dfea4602bf4245aed4f15f8191e4b53b27dace20216862ede80fb893c6bf5",
        strip_prefix = "bazeldnf-ecdd4b69be56bf4924a7f8fc2197761111d815c8",
        urls = [
            "https://github.com/rmohr/bazeldnf/archive/ecdd4b69be56bf4924a7f8fc2197761111d815c8.tar.gz",
        ],
    )
