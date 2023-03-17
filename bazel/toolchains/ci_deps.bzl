"""CI dependencies"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def ci_deps():
    """Install CI dependencies"""
    _shellcheck_deps()
    _terraform_deps()

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

def _terraform_deps():
    http_archive(
        name = "com_github_hashicorp_terraform_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.terraform.bazel",
        urls = [
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_linux_amd64.zip",
        ],
        sha256 = "9f3ca33d04f5335472829d1df7785115b60176d610ae6f1583343b0a2221a931",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.terraform.bazel",
        urls = [
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_linux_arm64.zip",
        ],
        sha256 = "39c182670c4e63e918e0a16080b1cc47bb16e158d7da96333d682d6a9cb8eb91",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.terraform.bazel",
        urls = [
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_darwin_amd64.zip",
        ],
        sha256 = "c218a6c0ef6692b25af16995c8c7bdf6739e9638fef9235c6aced3cd84afaf66",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.terraform.bazel",
        urls = [
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_darwin_arm64.zip",
        ],
        sha256 = "af8ff7576c8fc41496fdf97e9199b00d8d81729a6a0e821eaf4dfd08aa763540",
    )
