"""multirun_deps"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

def multirun_deps():
    git_repository(
        name = "com_github_ash2k_bazel_tools",
        commit = "4e045b9b4e3e613970ab68941b556a356239d433",
        remote = "https://github.com/ash2k/bazel-tools.git",
    )
