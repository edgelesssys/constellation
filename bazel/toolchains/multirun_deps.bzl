"""multirun_deps"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def multirun_deps():
    http_archive(
        name = "com_github_ash2k_bazel_tools",
        sha256 = "c5c2bb097ef427ab021f522828167c6d85c3e9077763629343282c51dbde03db",
        strip_prefix = "bazel-tools-415483a9e13342a6603a710b0296f6d85b8d26bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c5c2bb097ef427ab021f522828167c6d85c3e9077763629343282c51dbde03db",
            "https://github.com/ash2k/bazel-tools/archive/415483a9e13342a6603a710b0296f6d85b8d26bf.tar.gz",
        ],
        type = "tar.gz",
    )
