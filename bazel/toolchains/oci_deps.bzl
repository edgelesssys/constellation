"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-2.2.6",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/361c417e8c95cd7c3d8b5cf4b202e76bac8d41532131534ff8e6fa43aa161142",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v2.2.6/rules_oci-v2.2.6.tar.gz",
        ],
        sha256 = "361c417e8c95cd7c3d8b5cf4b202e76bac8d41532131534ff8e6fa43aa161142",
        patches = ["//bazel/toolchains:0001-disable-Windows-support.patch"],
        patch_args = ["-p1"],
    )
