"""Bazel rules for building OCI images"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def oci_deps():
    # TODO(malt3): This uses a patch on top of the normal rules_oci that removes broken Windows support.
    # Remove this override once https://github.com/bazel-contrib/rules_oci/issues/420 is fixed.
    http_archive(
        name = "rules_oci",
        strip_prefix = "rules_oci-21c48973f0e79744a617ba5f2014c5d5fca90f17",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dca0cfa2a8eb4ab79c231617964fc821f6d1a3bb9d996358975a5ceee5b8d25f",
            "https://github.com/bazel-contrib/rules_oci/archive/21c48973f0e79744a617ba5f2014c5d5fca90f17.tar.gz",
        ],
        sha256 = "dca0cfa2a8eb4ab79c231617964fc821f6d1a3bb9d996358975a5ceee5b8d25f",
    )
