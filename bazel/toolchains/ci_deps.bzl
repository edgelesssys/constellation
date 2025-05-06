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
            "https://cdn.confidential.cloud/constellation/cas/sha256/023070a287cd8cccd71515fedc843f1985bf96c436b7effaecce67290e7e0757",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.7/actionlint_1.7.7_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "023070a287cd8cccd71515fedc843f1985bf96c436b7effaecce67290e7e0757",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/401942f9c24ed71e4fe71b76c7d638f66d8633575c4016efd2977ce7c28317d0",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.7/actionlint_1.7.7_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "401942f9c24ed71e4fe71b76c7d638f66d8633575c4016efd2977ce7c28317d0",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/28e5de5a05fc558474f638323d736d822fff183d2d492f0aecb2b73cc44584f5",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.7/actionlint_1.7.7_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "28e5de5a05fc558474f638323d736d822fff183d2d492f0aecb2b73cc44584f5",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2693315b9093aeacb4ebd91a993fea54fc215057bf0da2659056b4bc033873db",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.7/actionlint_1.7.7_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "2693315b9093aeacb4ebd91a993fea54fc215057bf0da2659056b4bc033873db",
    )

def _gofumpt_deps():
    http_file(
        name = "com_github_mvdan_gofumpt_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/11604bbaf7321abcc2fca2c6a37b7e9198bb1e76e5a86f297c07201e8ab1fda9",
            "https://github.com/mvdan/gofumpt/releases/download/v0.8.0/gofumpt_v0.8.0_linux_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "11604bbaf7321abcc2fca2c6a37b7e9198bb1e76e5a86f297c07201e8ab1fda9",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/787c1d3d4d20e6fe2b0bf06a5a913ac0f50343dbf9a71540724a2b8092a0e6ca",
            "https://github.com/mvdan/gofumpt/releases/download/v0.8.0/gofumpt_v0.8.0_linux_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "787c1d3d4d20e6fe2b0bf06a5a913ac0f50343dbf9a71540724a2b8092a0e6ca",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0dda6600cf263b703a5ad93e792b06180c36afdee9638617a91dd552f2c6fb3e",
            "https://github.com/mvdan/gofumpt/releases/download/v0.8.0/gofumpt_v0.8.0_darwin_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "0dda6600cf263b703a5ad93e792b06180c36afdee9638617a91dd552f2c6fb3e",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e66e92b7a67d1d12839ab030fb7ae38e5e2273474af3762e67bc7fe9471fcd9",
            "https://github.com/mvdan/gofumpt/releases/download/v0.8.0/gofumpt_v0.8.0_darwin_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "7e66e92b7a67d1d12839ab030fb7ae38e5e2273474af3762e67bc7fe9471fcd9",
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
            "https://cdn.confidential.cloud/constellation/cas/sha256/2dc9793f94806f3e4b0158a36e33680e2b962be50c12a7785ccb3da00d2fb2e7",
            "https://github.com/bufbuild/buf/releases/download/v1.53.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "2dc9793f94806f3e4b0158a36e33680e2b962be50c12a7785ccb3da00d2fb2e7",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4b5a40388c6a6608c29c7961fefc97ea10585e684038b1e6f9716f2216eaf9a9",
            "https://github.com/bufbuild/buf/releases/download/v1.53.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "4b5a40388c6a6608c29c7961fefc97ea10585e684038b1e6f9716f2216eaf9a9",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/44f1c88c597694b607902adc7da5f0742511f3c379cc08058b140bcb5e8a4b11",
            "https://github.com/bufbuild/buf/releases/download/v1.53.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "44f1c88c597694b607902adc7da5f0742511f3c379cc08058b140bcb5e8a4b11",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3a12862c5e27fe57e27b3c4dec6aae3e923ec626b1b443eb2310a3487066735f",
            "https://github.com/bufbuild/buf/releases/download/v1.53.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "3a12862c5e27fe57e27b3c4dec6aae3e923ec626b1b443eb2310a3487066735f",
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
