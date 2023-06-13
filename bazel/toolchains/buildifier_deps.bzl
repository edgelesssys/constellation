"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        strip_prefix = "buildtools-6.1.2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d368c47bbfc055010f118efb2962987475418737e901f7782d2a966d1dc80296",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v6.1.2.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "d368c47bbfc055010f118efb2962987475418737e901f7782d2a966d1dc80296",
    )
