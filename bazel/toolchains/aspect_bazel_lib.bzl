load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aspect_bazel_lib():
    http_archive(
        name = "aspect_bazel_lib",
        sha256 = "4b32cf6feab38b887941db022020eea5a49b848e11e3d6d4d18433594951717a",
        strip_prefix = "bazel-lib-2.0.1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4b32cf6feab38b887941db022020eea5a49b848e11e3d6d4d18433594951717a",
            "https://github.com/aspect-build/bazel-lib/releases/download/v2.0.1/bazel-lib-v2.0.1.tar.gz",
        ],
        type = "tar.gz",
    )
