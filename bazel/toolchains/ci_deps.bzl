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
        name = "com_github_koalaman_shellcheck_linux_arm64",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/ad0c696c870c8525357b5127680cd79c0bdf58179af9acd091d43b1d6482da4a",
            "https://releases.hashicorp.com/terraform/1.5.5/terraform_1.5.5_linux_amd64.zip",
        ],
        sha256 = "ad0c696c870c8525357b5127680cd79c0bdf58179af9acd091d43b1d6482da4a",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_linux_arm64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b055aefe343d0b710d8a7afd31aeb702b37bbf4493bb9385a709991e48dfbcd2",
            "https://releases.hashicorp.com/terraform/1.5.5/terraform_1.5.5_linux_arm64.zip",
        ],
        sha256 = "b055aefe343d0b710d8a7afd31aeb702b37bbf4493bb9385a709991e48dfbcd2",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_amd64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6d61639e2141b7c23a9219c63994f729aa41f91110a1aa08b8a37969fb45e229",
            "https://releases.hashicorp.com/terraform/1.5.5/terraform_1.5.5_darwin_amd64.zip",
        ],
        sha256 = "6d61639e2141b7c23a9219c63994f729aa41f91110a1aa08b8a37969fb45e229",
        type = "zip",
    )
    http_archive(
        name = "com_github_hashicorp_terraform_darwin_arm64",
        build_file_content = """exports_files(["terraform"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c7fdeddb4739fdd5bada9d45fd786e2cbaf6e9e364693eee45c83e95281dad3a",
            "https://releases.hashicorp.com/terraform/1.5.5/terraform_1.5.5_darwin_arm64.zip",
        ],
        sha256 = "c7fdeddb4739fdd5bada9d45fd786e2cbaf6e9e364693eee45c83e95281dad3a",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/bdb57c353e2bbc43d2b097bb7289a6e65ef2526787f89316b4b452a9e5086be4",
            "https://github.com/mvdan/gofumpt/releases/download/v0.6.0/gofumpt_v0.6.0_linux_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "bdb57c353e2bbc43d2b097bb7289a6e65ef2526787f89316b4b452a9e5086be4",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/10ff2643b7b4b9425bda7f0ca2d4e54d85b09024fbfd9c21dbfd55017b907965",
            "https://github.com/mvdan/gofumpt/releases/download/v0.6.0/gofumpt_v0.6.0_linux_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "10ff2643b7b4b9425bda7f0ca2d4e54d85b09024fbfd9c21dbfd55017b907965",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/59e6047b3fa2fb65b60cf7f8be9b77cf6b31b428a9a24042ce29e85140868036",
            "https://github.com/mvdan/gofumpt/releases/download/v0.6.0/gofumpt_v0.6.0_darwin_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "59e6047b3fa2fb65b60cf7f8be9b77cf6b31b428a9a24042ce29e85140868036",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/725f7a922bf3f88bed6818a4553e80554cc5cdb67a72236a01707325aa2dbb7b",
            "https://github.com/mvdan/gofumpt/releases/download/v0.6.0/gofumpt_v0.6.0_darwin_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "725f7a922bf3f88bed6818a4553e80554cc5cdb67a72236a01707325aa2dbb7b",
    )

def _tfsec_deps():
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/21201f1615de0b4c143eba2da0f988bab3f68184646090b30ece1fdb501396ca",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.5/tfsec_1.28.5_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "21201f1615de0b4c143eba2da0f988bab3f68184646090b30ece1fdb501396ca",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a587a9cd879240074551067114b8f63d2249aab70cabf8f8d6884e2b67cfddad",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.5/tfsec_1.28.5_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a587a9cd879240074551067114b8f63d2249aab70cabf8f8d6884e2b67cfddad",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ad9a313d84aa11893672c7779a99f85a6acaab26c5a05ccc432db08bc4c0a37",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.5/tfsec_1.28.5_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "4ad9a313d84aa11893672c7779a99f85a6acaab26c5a05ccc432db08bc4c0a37",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/50117ac409bb3c1101453d74f48a639c7ab7ac2f40b023eb7004d84048913888",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.5/tfsec_1.28.5_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "50117ac409bb3c1101453d74f48a639c7ab7ac2f40b023eb7004d84048913888",
    )

def _golangci_lint_deps():
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e1c313fb5fc85a33890fdee5dbb1777d1f5829c84d655a47a55688f3aad5e501",
            "https://github.com/golangci/golangci-lint/releases/download/v1.56.2/golangci-lint-1.56.2-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.56.2-linux-amd64",
        type = "tar.gz",
        sha256 = "e1c313fb5fc85a33890fdee5dbb1777d1f5829c84d655a47a55688f3aad5e501",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0041594fde41ce43b75e65476a050fe9057881d8b5bccd472f18357e2ead3e04",
            "https://github.com/golangci/golangci-lint/releases/download/v1.56.2/golangci-lint-1.56.2-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.56.2-linux-arm64",
        type = "tar.gz",
        sha256 = "0041594fde41ce43b75e65476a050fe9057881d8b5bccd472f18357e2ead3e04",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15c4d19a2c85a04f67779047dbb9467ba176c71fff762a0d514a21bb75e4b42c",
            "https://github.com/golangci/golangci-lint/releases/download/v1.56.2/golangci-lint-1.56.2-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.56.2-darwin-amd64",
        type = "tar.gz",
        sha256 = "15c4d19a2c85a04f67779047dbb9467ba176c71fff762a0d514a21bb75e4b42c",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5f9ecda712c7ae08fbf872336fae3db866720e5865903d4c53903184b2a2c2dc",
            "https://github.com/golangci/golangci-lint/releases/download/v1.56.2/golangci-lint-1.56.2-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.56.2-darwin-arm64",
        type = "tar.gz",
        sha256 = "5f9ecda712c7ae08fbf872336fae3db866720e5865903d4c53903184b2a2c2dc",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1033f26361e6fc30ffcfab9d4e4274ffd4af88d9c97de63d2e1721c4a07c1380",
            "https://github.com/bufbuild/buf/releases/download/v1.29.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "1033f26361e6fc30ffcfab9d4e4274ffd4af88d9c97de63d2e1721c4a07c1380",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a725e0ab1c6b1e97b31f9d1d946f8b1d56586a96715fae4a7ecc88b6cf601cea",
            "https://github.com/bufbuild/buf/releases/download/v1.29.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a725e0ab1c6b1e97b31f9d1d946f8b1d56586a96715fae4a7ecc88b6cf601cea",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7ec6c2fd8f7e5e2ddba1e9ebff51eb9b0d6b67b85e105138dd064057c7b32db8",
            "https://github.com/bufbuild/buf/releases/download/v1.29.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7ec6c2fd8f7e5e2ddba1e9ebff51eb9b0d6b67b85e105138dd064057c7b32db8",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b97225a3b3f18bdabb36e83d9aba2e6419ead0c6ca0894d10a95517be5fd302f",
            "https://github.com/bufbuild/buf/releases/download/v1.29.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "b97225a3b3f18bdabb36e83d9aba2e6419ead0c6ca0894d10a95517be5fd302f",
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
        sha256 = "f43e1c3387de24547506ab05d24e5309c0ce0b228c23bd8aa64e9ec4b8206651",
        strip_prefix = "linux-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f43e1c3387de24547506ab05d24e5309c0ce0b228c23bd8aa64e9ec4b8206651",
            "https://get.helm.sh/helm-v3.14.0-linux-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_linux_arm64",
        sha256 = "b29e61674731b15f6ad3d1a3118a99d3cc2ab25a911aad1b8ac8c72d5a9d2952",
        strip_prefix = "linux-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b29e61674731b15f6ad3d1a3118a99d3cc2ab25a911aad1b8ac8c72d5a9d2952",
            "https://get.helm.sh/helm-v3.14.0-linux-arm64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_amd64",
        sha256 = "804586896496f7b3da97f56089ea00f220e075e969b6fdf6c0b7b9cdc22de120",
        strip_prefix = "darwin-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/804586896496f7b3da97f56089ea00f220e075e969b6fdf6c0b7b9cdc22de120",
            "https://get.helm.sh/helm-v3.14.0-darwin-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_arm64",
        sha256 = "c2f36f3289a01c7c93ca11f84d740a170e0af1d2d0280bd523a409a62b8dfa1d",
        strip_prefix = "darwin-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c2f36f3289a01c7c93ca11f84d740a170e0af1d2d0280bd523a409a62b8dfa1d",
            "https://get.helm.sh/helm-v3.14.0-darwin-arm64.tar.gz",
        ],
    )

def _ghh_deps():
    http_archive(
        name = "com_github_katexochen_ghh_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e53411fba3e3445bd08d4d7ce0ee9e526e9fcb12045616c80b9eb1cc342f0c90",
            "https://github.com/katexochen/ghh/releases/download/v0.3.3/ghh_0.3.3_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "e53411fba3e3445bd08d4d7ce0ee9e526e9fcb12045616c80b9eb1cc342f0c90",
    )
    http_archive(
        name = "com_github_katexochen_ghh_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96e8c073fca7b8b56490359b3df0386fac9138224699c71b33c14abc2373452b",
            "https://github.com/katexochen/ghh/releases/download/v0.3.3/ghh_0.3.3_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "96e8c073fca7b8b56490359b3df0386fac9138224699c71b33c14abc2373452b",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0fbffabe8f80c640403ef9b03348bd34e1f7de1321b0da48a36ae0403fabd09a",
            "https://github.com/katexochen/ghh/releases/download/v0.3.3/ghh_0.3.3_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "0fbffabe8f80c640403ef9b03348bd34e1f7de1321b0da48a36ae0403fabd09a",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32457fadf46c1b3e15c1caeadd176e8ff67152de744c83e7afaeb308bd514193",
            "https://github.com/katexochen/ghh/releases/download/v0.3.3/ghh_0.3.3_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "32457fadf46c1b3e15c1caeadd176e8ff67152de744c83e7afaeb308bd514193",
    )
