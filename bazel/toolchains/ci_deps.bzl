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
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c3be12b05d5c177a04c29e3c78ce89ac86f1595681cab149b65b97c4e227198",
            "https://github.com/koalaman/shellcheck/releases/download/v0.11.0/shellcheck-v0.11.0.linux.x86_64.tar.xz",
        ],
        strip_prefix = "shellcheck-v0.11.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
        sha256 = "8c3be12b05d5c177a04c29e3c78ce89ac86f1595681cab149b65b97c4e227198",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/12b331c1d2db6b9eb13cfca64306b1b157a86eb69db83023e261eaa7e7c14588",
            "https://github.com/koalaman/shellcheck/releases/download/v0.11.0/shellcheck-v0.11.0.linux.aarch64.tar.xz",
        ],
        strip_prefix = "shellcheck-v0.11.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
        sha256 = "12b331c1d2db6b9eb13cfca64306b1b157a86eb69db83023e261eaa7e7c14588",
    )
    http_archive(
        name = "com_github_koalaman_shellcheck_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3c89db4edcab7cf1c27bff178882e0f6f27f7afdf54e859fa041fca10febe4c6",
            "https://github.com/koalaman/shellcheck/releases/download/v0.11.0/shellcheck-v0.11.0.darwin.x86_64.tar.xz",
        ],
        strip_prefix = "shellcheck-v0.11.0",
        build_file_content = """exports_files(["shellcheck"], visibility = ["//visibility:public"])""",
        type = "tar.xz",
        sha256 = "3c89db4edcab7cf1c27bff178882e0f6f27f7afdf54e859fa041fca10febe4c6",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/be92c2652ab7b6d08425428797ceabeb16e31a781c07bc388456b4e592f3e36a",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.8/actionlint_1.7.8_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "be92c2652ab7b6d08425428797ceabeb16e31a781c07bc388456b4e592f3e36a",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4c65dbb2d59b409cdd75d47ffa8fa32af8f0eee573ac510468dc2275c48bf07c",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.8/actionlint_1.7.8_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "4c65dbb2d59b409cdd75d47ffa8fa32af8f0eee573ac510468dc2275c48bf07c",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/16b85caf792b34bcc40f7437736c4347680da0a1b034353a85012debbd71a461",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.8/actionlint_1.7.8_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "16b85caf792b34bcc40f7437736c4347680da0a1b034353a85012debbd71a461",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ffb1f6c429a51dc9f37af9d11f96c16bd52f54b713bf7f8bd92f7fce9fd4284a",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.8/actionlint_1.7.8_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "ffb1f6c429a51dc9f37af9d11f96c16bd52f54b713bf7f8bd92f7fce9fd4284a",
    )

def _gofumpt_deps():
    http_file(
        name = "com_github_mvdan_gofumpt_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/72cf61b12fef91eab6df6db4a4284f30616b5ead330112e28a1fa1cb15e57339",
            "https://github.com/mvdan/gofumpt/releases/download/v0.9.2/gofumpt_v0.9.2_linux_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "72cf61b12fef91eab6df6db4a4284f30616b5ead330112e28a1fa1cb15e57339",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5acaa5a554050f55fc81ef02a8b0d14ab6b3c058a84513885286dc52d3451645",
            "https://github.com/mvdan/gofumpt/releases/download/v0.9.2/gofumpt_v0.9.2_linux_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "5acaa5a554050f55fc81ef02a8b0d14ab6b3c058a84513885286dc52d3451645",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4172b912ec514038605f334fef9ed7b3f12ca3e40024cb0a622eab3073a55e57",
            "https://github.com/mvdan/gofumpt/releases/download/v0.9.2/gofumpt_v0.9.2_darwin_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "4172b912ec514038605f334fef9ed7b3f12ca3e40024cb0a622eab3073a55e57",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c241fb742599a6cb0563d7377f59def65d451b23dd718dbc6ddf4ab5e695e8f1",
            "https://github.com/mvdan/gofumpt/releases/download/v0.9.2/gofumpt_v0.9.2_darwin_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "c241fb742599a6cb0563d7377f59def65d451b23dd718dbc6ddf4ab5e695e8f1",
    )

