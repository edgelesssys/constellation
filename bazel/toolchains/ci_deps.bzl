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
            "https://cdn.confidential.cloud/constellation/cas/sha256/6c881ab0698e4e6ea235245f22832860544f17ba386442fe7e9d629f8cbedf87",
            "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.x86_64.tar.xz",
        ],
        sha256 = "6c881ab0698e4e6ea235245f22832860544f17ba386442fe7e9d629f8cbedf87",
        strip_prefix = "shellcheck-v0.10.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/324a7e89de8fa2aed0d0c28f3dab59cf84c6d74264022c00c22af665ed1a09bb",
            "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.linux.aarch64.tar.xz",
        ],
        sha256 = "324a7e89de8fa2aed0d0c28f3dab59cf84c6d74264022c00c22af665ed1a09bb",
        strip_prefix = "shellcheck-v0.10.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ef27684f23279d112d8ad84e0823642e43f838993bbb8c0963db9b58a90464c2",
            "https://github.com/koalaman/shellcheck/releases/download/v0.10.0/shellcheck-v0.10.0.darwin.x86_64.tar.xz",
        ],
        sha256 = "ef27684f23279d112d8ad84e0823642e43f838993bbb8c0963db9b58a90464c2",
        strip_prefix = "shellcheck-v0.10.0",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/f53c34493657dfea83b657e4b62cc68c25fbc383dff64c8d581613b037aacaa3",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.1/actionlint_1.7.1_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f53c34493657dfea83b657e4b62cc68c25fbc383dff64c8d581613b037aacaa3",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/21a20f38b19dc962d89e17fe1c6f116199e9e0d343ab33361868def14cc220fc",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.1/actionlint_1.7.1_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "21a20f38b19dc962d89e17fe1c6f116199e9e0d343ab33361868def14cc220fc",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee24184e2e7003c19eb739717b34b7c65d096f2ca0df8d571837b4f20112d573",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.1/actionlint_1.7.1_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "ee24184e2e7003c19eb739717b34b7c65d096f2ca0df8d571837b4f20112d573",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a72f66f28a4cc294670abb7a5e3392033700e00cc6a385c32fb769971b71ec9f",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.1/actionlint_1.7.1_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a72f66f28a4cc294670abb7a5e3392033700e00cc6a385c32fb769971b71ec9f",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/8cbd8d64cbd1f25b38f33fa04db602466dade79e99c99dc9da053b5962d34014",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.6/tfsec_1.28.6_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "8cbd8d64cbd1f25b38f33fa04db602466dade79e99c99dc9da053b5962d34014",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4bc7b0f0592be4fa384cff52af5b1cdd2066ba7a06001bea98690340851c0bce",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.6/tfsec_1.28.6_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "4bc7b0f0592be4fa384cff52af5b1cdd2066ba7a06001bea98690340851c0bce",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b31e954819faa7d6151b999548cefb782f2f4dc64b355c8747e44d4b0b2faca",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.6/tfsec_1.28.6_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "3b31e954819faa7d6151b999548cefb782f2f4dc64b355c8747e44d4b0b2faca",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aa132b7e0e69e16f1c9320257841751e52c42d9791b7f900de72cf0b06ffe74c",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.6/tfsec_1.28.6_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "aa132b7e0e69e16f1c9320257841751e52c42d9791b7f900de72cf0b06ffe74c",
    )

