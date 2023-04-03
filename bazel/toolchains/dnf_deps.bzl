"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "904ecf035e73fe0988a73ffe2af83d213b18d885ad1f5cd0dc79aeb58340345e",
        strip_prefix = "bazeldnf-2498c863339d5cdf30d6637771784d934b8cb3da",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/904ecf035e73fe0988a73ffe2af83d213b18d885ad1f5cd0dc79aeb58340345e",
            "https://github.com/rmohr/bazeldnf/archive/2498c863339d5cdf30d6637771784d934b8cb3da.tar.gz",
        ],
        type = "tar.gz",
    )
