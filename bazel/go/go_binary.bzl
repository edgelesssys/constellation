"""
This module contains rules and macros for building Go binaries.
"""

load("@io_bazel_rules_go//go:def.bzl", _go_binary = "go_binary")

def go_binary(**kwargs):
    """go_binary is a wrapper for go_binary that sets default values.

    Args:
          **kwargs: all arguments are passed to go_binary.
    """

    # use pure go net package "netgo"
    kwargs.setdefault("gotags", [])
    kwargs["gotags"].append("netgo")

    _go_binary(
        **kwargs
    )
