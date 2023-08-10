"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        strip_prefix = "rules_python-0.24.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/84aec9e21cc56fbc7f1335035a71c850d1b9b5cc6ff497306f84cced9a769841",
            "https://github.com/bazelbuild/rules_python/releases/download/0.24.0/rules_python-0.24.0.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "84aec9e21cc56fbc7f1335035a71c850d1b9b5cc6ff497306f84cced9a769841",
    )
