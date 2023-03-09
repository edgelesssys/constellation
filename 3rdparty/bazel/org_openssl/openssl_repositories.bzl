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
        sha256 = "24957cea0a43ee70bff9b78256698a0a11e89455483b30b4c41d1f6bfd20d269",
        strip_prefix = "rules_perl-db026ffa0d89fdff97de0c27199d84bd2c3b696c",
        urls = [
            "https://github.com/malt3/rules_perl/archive/db026ffa0d89fdff97de0c27199d84bd2c3b696c.tar.gz",
        ],
    )
