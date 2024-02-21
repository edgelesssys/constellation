"""buildifier repository rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def buildifier_deps():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        strip_prefix = "buildtools-6.4.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/05c3c3602d25aeda1e9dbc91d3b66e624c1f9fdadf273e5480b489e744ca7269",
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v6.4.0.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "05c3c3602d25aeda1e9dbc91d3b66e624c1f9fdadf273e5480b489e744ca7269",
    )
