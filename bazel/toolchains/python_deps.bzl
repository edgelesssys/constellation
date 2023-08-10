"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        strip_prefix = "rules_python-0.24.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0a8003b044294d7840ac7d9d73eef05d6ceb682d7516781a4ec62eeb34702578",
            "https://github.com/bazelbuild/rules_python/releases/download/0.24.0/rules_python-0.24.0.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "0a8003b044294d7840ac7d9d73eef05d6ceb682d7516781a4ec62eeb34702578",
    )
