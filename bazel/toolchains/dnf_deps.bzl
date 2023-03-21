"""bazeldnf is a Bazel extension to manage RPM packages for use within bazel"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def dnf_deps():
    http_archive(
        name = "bazeldnf",
        sha256 = "6104de1d657ae524bef5af86b153b82f114f532fe2e7eb02beb2e950550a88fe",
        strip_prefix = "bazeldnf-45f5d74ba73710b538c57c9d43d88c583aab9d3a",
        urls = [
            "https://github.com/rmohr/bazeldnf/archive/45f5d74ba73710b538c57c9d43d88c583aab9d3a.tar.gz",
        ],
    )
