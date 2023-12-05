"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        strip_prefix = "rules_python-0.27.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e85ae30de33625a63eca7fc40a94fea845e641888e52f32b6beea91e8b1b2793",
            "https://github.com/bazelbuild/rules_python/releases/download/0.27.1/rules_python-0.27.1.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "e85ae30de33625a63eca7fc40a94fea845e641888e52f32b6beea91e8b1b2793",
    )
