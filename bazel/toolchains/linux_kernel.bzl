"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/6.2.0-100.constellation/kernel-6.2.0-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "2fa13500cb4f6cf0b921941876579f65fbcefc897eb6ef0ae509ce6053e634ec",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/6.2.0-100.constellation/kernel-core-6.2.0-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "bc48e9927df6e9f0f7b174fe916765e0a39ae817207a5fc84c53b7b35a3a2e42",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/6.2.0-100.constellation/kernel-modules-6.2.0-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "c957c404b16cd09b70badd4e5bd9252e2bf5bde2d41b4e329b1b635a6b4e97e0",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/6.2.0-100.constellation/kernel-modules-core-6.2.0-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "c8819019b1bfb75a888b4339bb5f3e0e5d8939fe6685419c2253f81ddb2f5a88",
    )

    # mainline kernel
    http_file(
        name = "kernel_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/303506771939b324c32c2d7df4ce2a8ca08af4fe0fec77712084bdd3c1481bc9",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.7.6/100.fc38/x86_64/kernel-6.7.6-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "303506771939b324c32c2d7df4ce2a8ca08af4fe0fec77712084bdd3c1481bc9",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f51347ddeca348494fe20a898b455f84e1e7c4cda6832fb5dc2d092b94ddc039",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.7.6/100.fc38/x86_64/kernel-core-6.7.6-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "f51347ddeca348494fe20a898b455f84e1e7c4cda6832fb5dc2d092b94ddc039",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a0c65aac2628fc24e460f68eb2b1a9b8d749f319d10962257dcfeee7cadb09c",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.7.6/100.fc38/x86_64/kernel-modules-6.7.6-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "4a0c65aac2628fc24e460f68eb2b1a9b8d749f319d10962257dcfeee7cadb09c",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/127a1b895ba6a7842e8503770ccc3b412fca195a9f750bb3f94788c2384ab577",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.7.6/100.fc38/x86_64/kernel-modules-core-6.7.6-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "127a1b895ba6a7842e8503770ccc3b412fca195a9f750bb3f94788c2384ab577",
    )
