""" Bazel cc_library definitions for Nixpkgs. """

load("@io_tweag_rules_nixpkgs//nixpkgs:nixpkgs.bzl", "nixpkgs_flake_package")

def nixpkgs_cc_library_deps():
    """ Generate cc_library rules for Nixpkgs. """
    return [
        nixpkgs_flake_package(
            name = "org_openssl_%s" % system,
            nix_flake_file = "//:flake.nix",
            nix_flake_lock_file = "//:flake.lock",
            package = "packages.%s.openssl" % system,
            build_file_content = OPENSSL_BUILD,
        )
        for system in openssl_systems
    ] + [
        nixpkgs_flake_package(
            name = "cryptsetup_%s" % system,
            nix_flake_file = "//:flake.nix",
            nix_flake_lock_file = "//:flake.lock",
            package = "cryptsetup",
            build_file_content = CRYPTSETUP_BUILD,
        )
        for system in cryptsetup_systems
    ] + [
        nixpkgs_flake_package(
            name = "libvirt_%s" % system,
            nix_flake_file = "//:flake.nix",
            nix_flake_lock_file = "//:flake.lock",
            package = "libvirt",
            build_file_content = LIBVIRT_BUILD,
        )
        for system in libvirt_systems
    ]

openssl_systems = [
    "aarch64-linux",
    "aarch64-darwin",
    "x86_64-linux",
    "x86_64-darwin",
]

cryptsetup_systems = [
    "x86_64-linux",
]

libvirt_systems = [
    "x86_64-linux",
]

OPENSSL_BUILD = """\
load("@rules_cc//cc:defs.bzl", "cc_library")
filegroup(
    name = "include",
    srcs = glob(["include/**/*.h"]),
    visibility = ["//visibility:public"],
)
cc_library(
    name = "org_openssl",
    srcs = glob(["lib/**/*.a"]),
    hdrs = [":include"],
    strip_include_prefix = "include",
    visibility = ["//visibility:public"],
)
"""

CRYPTSETUP_BUILD = """\
exports_files(["closure.tar", "rpath", "dynamic-linker"])
filegroup(
    name = "include",
    srcs = glob(["include/**/*.h"]),
    visibility = ["//visibility:public"],
)
cc_library(
    name = "cryptsetup",
    srcs = glob(["lib/**/*.so*"]),
    hdrs = [":include"],
    strip_include_prefix = "include",
    target_compatible_with = [
        "@platforms//os:linux",
        "@platforms//cpu:x86_64",
    ],
    visibility = ["//visibility:public"],
)
"""

LIBVIRT_BUILD = """\
exports_files(["bin-linktree.tar", "closure.tar", "rpath", "dynamic-linker"])
load("@rules_cc//cc:defs.bzl", "cc_library")
filegroup(
    name = "include",
    srcs = glob(["include/**/*.h"]),
    visibility = ["//visibility:public"],
)
cc_library(
    name = "libvirt",
    srcs = glob([
        "lib/*.so",
        "lib/*.so.*",
    ]),
    hdrs = [":include"],
    strip_include_prefix = "include",
    target_compatible_with = [
        "@platforms//os:linux",
        "@platforms//cpu:x86_64",
    ],
    visibility = ["//visibility:public"],
)
"""
