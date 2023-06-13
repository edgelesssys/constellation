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
            "https://cdn.confidential.cloud/constellation/cas/sha256/3c5818744143a5d6754edd3dcc4c2b32c9dfcdd3bb30e0e108fb5e5c505262d4",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.24/actionlint_1.6.24_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "3c5818744143a5d6754edd3dcc4c2b32c9dfcdd3bb30e0e108fb5e5c505262d4",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93cc9d1f4a01f0658423b41ecf3bd8c17c619003ec683be8bac9264d0361d0d8",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.24/actionlint_1.6.24_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "93cc9d1f4a01f0658423b41ecf3bd8c17c619003ec683be8bac9264d0361d0d8",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ce9dd9653700e3ed00464dffddd3e2a61358cf96425f2f3dff840dfc1e105eab",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.24/actionlint_1.6.24_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "ce9dd9653700e3ed00464dffddd3e2a61358cf96425f2f3dff840dfc1e105eab",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5477f8a5a4073ef086525a2512b2bf1201641cd544034ad0c66f329590638242",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.24/actionlint_1.6.24_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "5477f8a5a4073ef086525a2512b2bf1201641cd544034ad0c66f329590638242",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/c9cf72d12058a131746edd409ed94ccd578fbd178899d1ed41ceae3ce5f54501",
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.52.2-linux-amd64",
        type = "tar.gz",
        sha256 = "c9cf72d12058a131746edd409ed94ccd578fbd178899d1ed41ceae3ce5f54501",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fc09a97f8888809fab83a316f7da70c8ed74d4863b7eed7d872cec41911a55e8",
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.52.2-linux-arm64",
        type = "tar.gz",
        sha256 = "fc09a97f8888809fab83a316f7da70c8ed74d4863b7eed7d872cec41911a55e8",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e57f2599de73c4da1d36d5255b9baec63f448b3d7fb726ebd3cd64dabbd3ee4a",
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.52.2-darwin-amd64",
        type = "tar.gz",
        sha256 = "e57f2599de73c4da1d36d5255b9baec63f448b3d7fb726ebd3cd64dabbd3ee4a",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/89e523d45883903cfc472ab65621073f850abd4ffbb7720bbdd7ba66ee490bc8",
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.52.2-darwin-arm64",
        type = "tar.gz",
        sha256 = "89e523d45883903cfc472ab65621073f850abd4ffbb7720bbdd7ba66ee490bc8",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f79670efc64d0527e0b915a272aea8262b4864ad9298e8d1cf39b7b08517607c",
            "https://github.com/bufbuild/buf/releases/download/v1.21.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f79670efc64d0527e0b915a272aea8262b4864ad9298e8d1cf39b7b08517607c",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7ae3d4bf6ff121172dd949bd4d25342e03a0b7f10cf8e8ccdc8f98a664b79794",
            "https://github.com/bufbuild/buf/releases/download/v1.21.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7ae3d4bf6ff121172dd949bd4d25342e03a0b7f10cf8e8ccdc8f98a664b79794",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a9a9af802630d7547dfbd79727c462ec7acd4978b91b922957438d4aec99ac9",
            "https://github.com/bufbuild/buf/releases/download/v1.21.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "4a9a9af802630d7547dfbd79727c462ec7acd4978b91b922957438d4aec99ac9",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8b5d0285b11c14abd17fa8d76049e5ba90e8776784cc57aa0af77052ee335e99",
            "https://github.com/bufbuild/buf/releases/download/v1.21.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "8b5d0285b11c14abd17fa8d76049e5ba90e8776784cc57aa0af77052ee335e99",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/3d8f4c4f1aa59b4f983afe87131bf574738973bd18bade8a164dee6bdca22bf8",
            "https://github.com/katexochen/ghh/releases/download/v0.2.1/ghh_0.2.1_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "3d8f4c4f1aa59b4f983afe87131bf574738973bd18bade8a164dee6bdca22bf8",
    )
    http_archive(
        name = "com_github_katexochen_ghh_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c7a13bb1f32ff984d10246159db003290d4008154b0e92615c7067a180cc8c35",
            "https://github.com/katexochen/ghh/releases/download/v0.2.1/ghh_0.2.1_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "c7a13bb1f32ff984d10246159db003290d4008154b0e92615c7067a180cc8c35",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4caf11df2ab8d26222c6db35a44f53b5d8bfd0d8b12c79a0b9f8469efd0acfc",
            "https://github.com/katexochen/ghh/releases/download/v0.2.1/ghh_0.2.1_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "d4caf11df2ab8d26222c6db35a44f53b5d8bfd0d8b12c79a0b9f8469efd0acfc",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7601f0926ba0df7ad36ac05e20a60347f97ac443262dc691bc5060f27d727b39",
            "https://github.com/katexochen/ghh/releases/download/v0.2.1/ghh_0.2.1_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "7601f0926ba0df7ad36ac05e20a60347f97ac443262dc691bc5060f27d727b39",
    )
