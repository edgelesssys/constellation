"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        sha256 = "94750828b18044533e98a129003b6a68001204038dc4749f40b195b24c38f49f",
        strip_prefix = "rules_python-0.23.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/94750828b18044533e98a129003b6a68001204038dc4749f40b195b24c38f49f",
            "https://github.com/bazelbuild/rules_python/releases/download/0.23.1/rules_python-0.23.1.tar.gz",
        ],
        type = "tar.gz",
    )
