"""multirun_deps"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def multirun_deps():
    http_archive(
        name = "com_github_ash2k_bazel_tools",
        sha256 = "dc32a65c69c843f1ba2a328b79974163896e5b8ed283cd711abe12bf7cd12ffc",
        strip_prefix = "bazel-tools-415483a9e13342a6603a710b0296f6d85b8d26bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc32a65c69c843f1ba2a328b79974163896e5b8ed283cd711abe12bf7cd12ffc",
            "https://github.com/ash2k/bazel-tools/archive/415483a9e13342a6603a710b0296f6d85b8d26bf.tar.gz",
        ],
        type = "tar.gz",
    )
