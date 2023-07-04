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
            "https://cdn.confidential.cloud/constellation/cas/sha256/80a54660e73ad55a0818372bdaa0dced82eb86f618e6bf1621e73f099e61c027",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.25/actionlint_1.6.25_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "80a54660e73ad55a0818372bdaa0dced82eb86f618e6bf1621e73f099e61c027",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8bedeea8ed636891fd7351fa7ccbc75fdb5bee6efde5320162f712e8457e73ea",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.25/actionlint_1.6.25_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "8bedeea8ed636891fd7351fa7ccbc75fdb5bee6efde5320162f712e8457e73ea",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/30d69622ff9fbf564081515bf7d20538f2cb590150ef0c69fdcc56fa23fe85f1",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.25/actionlint_1.6.25_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "30d69622ff9fbf564081515bf7d20538f2cb590150ef0c69fdcc56fa23fe85f1",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9153ebe7be2a33c9047e60aeb0d8d7b831b22fe99bbea63d365500c68245d6df",
            "https://github.com/rhysd/actionlint/releases/download/v1.6.25/actionlint_1.6.25_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "9153ebe7be2a33c9047e60aeb0d8d7b831b22fe99bbea63d365500c68245d6df",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/2298f73b9bc03b88b91fee06c5d519fc7f9d7f328e2c388615bbd7e85a9d6cae",
            "https://github.com/golangci/golangci-lint/releases/download/v1.53.2/golangci-lint-1.53.2-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.53.2-linux-amd64",
        type = "tar.gz",
        sha256 = "2298f73b9bc03b88b91fee06c5d519fc7f9d7f328e2c388615bbd7e85a9d6cae",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c9308fac5217ab83de0966291e119f9643cd185ca9901fde9c67f61641da49e5",
            "https://github.com/golangci/golangci-lint/releases/download/v1.53.2/golangci-lint-1.53.2-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.53.2-linux-arm64",
        type = "tar.gz",
        sha256 = "c9308fac5217ab83de0966291e119f9643cd185ca9901fde9c67f61641da49e5",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a4e83f5bfe52f42134c9783aa68ba31104c36e2ad4c221a3c77510dda66ae81c",
            "https://github.com/golangci/golangci-lint/releases/download/v1.53.2/golangci-lint-1.53.2-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.53.2-darwin-amd64",
        type = "tar.gz",
        sha256 = "a4e83f5bfe52f42134c9783aa68ba31104c36e2ad4c221a3c77510dda66ae81c",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/63f6c3dab479dac20f39d4a91c4a2c019c054968e2d044f03ee47a3e41253623",
            "https://github.com/golangci/golangci-lint/releases/download/v1.53.2/golangci-lint-1.53.2-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.53.2-darwin-arm64",
        type = "tar.gz",
        sha256 = "63f6c3dab479dac20f39d4a91c4a2c019c054968e2d044f03ee47a3e41253623",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1db51318e49f12095c97866c9b5d939dfec318b50362bba8a3a9545c4cff456b",
            "https://github.com/bufbuild/buf/releases/download/v1.23.1/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "1db51318e49f12095c97866c9b5d939dfec318b50362bba8a3a9545c4cff456b",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/25771076a40744840dcf23b7fc95b50c84687492e5730fbea1330d33693f55cf",
            "https://github.com/bufbuild/buf/releases/download/v1.23.1/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "25771076a40744840dcf23b7fc95b50c84687492e5730fbea1330d33693f55cf",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c1c0c2d045ce3aee1db3450014a7d8b978acae38a643d9319233c81c0f065df",
            "https://github.com/bufbuild/buf/releases/download/v1.23.1/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7c1c0c2d045ce3aee1db3450014a7d8b978acae38a643d9319233c81c0f065df",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e815901dac2384e9a6ca3f404e989ed1b4815e1ba7b986926af8bd151c68a710",
            "https://github.com/bufbuild/buf/releases/download/v1.23.1/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "e815901dac2384e9a6ca3f404e989ed1b4815e1ba7b986926af8bd151c68a710",
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
