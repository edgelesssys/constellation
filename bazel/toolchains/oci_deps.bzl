"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-32139d8e882710c3472030e804711a4835e4428b",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/48603628642552edab84ba0ae343ec0039452f0975751ab0cb4f3cf3aa080978",
            "https://github.com/bazel-contrib/rules_oci/archive/32139d8e882710c3472030e804711a4835e4428b.tar.gz",
        ],
        sha256 = "48603628642552edab84ba0ae343ec0039452f0975751ab0cb4f3cf3aa080978",
        patches = ["//bazel/toolchains:oci_deps.patch"],
        patch_args = ["-p1"],
    )
