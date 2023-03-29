"""multirun_deps"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def multirun_deps():
    http_archive(
        name = "com_github_ash2k_bazel_tools",
        sha256 = "0ad31a16c9e48b01a1a11daf908227a6bf6106269187cccf7398625fea2ba45a",
        strip_prefix = "bazel-tools-4e045b9b4e3e613970ab68941b556a356239d433",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0ad31a16c9e48b01a1a11daf908227a6bf6106269187cccf7398625fea2ba45a",
            "https://github.com/ash2k/bazel-tools/archive/4e045b9b4e3e613970ab68941b556a356239d433.tar.gz",
        ],
        type = "tar.gz",
    )
