"""python toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def python_deps():
    http_archive(
        name = "rules_python",
        strip_prefix = "rules_python-0.25.0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5868e73107a8e85d8f323806e60cad7283f34b32163ea6ff1020cf27abef6036",
            "https://github.com/bazelbuild/rules_python/releases/download/0.25.0/rules_python-0.25.0.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "5868e73107a8e85d8f323806e60cad7283f34b32163ea6ff1020cf27abef6036",
    )
