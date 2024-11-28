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
            "https://cdn.confidential.cloud/constellation/cas/sha256/fc0a6886bbb9a23a39eeec4b176193cadb54ddbe77cdbb19b637933919545395",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.4/actionlint_1.7.4_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "fc0a6886bbb9a23a39eeec4b176193cadb54ddbe77cdbb19b637933919545395",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_linux_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ede03682dc955381d057dde95bb85ce9ca418122209a8a313b617d4adec56416",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.4/actionlint_1.7.4_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "ede03682dc955381d057dde95bb85ce9ca418122209a8a313b617d4adec56416",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_amd64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/63a3ba90ee2325afad3ff2e64a4d80688c261e6c68be8e6ab91214637bf936b8",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.4/actionlint_1.7.4_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "63a3ba90ee2325afad3ff2e64a4d80688c261e6c68be8e6ab91214637bf936b8",
    )
    http_archive(
        name = "com_github_rhysd_actionlint_darwin_arm64",
        build_file_content = """exports_files(["actionlint"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cbd193bb490f598d77e179261d7b76dfebd049dddede5803ba21cbf6a469aeee",
            "https://github.com/rhysd/actionlint/releases/download/v1.7.4/actionlint_1.7.4_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "cbd193bb490f598d77e179261d7b76dfebd049dddede5803ba21cbf6a469aeee",
    )

def _gofumpt_deps():
    http_file(
        name = "com_github_mvdan_gofumpt_linux_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6ff459c1dcae3b0b00844c1a5a4a5b0f547237d8a4f3624aaea8d424aeef24c6",
            "https://github.com/mvdan/gofumpt/releases/download/v0.7.0/gofumpt_v0.7.0_linux_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "6ff459c1dcae3b0b00844c1a5a4a5b0f547237d8a4f3624aaea8d424aeef24c6",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_linux_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/00c18c88ef50437629626ba20d677f4648684cb280952814cdd887677d42cbd3",
            "https://github.com/mvdan/gofumpt/releases/download/v0.7.0/gofumpt_v0.7.0_linux_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "00c18c88ef50437629626ba20d677f4648684cb280952814cdd887677d42cbd3",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_amd64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7d05e092da45c5ec96344ab635b1d6547c3e27c840ba39bc76989934efd7ce3",
            "https://github.com/mvdan/gofumpt/releases/download/v0.7.0/gofumpt_v0.7.0_darwin_amd64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "b7d05e092da45c5ec96344ab635b1d6547c3e27c840ba39bc76989934efd7ce3",
    )
    http_file(
        name = "com_github_mvdan_gofumpt_darwin_arm64",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/08f23114760a090b090706d92b8c52b9875b9eb352d76c77aa354d6aa20b045a",
            "https://github.com/mvdan/gofumpt/releases/download/v0.7.0/gofumpt_v0.7.0_darwin_arm64",
        ],
        executable = True,
        downloaded_file_path = "gofumpt",
        sha256 = "08f23114760a090b090706d92b8c52b9875b9eb352d76c77aa354d6aa20b045a",
    )

def _tfsec_deps():
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9d783fa225a570f034000136973afba86a1708c919a539b72b3ea954a198289c",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.11/tfsec_1.28.11_linux_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "9d783fa225a570f034000136973afba86a1708c919a539b72b3ea954a198289c",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_linux_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/68b5c4f6b7c459dd890ecff94b0732e456ef45974894f58bbb90fbb4816f3e52",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.11/tfsec_1.28.11_linux_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "68b5c4f6b7c459dd890ecff94b0732e456ef45974894f58bbb90fbb4816f3e52",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_amd64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d377597f2fd4e6956bb7beb711d627b0e0204c421c17e2cd062213222c2f3001",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.11/tfsec_1.28.11_darwin_amd64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "d377597f2fd4e6956bb7beb711d627b0e0204c421c17e2cd062213222c2f3001",
    )
    http_archive(
        name = "com_github_aquasecurity_tfsec_darwin_arm64",
        build_file_content = """exports_files(["tfsec"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/14db6b40049226ebc779789196f99eb4977bb93c99fa51c8b72b603e6cdf44e7",
            "https://github.com/aquasecurity/tfsec/releases/download/v1.28.11/tfsec_1.28.11_darwin_arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "14db6b40049226ebc779789196f99eb4977bb93c99fa51c8b72b603e6cdf44e7",
    )

def _golangci_lint_deps():
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/77cb0af99379d9a21d5dc8c38364d060e864a01bd2f3e30b5e8cc550c3a54111",
            "https://github.com/golangci/golangci-lint/releases/download/v1.61.0/golangci-lint-1.61.0-linux-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.61.0-linux-amd64",
        type = "tar.gz",
        sha256 = "77cb0af99379d9a21d5dc8c38364d060e864a01bd2f3e30b5e8cc550c3a54111",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_linux_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af60ac05566d9351615cb31b4cc070185c25bf8cbd9b09c1873aa5ec6f3cc17e",
            "https://github.com/golangci/golangci-lint/releases/download/v1.61.0/golangci-lint-1.61.0-linux-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.61.0-linux-arm64",
        type = "tar.gz",
        sha256 = "af60ac05566d9351615cb31b4cc070185c25bf8cbd9b09c1873aa5ec6f3cc17e",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_amd64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5c280ef3284f80c54fd90d73dc39ca276953949da1db03eb9dd0fbf868cc6e55",
            "https://github.com/golangci/golangci-lint/releases/download/v1.61.0/golangci-lint-1.61.0-darwin-amd64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.61.0-darwin-amd64",
        type = "tar.gz",
        sha256 = "5c280ef3284f80c54fd90d73dc39ca276953949da1db03eb9dd0fbf868cc6e55",
    )
    http_archive(
        name = "com_github_golangci_golangci_lint_darwin_arm64",
        build_file = "//bazel/toolchains:BUILD.golangci.bazel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/544334890701e4e04a6e574bc010bea8945205c08c44cced73745a6378012d36",
            "https://github.com/golangci/golangci-lint/releases/download/v1.61.0/golangci-lint-1.61.0-darwin-arm64.tar.gz",
        ],
        strip_prefix = "golangci-lint-1.61.0-darwin-arm64",
        type = "tar.gz",
        sha256 = "544334890701e4e04a6e574bc010bea8945205c08c44cced73745a6378012d36",
    )

