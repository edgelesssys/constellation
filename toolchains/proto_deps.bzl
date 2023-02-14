"""proto toolchain rules"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_deps():
    http_archive(
        name = "rules_proto",
        sha256 = "bc12122a5ae4b517fa423ea03a8d82ea6352d5127ea48cb54bc324e8ab78493c",
        strip_prefix = "rules_proto-af6481970a34554c6942d993e194a9aed7987780",
        urls = [
            "https://github.com/bazelbuild/rules_proto/archive/af6481970a34554c6942d993e194a9aed7987780.tar.gz",
            "https://storage.googleapis.com/builddeps/bc12122a5ae4b517fa423ea03a8d82ea6352d5127ea48cb54bc324e8ab78493c",
        ],
    )
