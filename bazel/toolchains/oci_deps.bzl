"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-2.2.1",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cfea16076ebbec1faea494882ab97d94b1a62d6bcd5aceabad8f95ea0d0a1361",
            "https://github.com/bazel-contrib/rules_oci/releases/download/v2.2.1/rules_oci-v2.2.1.tar.gz",
        ],
        sha256 = "cfea16076ebbec1faea494882ab97d94b1a62d6bcd5aceabad8f95ea0d0a1361",
        patches = ["//bazel/toolchains:0001-disable-Windows-support.patch"],
        patch_args = ["-p1"],
    )
