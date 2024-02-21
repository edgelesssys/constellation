"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        strip_prefix = "rules_python-0.31.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c68bdc4fbec25de5b5493b8819cfc877c4ea299c0dcb15c244c5a00208cde311",
            "https://github.com/bazelbuild/rules_python/releases/download/0.31.0/rules_python-0.31.0.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "c68bdc4fbec25de5b5493b8819cfc877c4ea299c0dcb15c244c5a00208cde311",
    )
