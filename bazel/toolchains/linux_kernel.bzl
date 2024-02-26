"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b4a8e5217cb62241631d1a7357979face1ad455a08cd4ca8f59c2252a90047f6",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.79-100.constellation/kernel-6.1.79-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "b4a8e5217cb62241631d1a7357979face1ad455a08cd4ca8f59c2252a90047f6",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/03e9ae8508cf1cd964216eed858c69d55629d8d27a44965857e408defcfe4785",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.79-100.constellation/kernel-core-6.1.79-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "03e9ae8508cf1cd964216eed858c69d55629d8d27a44965857e408defcfe4785",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c306c13f17024915682304e4f05ca21dd9533a34921684520135bc7e69e3d327",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.79-100.constellation/kernel-modules-6.1.79-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "c306c13f17024915682304e4f05ca21dd9533a34921684520135bc7e69e3d327",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2183a23562e1d69a64e799baf1cf64ae34c4058f24993e3fe7645e3e363e899a",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.79-100.constellation/kernel-modules-core-6.1.79-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "2183a23562e1d69a64e799baf1cf64ae34c4058f24993e3fe7645e3e363e899a",
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
