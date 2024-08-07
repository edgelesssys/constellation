"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-2.0.0-beta1",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f70f07f9d0d6c275d7ec7d3c7f236d9b552ba3205e8f37df9c1125031cf967cc",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v2.0.0-beta1/rules_oci-v2.0.0-beta1.tar.gz",
        ],
        sha256 = "f70f07f9d0d6c275d7ec7d3c7f236d9b552ba3205e8f37df9c1125031cf967cc",
        patches = ["//bazel/toolchains:0001-disable-Windows-support.patch"],
        patch_args = ["-p1"],
    )
