"""
This module contains rules and macros for building and testing Go code.
"""

load("@io_bazel_rules_go//go:def.bzl", _go_test = "go_test")

def go_test(ld = None, count = 3, **kwargs):
    """go_test is a wrapper for go_test that uses default settings for Constellation.

    It adds the following:
    - Sets test count to 3.
    - Sets race detector to on by default (except Mac OS)
    - Optionally sets the interpreter path to ld.

    Args:
          ld: path to interpreter to that will be written into the elf header.
          count: number of times each test should be executed. defaults to 3.
          **kwargs: all other arguments are passed to go_test.
    """

    # Sets test count to 3.
    kwargs.setdefault("args", [])
    kwargs["args"].append("--test.count={}".format(count))

    # enable race detector by default
    race_value = select({
        "@platforms//os:macos": "off",
        "//conditions:default": "on",
    })
    pure_value = select({
        "@platforms//os:macos": "on",
        "//conditions:default": "off",
    })
    kwargs.setdefault("race", race_value)
    kwargs.setdefault("pure", pure_value)

    # set gc_linkopts to set the interpreter path to ld.
    kwargs.setdefault("gc_linkopts", [])
    if ld:
        kwargs["gc_linkopts"] += ["-I", ld]

    _go_test(**kwargs)
