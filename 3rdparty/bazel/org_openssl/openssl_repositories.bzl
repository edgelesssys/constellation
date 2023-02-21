"""A module defining the third party dependency OpenSSL"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

def openssl_repositories():
    maybe(
        http_archive,
        name = "org_openssl",
        build_file = Label("//3rdparty/bazel/org_openssl:BUILD.openssl.bazel"),
        sha256 = "c5ac01e760ee6ff0dab61d6b2bbd30146724d063eb322180c6f18a6f74e4b6aa",
        strip_prefix = "openssl-1.1.1s",
        urls = [
            "https://www.openssl.org/source/openssl-1.1.1s.tar.gz",
            "https://github.com/openssl/openssl/archive/OpenSSL_1_1_1s.tar.gz",
        ],
    )

    maybe(
        http_archive,
        name = "nasm",
        build_file = Label("//3rdparty/bazel/org_openssl:BUILD.nasm.bazel"),
        sha256 = "f5c93c146f52b4f1664fa3ce6579f961a910e869ab0dae431bd871bdd2584ef2",
        strip_prefix = "nasm-2.15.05",
        urls = [
            "https://mirror.bazel.build/www.nasm.us/pub/nasm/releasebuilds/2.15.05/win64/nasm-2.15.05-win64.zip",
            "https://www.nasm.us/pub/nasm/releasebuilds/2.15.05/win64/nasm-2.15.05-win64.zip",
        ],
    )

    maybe(
        http_archive,
        name = "rules_perl",
        sha256 = "391edb08802860ba733d402c6376cfe1002b598b90d2240d9d302ecce2289a64",
        strip_prefix = "rules_perl-7f10dada09fcba1dc79a6a91da2facc25e72bd7d",
        urls = [
            "https://github.com/bazelbuild/rules_perl/archive/7f10dada09fcba1dc79a6a91da2facc25e72bd7d.tar.gz",
        ],
    )
