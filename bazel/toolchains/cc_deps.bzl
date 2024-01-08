"""bazel rules_cc"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def rules_cc_deps():
    http_archive(
        name = "rules_cc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2037875b9a4456dce4a79d112a8ae885bbc4aad968e6587dca6e64f3a0900cdf",
            "https://github.com/bazelbuild/rules_cc/releases/download/0.0.9/rules_cc-0.0.9.tar.gz",
        ],
        sha256 = "2037875b9a4456dce4a79d112a8ae885bbc4aad968e6587dca6e64f3a0900cdf",
        strip_prefix = "rules_cc-0.0.9",
        type = "tar.gz",
    )
