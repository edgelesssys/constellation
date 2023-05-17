"""multirun_deps"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def multirun_deps():
    http_archive(
        name = "com_github_ash2k_bazel_tools",
        sha256 = "a911dab6711bc12a00f02cc94b66ced7dc57650e382ebd4f17c9cdb8ec2cbd56",
        strip_prefix = "bazel-tools-2add5bb84c2837a82a44b57e83c7414247aed43a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a911dab6711bc12a00f02cc94b66ced7dc57650e382ebd4f17c9cdb8ec2cbd56",
            "https://github.com/ash2k/bazel-tools/archive/2add5bb84c2837a82a44b57e83c7414247aed43a.tar.gz",
        ],
        type = "tar.gz",
    )