def _tfsec_deps():
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/329ae7f67f2f1813ebe08de498719ea7003c75d3ca24bb0b038369062508008e",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.14/tfsec_1.28.14_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "329ae7f67f2f1813ebe08de498719ea7003c75d3ca24bb0b038369062508008e",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/13dcbd3602027be49ce6cab7e1c24b0a8e833f0143fe327b0a13b87686541ce0",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.14/tfsec_1.28.14_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "13dcbd3602027be49ce6cab7e1c24b0a8e833f0143fe327b0a13b87686541ce0",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0aeef31f83d6f44ba9ba5b6cbb954304c772dee73ac704e38896940f94af887a",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.14/tfsec_1.28.14_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "0aeef31f83d6f44ba9ba5b6cbb954304c772dee73ac704e38896940f94af887a",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f39d59a3f9be4eeb3d965657653ad62243103a3d921ce52ca8f907cff45896f5",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.14/tfsec_1.28.14_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f39d59a3f9be4eeb3d965657653ad62243103a3d921ce52ca8f907cff45896f5",
    )

def _golangci_lint_deps():
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e55e0eb515936c0fbd178bce504798a9bd2f0b191e5e357768b18fd5415ee541",
            "https://github.com/golangci/golangci-lint/releases/download/v2.1.6/golangci-lint-2.1.6-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-2.1.6-linux-amd64",
        type = "tar.gz",
        sha256 = "e55e0eb515936c0fbd178bce504798a9bd2f0b191e5e357768b18fd5415ee541",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/582eb73880f4408d7fb89f12b502d577bd7b0b63d8c681da92bb6b9d934d4363",
            "https://github.com/golangci/golangci-lint/releases/download/v2.1.6/golangci-lint-2.1.6-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-2.1.6-linux-arm64",
        type = "tar.gz",
        sha256 = "582eb73880f4408d7fb89f12b502d577bd7b0b63d8c681da92bb6b9d934d4363",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e091107c4ca7e283902343ba3a09d14fb56b86e071effd461ce9d67193ef580e",
            "https://github.com/golangci/golangci-lint/releases/download/v2.1.6/golangci-lint-2.1.6-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-2.1.6-darwin-amd64",
        type = "tar.gz",
        sha256 = "e091107c4ca7e283902343ba3a09d14fb56b86e071effd461ce9d67193ef580e",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/90783fa092a0f64a4f7b7d419f3da1f53207e300261773babe962957240e9ea6",
            "https://github.com/golangci/golangci-lint/releases/download/v2.1.6/golangci-lint-2.1.6-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-2.1.6-darwin-arm64",
        type = "tar.gz",
        sha256 = "90783fa092a0f64a4f7b7d419f3da1f53207e300261773babe962957240e9ea6",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fa10faf16973f3861992cc2687b651350d70eafd467aea72cf0994556c2a0927",
            "https://github.com/bufbuild/buf/releases/download/v1.54.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "fa10faf16973f3861992cc2687b651350d70eafd467aea72cf0994556c2a0927",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f41ef4431858556ece6a77662d6b9317fa4406585998cb3dffb7403b3e86713e",
            "https://github.com/bufbuild/buf/releases/download/v1.54.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f41ef4431858556ece6a77662d6b9317fa4406585998cb3dffb7403b3e86713e",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/22c9836a836b867e49e9d0ef223fd934cbf2690e7400facddb9be07b8809f889",
            "https://github.com/bufbuild/buf/releases/download/v1.54.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "22c9836a836b867e49e9d0ef223fd934cbf2690e7400facddb9be07b8809f889",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f01f32a690efab3ef22a1c821aebc0c4bec7ca63faddbf64408d7d614e9d7f92",
            "https://github.com/bufbuild/buf/releases/download/v1.54.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "f01f32a690efab3ef22a1c821aebc0c4bec7ca63faddbf64408d7d614e9d7f92",
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
