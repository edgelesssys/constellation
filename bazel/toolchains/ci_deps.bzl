"""CI dependencies"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def ci_deps():
    _shellcheck_deps()

def _shellcheck_deps():
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_x86_64",
        urls = [
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.linux.x86_64.tar.xz",
        ],
        sha256 = "700324c6dd0ebea0117591c6cc9d7350d9c7c5c287acbad7630fa17b1d4d9e2f",
        strip_prefix = "shellcheck-v0.9.0",
        build_file = "//bazel/toolchains:BUILD.shellcheck.bazel",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_aaarch64",
        urls = [
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.linux.aarch64.tar.xz",
        ],
        strip_prefix = "shellcheck-v0.9.0",
        build_file = "//bazel/toolchains:BUILD.shellcheck.bazel",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_darwin_x86_64",
        urls = [
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.darwin.x86_64.tar.xz",
        ],
        strip_prefix = "shellcheck-v0.9.0",
        build_file = "//bazel/toolchains:BUILD.shellcheck.bazel",
    )