def _buf_deps():
    http_archive(
        name = "com_github_bufbuild_buf_linux_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/deebd48a6bf85b073d7c7800c17b330376487e86852d4905c76a205b6fd795d4",
            "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Linux-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "deebd48a6bf85b073d7c7800c17b330376487e86852d4905c76a205b6fd795d4",
    )
    http_archive(
        name = "com_github_bufbuild_buf_linux_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2d3ebfed036881d0615e5b24288cf788791b45848f26e915e3efe7ee9c10735d",
            "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Linux-aarch64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "2d3ebfed036881d0615e5b24288cf788791b45848f26e915e3efe7ee9c10735d",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_amd64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7fef3c482ac440cc09c40864498ef1f44745fde82428ddf52edd2012d3a036a4",
            "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Darwin-x86_64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "7fef3c482ac440cc09c40864498ef1f44745fde82428ddf52edd2012d3a036a4",
    )
    http_archive(
        name = "com_github_bufbuild_buf_darwin_arm64",
        strip_prefix = "buf/bin",
        build_file_content = """exports_files(["buf"], visibility = ["//visibility:public"])""",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e5309c70c7bb4a06d799ab7c7601c0d647c704085593d5cd981db29f986e469b",
            "https://github.com/bufbuild/buf/releases/download/v1.45.0/buf-Darwin-arm64.tar.gz",
        ],
        type = "tar.gz",
        sha256 = "e5309c70c7bb4a06d799ab7c7601c0d647c704085593d5cd981db29f986e469b",
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
