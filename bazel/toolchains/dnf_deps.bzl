"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "b10513edd9cb57717f538a928922b57cf4ee96f0cc4e37312e805a354222abb4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b10513edd9cb57717f538a928922b57cf4ee96f0cc4e37312e805a354222abb4",
            "https://github.com/malt3/bazeldnf/releases/download/v0.5.8-rc0-malt3/bazeldnf-v0.5.8-rc0-malt3.tar.gz",
        ],
        type = "tar.gz",
    )
