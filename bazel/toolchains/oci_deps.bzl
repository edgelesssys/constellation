"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-3d43cb1a1bb2f5edc15c7f48b406be3fb225e673",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ba75ecc92504355bf519786e3f59506ba43b59038408c7d43f50f30757caf09f",
            "https://github.com/bazel-contrib/rules_oci/archive/3d43cb1a1bb2f5edc15c7f48b406be3fb225e673.tar.gz",
        ],
        sha256 = "ba75ecc92504355bf519786e3f59506ba43b59038408c7d43f50f30757caf09f",
        patches = ["//bazel/toolchains:oci_deps.patch"],
        patch_args = ["-p1"],
    )