def _golangci_lint_deps():
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2e7c1b2af70ecb7ce18da79a3951db7046dcb709566c018fb93c61e8733b2239",
            "https://github.com/golangci/golangci-lint/releases/download/v1.58.1/golangci-lint-1.58.1-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.58.1-linux-amd64",
        type = "tar.gz",
        sha256 = "2e7c1b2af70ecb7ce18da79a3951db7046dcb709566c018fb93c61e8733b2239",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e78f55156ef7f8bd4a746aa695c6bd299dc0aa71b91ec066680d2c59e443708b",
            "https://github.com/golangci/golangci-lint/releases/download/v1.58.1/golangci-lint-1.58.1-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.58.1-linux-arm64",
        type = "tar.gz",
        sha256 = "e78f55156ef7f8bd4a746aa695c6bd299dc0aa71b91ec066680d2c59e443708b",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79f587134955808e245aa78a734bc28f6e03fc42efb3e91d2a705930909fe0c0",
            "https://github.com/golangci/golangci-lint/releases/download/v1.58.1/golangci-lint-1.58.1-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.58.1-darwin-amd64",
        type = "tar.gz",
        sha256 = "79f587134955808e245aa78a734bc28f6e03fc42efb3e91d2a705930909fe0c0",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bd59615a2fdf8e8ff544ba79e15c53f81a538bd68cccf52f4802b65e0216199d",
            "https://github.com/golangci/golangci-lint/releases/download/v1.58.1/golangci-lint-1.58.1-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.58.1-darwin-arm64",
        type = "tar.gz",
        sha256 = "bd59615a2fdf8e8ff544ba79e15c53f81a538bd68cccf52f4802b65e0216199d",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8a94dce37ce72c89c82e6c4baf77797a2a4a2eef3b02a7f39b40ef7fb0f39f94",
            "https://github.com/bufbuild/buf/releases/download/v1.31.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "8a94dce37ce72c89c82e6c4baf77797a2a4a2eef3b02a7f39b40ef7fb0f39f94",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ed435fe40a8b719354f746261cf9dbfcbfa4165fdb600e2f324ad8f6fe488dd2",
            "https://github.com/bufbuild/buf/releases/download/v1.31.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "ed435fe40a8b719354f746261cf9dbfcbfa4165fdb600e2f324ad8f6fe488dd2",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a88b4ccf6aee4f7d525917bf4636253faa7c13b8f45c4c732a7fea55441ef835",
            "https://github.com/bufbuild/buf/releases/download/v1.31.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "a88b4ccf6aee4f7d525917bf4636253faa7c13b8f45c4c732a7fea55441ef835",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7519df87d3f681d5348f00e96215543edc55c62d821527056b5d8201d8982f28",
            "https://github.com/bufbuild/buf/releases/download/v1.31.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7519df87d3f681d5348f00e96215543edc55c62d821527056b5d8201d8982f28",
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
        sha256 = "a5844ef2c38ef6ddf3b5a8f7d91e7e0e8ebc39a38bb3fc8013d629c1ef29c259",
        strip_prefix = "linux-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a5844ef2c38ef6ddf3b5a8f7d91e7e0e8ebc39a38bb3fc8013d629c1ef29c259",
            "https://get.helm.sh/helm-v3.14.4-linux-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_linux_arm64",
        sha256 = "113ccc53b7c57c2aba0cd0aa560b5500841b18b5210d78641acfddc53dac8ab2",
        strip_prefix = "linux-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/113ccc53b7c57c2aba0cd0aa560b5500841b18b5210d78641acfddc53dac8ab2",
            "https://get.helm.sh/helm-v3.14.4-linux-arm64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_amd64",
        sha256 = "73434aeac36ad068ce2e5582b8851a286dc628eae16494a26e2ad0b24a7199f9",
        strip_prefix = "darwin-amd64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/73434aeac36ad068ce2e5582b8851a286dc628eae16494a26e2ad0b24a7199f9",
            "https://get.helm.sh/helm-v3.14.4-darwin-amd64.tar.gz",
        ],
        type = "tar.gz",
    )
    http_archive(
        name = "com_github_helm_helm_darwin_arm64",
        sha256 = "61e9c5455f06b2ad0a1280975bf65892e707adc19d766b0cf4e9006e3b7b4b6c",
        strip_prefix = "darwin-arm64",
        build_file_content = """exports_files(["helm"], visibility = ["//visibility:public"])""",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/61e9c5455f06b2ad0a1280975bf65892e707adc19d766b0cf4e9006e3b7b4b6c",
            "https://get.helm.sh/helm-v3.14.4-darwin-arm64.tar.gz",
        ],
    )

def _ghh_deps():
    http_archive(
        name = "com_github_katexochen_ghh_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32f8a4110d88d80e163212a89a3538a13326494840ac97183d1b20bcc9eac7ba",
            "https://github.com/katexochen/ghh/releases/download/v0.3.5/ghh_0.3.5_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "32f8a4110d88d80e163212a89a3538a13326494840ac97183d1b20bcc9eac7ba",
    )
    http_archive(
        name = "com_github_katexochen_ghh_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b43ef1dd2f851eed7c69c87f4f73dd923bd1170cefbde247933d5b398a3319d1",
            "https://github.com/katexochen/ghh/releases/download/v0.3.5/ghh_0.3.5_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "b43ef1dd2f851eed7c69c87f4f73dd923bd1170cefbde247933d5b398a3319d1",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7db9ebb62faf2a31f56e7a8994a971adddec98b3238880ae58b01eb549b8bba3",
            "https://github.com/katexochen/ghh/releases/download/v0.3.5/ghh_0.3.5_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "7db9ebb62faf2a31f56e7a8994a971adddec98b3238880ae58b01eb549b8bba3",
    )
    http_archive(
        name = "com_github_katexochen_ghh_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/78a2c8b321893736a2b5de59898794a72b878db9329157f348489c73d4592c6f",
            "https://github.com/katexochen/ghh/releases/download/v0.3.5/ghh_0.3.5_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        build_file_content = """exports_files(["ghh"], visibility = ["//visibility:public"])""",
        sha256 = "78a2c8b321893736a2b5de59898794a72b878db9329157f348489c73d4592c6f",
    )
