"""CI dependencies"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

def ci_deps():
    """Install CI dependencies"""
    _shellcheck_deps()
    _terraform_deps()
    _actionlint_deps()
    _gofumpt_deps()
    _tfsec_deps()
    _golangci_lint_deps()
    _buf_deps()
    _talos_docgen_deps()
    _helm_deps()
    _ghh_deps()

def _shellcheck_deps():
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/700324c6dd0ebea0117591c6cc9d7350d9c7c5c287acbad7630fa17b1d4d9e2f",
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.linux.x86_64.tar.xz",
        ],
        sha256 = "700324c6dd0ebea0117591c6cc9d7350d9c7c5c287acbad7630fa17b1d4d9e2f",
        strip_prefix = "shellcheck-v0.9.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_aamd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/179c579ef3481317d130adebede74a34dbbc2df961a70916dd4039ebf0735fae",
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.linux.aarch64.tar.xz",
        ],
        sha256 = "179c579ef3481317d130adebede74a34dbbc2df961a70916dd4039ebf0735fae",
        strip_prefix = "shellcheck-v0.9.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7d3730694707605d6e60cec4efcb79a0632d61babc035aa16cda1b897536acf5",
            "https://github.com/koalaman/shellcheck/releases/download/v0.9.0/shellcheck-v0.9.0.darwin.x86_64.tar.xz",
        ],
        sha256 = "7d3730694707605d6e60cec4efcb79a0632d61babc035aa16cda1b897536acf5",
        strip_prefix = "shellcheck-v0.9.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
    )

def _terraform_deps():
    http_archive(
        name = "com_github_hashicorp_terraform_linux_amd64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f3ca33d04f5335472829d1df7785115b60176d610ae6f1583343b0a2221a931",
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_linux_amd64.zip",
        ],
        sha256 = "9f3ca33d04f5335472829d1df7785115b60176d610ae6f1583343b0a2221a931",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_linux_arm64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/39c182670c4e63e918e0a16080b1cc47bb16e158d7da96333d682d6a9cb8eb91",
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_linux_arm64.zip",
        ],
        sha256 = "39c182670c4e63e918e0a16080b1cc47bb16e158d7da96333d682d6a9cb8eb91",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_amd64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c218a6c0ef6692b25af16995c8c7bdf6739e9638fef9235c6aced3cd84afaf66",
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_darwin_amd64.zip",
        ],
        sha256 = "c218a6c0ef6692b25af16995c8c7bdf6739e9638fef9235c6aced3cd84afaf66",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_arm64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af8ff7576c8fc41496fdf97e9199b00d8d81729a6a0e821eaf4dfd08aa763540",
            "https://releases.hashicorp.com/terraform/1.4.2/terraform_1.4.2_darwin_arm64.zip",
        ],
        sha256 = "af8ff7576c8fc41496fdf97e9199b00d8d81729a6a0e821eaf4dfd08aa763540",
        type = "zip",
    )

def _actionlint_deps():
    http_archive(
        name = "com_github_rhysd_actionlint_linux_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f0294c342af98fad4ff917bc32032f28e1b55f76aedf291886ec10bbed7c12e1",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.26/actionlint_1.6.26_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f0294c342af98fad4ff917bc32032f28e1b55f76aedf291886ec10bbed7c12e1",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a1056d85d614af4f6e5517ed2911dab2621b8e97c368c8b265328f9c22801648",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.26/actionlint_1.6.26_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a1056d85d614af4f6e5517ed2911dab2621b8e97c368c8b265328f9c22801648",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bfa890e77a8508603c785af09a30bbab3a3255d291d8d27efc3f20ac8e303a8e",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.26/actionlint_1.6.26_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "bfa890e77a8508603c785af09a30bbab3a3255d291d8d27efc3f20ac8e303a8e",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5e131ab7de7ad051e1923b80d167aaa414734e97c720698c48778250e1dd2590",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.26/actionlint_1.6.26_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "5e131ab7de7ad051e1923b80d167aaa414734e97c720698c48778250e1dd2590",
    )

def _gofumpt_deps():
    http_file(
        name = "com_github_mvdan_gofumpt_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/759c6ab56bfbf62cafb35944aef1e0104a117e0aebfe44816fd79ef4b28521e4",
            "https://github.com/mvdan/gofumpt/releases/download/v0.5.0/gofumpt_v0.5.0_linux_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "759c6ab56bfbf62cafb35944aef1e0104a117e0aebfe44816fd79ef4b28521e4",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fba20ffd06606c89a500e3cc836408a09e4767e2f117c97724237ae4ecadf82e",
            "https://github.com/mvdan/gofumpt/releases/download/v0.5.0/gofumpt_v0.5.0_linux_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "fba20ffd06606c89a500e3cc836408a09e4767e2f117c97724237ae4ecadf82e",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/870f05a23541aad3d20d208a3ea17606169a240f608ac1cf987426198c14b2ed",
            "https://github.com/mvdan/gofumpt/releases/download/v0.5.0/gofumpt_v0.5.0_darwin_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "870f05a23541aad3d20d208a3ea17606169a240f608ac1cf987426198c14b2ed",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f2df95d5fad8498ad8eeb0be8abdb8bb8d05e8130b332cb69751dfd090fabac4",
            "https://github.com/mvdan/gofumpt/releases/download/v0.5.0/gofumpt_v0.5.0_darwin_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "f2df95d5fad8498ad8eeb0be8abdb8bb8d05e8130b332cb69751dfd090fabac4",
    )

def _tfsec_deps():
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/57b902b31da3eed12448a4e82a8aca30477e4bcd1bf99e3f65310eae0889f88d",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.1/tfsec_1.28.1_linux_amd64.tar.gz",
        ],
        sha256 = "57b902b31da3eed12448a4e82a8aca30477e4bcd1bf99e3f65310eae0889f88d",
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/20daad803d2a7a781f2ef0ee72ba4ed4ae17dcb41a43a330ae7b98347762bec9",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.1/tfsec_1.28.1_linux_arm64.tar.gz",
        ],
        sha256 = "20daad803d2a7a781f2ef0ee72ba4ed4ae17dcb41a43a330ae7b98347762bec9",
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6d9f5a747b1fcc1b6c314d30f4ff4d753371e5690309a99a5dd653d719d20d2d",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.1/tfsec_1.28.1_darwin_amd64.tar.gz",
        ],
        sha256 = "6d9f5a747b1fcc1b6c314d30f4ff4d753371e5690309a99a5dd653d719d20d2d",
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6d664dcdd37e2809d1b4f14b310ccda0973b4a29e4624e902286e4964d101e22",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.1/tfsec_1.28.1_darwin_arm64.tar.gz",
        ],
        sha256 = "6d664dcdd37e2809d1b4f14b310ccda0973b4a29e4624e902286e4964d101e22",
        type = "tar.gz",
    )

def _golangci_lint_deps():
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17c9ca05253efe833d47f38caf670aad2202b5e6515879a99873fabd4c7452b3",
            "https://github.com/golangci/golangci-lint/releases/download/v1.54.2/golangci-lint-1.54.2-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.54.2-linux-amd64",
        type = "tar.gz",
        sha256 = "17c9ca05253efe833d47f38caf670aad2202b5e6515879a99873fabd4c7452b3",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9f14b33473c65fcfbf411ec054b53a87dbb849f4e09ee438f1ee76dbf3f3d4e",
            "https://github.com/golangci/golangci-lint/releases/download/v1.54.2/golangci-lint-1.54.2-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.54.2-linux-arm64",
        type = "tar.gz",
        sha256 = "a9f14b33473c65fcfbf411ec054b53a87dbb849f4e09ee438f1ee76dbf3f3d4e",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/925c4097eae9e035b0b052a66d0a149f861e2ab611a4e677c7ffd2d4e05b9b89",
            "https://github.com/golangci/golangci-lint/releases/download/v1.54.2/golangci-lint-1.54.2-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.54.2-darwin-amd64",
        type = "tar.gz",
        sha256 = "925c4097eae9e035b0b052a66d0a149f861e2ab611a4e677c7ffd2d4e05b9b89",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7b33fb1be2f26b7e3d1f3c10ce9b2b5ce6d13bb1d8468a4b2ba794f05b4445e1",
            "https://github.com/golangci/golangci-lint/releases/download/v1.54.2/golangci-lint-1.54.2-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.54.2-darwin-arm64",
        type = "tar.gz",
        sha256 = "7b33fb1be2f26b7e3d1f3c10ce9b2b5ce6d13bb1d8468a4b2ba794f05b4445e1",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7286b1c6c84392f327991fd4c7b2e7f1bcff141cd1249e797a93d094c2f662ba",
            "https://github.com/bufbuild/buf/releases/download/v1.26.1/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7286b1c6c84392f327991fd4c7b2e7f1bcff141cd1249e797a93d094c2f662ba",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/939ed1b101a4919d86176750c146807e0e3a5f4d15611b11d3e126322eac6d9a",
            "https://github.com/bufbuild/buf/releases/download/v1.26.1/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "939ed1b101a4919d86176750c146807e0e3a5f4d15611b11d3e126322eac6d9a",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d7aa5bc29e7b4819dd138f1de170ff5180d7a6a5ea5f4005df5aad55e16f7143",
            "https://github.com/bufbuild/buf/releases/download/v1.26.1/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "d7aa5bc29e7b4819dd138f1de170ff5180d7a6a5ea5f4005df5aad55e16f7143",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de71605e1a3a9283b3c895accb33f405e050e4e88ded91131dedf9928188c1a6",
            "https://github.com/bufbuild/buf/releases/download/v1.26.1/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "de71605e1a3a9283b3c895accb33f405e050e4e88ded91131dedf9928188c1a6",
    )

def _talos_docgen_deps():
    http_file(
        name = "com_github_siderolabs_talos_hack_docgen_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bd1059b49a6db7473b4f991a244e338da887d8017ab556739000abd2cc367c13",
        ],
        executable = True,
        downloaded_file_path = "docgen",
        sha256 = "bd1059b49a6db7473b4f991a244e338da887d8017ab556739000abd2cc367c13",
    )
    http_file(
        name = "com_github_siderolabs_talos_hack_docgen_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d06adae41a975a94abaa39cd809464d7a8f7648903d321332c12c73002cc622a",
        ],
        executable = True,
        downloaded_file_path = "docgen",
        sha256 = "d06adae41a975a94abaa39cd809464d7a8f7648903d321332c12c73002cc622a",
    )
    http_file(
        name = "com_github_siderolabs_talos_hack_docgen_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a87c52a3e947fe90396427a5cd92e6864f46b5db103f84c1cad449e97ca54cec",
        ],
        executable = True,
        downloaded_file_path = "docgen",
        sha256 = "a87c52a3e947fe90396427a5cd92e6864f46b5db103f84c1cad449e97ca54cec",
    )
    http_file(
        name = "com_github_siderolabs_talos_hack_docgen_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4aa7ed0de31932d541aa11c9b75ed214ffc28dbd618f489fb5a598407aca072e",
        ],
        executable = True,
        downloaded_file_path = "docgen",
        sha256 = "4aa7ed0de31932d541aa11c9b75ed214ffc28dbd618f489fb5a598407aca072e",
    )

def _helm_deps():
    http_archive(
        name = "com_github_helm_helm_linux_amd64",
        sha256 = "781d826daec584f9d50a01f0f7dadfd25a3312217a14aa2fbb85107b014ac8ca",
        strip_prefix = "linux-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/781d826daec584f9d50a01f0f7dadfd25a3312217a14aa2fbb85107b014ac8ca",
            "https://get.helm.sh/helm-v3.11.2-linux-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_linux_arm64",
        sha256 = "0a60baac83c3106017666864e664f52a4e16fbd578ac009f9a85456a9241c5db",
        strip_prefix = "linux-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0a60baac83c3106017666864e664f52a4e16fbd578ac009f9a85456a9241c5db",
            "https://get.helm.sh/helm-v3.11.2-linux-arm64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_amd64",
        sha256 = "404938fd2c6eff9e0dab830b0db943fca9e1572cd3d7ee40904705760faa390f",
        strip_prefix = "darwin-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/404938fd2c6eff9e0dab830b0db943fca9e1572cd3d7ee40904705760faa390f",
            "https://get.helm.sh/helm-v3.11.2-darwin-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_arm64",
        sha256 = "f61a3aa55827de2d8c64a2063fd744b618b443ed063871b79f52069e90813151",
        strip_prefix = "darwin-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f61a3aa55827de2d8c64a2063fd744b618b443ed063871b79f52069e90813151",
            "https://get.helm.sh/helm-v3.11.2-darwin-arm64.tar.gz",
        ],
    )

def _ghh_deps():
    http_archive(
        name = "com_github_katexochen_ghh_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2f86968d4e11de5d7a4430fd973e41c71c25064d1e71dcc21f8c68174ecde522",
            "https://github.com/katexochen/ghh/releases/download/v0.3.1/ghh_0.3.1_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "2f86968d4e11de5d7a4430fd973e41c71c25064d1e71dcc21f8c68174ecde522",
    )
    http_archive(
        name = "com_github_katexochen_ghh_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e8ae94b120428f0b8d9085a05df8c61ba201ee3811fa1ebf1ef23955b2092635",
            "https://github.com/katexochen/ghh/releases/download/v0.3.1/ghh_0.3.1_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "e8ae94b120428f0b8d9085a05df8c61ba201ee3811fa1ebf1ef23955b2092635",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9b324ed7a516b0a6869cb8448ad5ef56bf0d8d5f6a8ad6b63d57e61fec570a2e",
            "https://github.com/katexochen/ghh/releases/download/v0.3.1/ghh_0.3.1_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "9b324ed7a516b0a6869cb8448ad5ef56bf0d8d5f6a8ad6b63d57e61fec570a2e",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bc8252749eb55607a0b32232c72d64742d49d5b2f187bf96c0f31c865e94fcc8",
            "https://github.com/katexochen/ghh/releases/download/v0.3.1/ghh_0.3.1_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "bc8252749eb55607a0b32232c72d64742d49d5b2f187bf96c0f31c865e94fcc8",
    )
