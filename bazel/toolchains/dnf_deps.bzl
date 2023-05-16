"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "904ecf035e73fe0988a73ffe2af83d213b18d885ad1f5cd0dc79aeb58340345e",
        strip_prefix = "bazeldnf-6a9f5247f6fea6aabf9a4538305e1ab6d2a4356a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/904ecf035e73fe0988a73ffe2af83d213b18d885ad1f5cd0dc79aeb58340345e",
            "https://github.com/rmohr/bazeldnf/archive/6a9f5247f6fea6aabf9a4538305e1ab6d2a4356a.tar.gz",
        ],
        type = "tar.gz",
    )
