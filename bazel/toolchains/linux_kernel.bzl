"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def nvidia_kos():
    """nvidia ko"""

    # nvidia kernel module
    http_file(
        name = "nvidia_ko",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/nvidia.ko",
        ],
        sha256 = "ee2555a032cf17f2756312ae2004673acc461c3d6fbc4b3021b2d4735034fb11",
        downloaded_file_path = "nvidia.ko",
    )
    http_file(
        name = "nvidia_drm_ko",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/nvidia-drm.ko",
        ],
        sha256 = "78e08dce97ba7306bbd5658183dbcb1221bdf9cb8fe2fc5528797b0fa0e9e31d",
        downloaded_file_path = "nvidia-drm.ko",
    )
    http_file(
        name = "nvidia_modeset_ko",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/nvidia-modeset.ko",
        ],
        sha256 = "18d669cc4c089f896457560b69cc6eb30344a434de5a35ab5846ac65b88dde5e",
        downloaded_file_path = "nvidia-modeset.ko",
    )
    http_file(
        name = "nvidia_peermem_ko",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/nvidia-peermem.ko",
        ],
        sha256 = "52ce0116713a35a4db6b36a6d029a3d1a4ae1d30c032c8c71281545e878e5923",
        downloaded_file_path = "nvidia-peermem.ko",
    )
    http_file(
        name = "nvidia_uvm_ko",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/nvidia-uvm.ko",
        ],
        sha256 = "c0c4e044f2bbaa939d9e1bfb6e11ce1b441aeaac683defc2aa33bfb7b2b6c217",
        downloaded_file_path = "nvidia-uvm.ko",
    )
    http_file(
        name = "gsp_ga10x",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/gsp_ga10x.bin",
        ],
        sha256 = "1f6d303a192388b3ccd97f468fa4ed64b5921a8b76ae1d4660f2caed5568dc17",
        downloaded_file_path = "gsp_ga10x.bin",
    )
    http_file(
        name = "gsp_tu10x",
        urls = [
            "https://cdn.confidential.cloud/constellation/kernel/nvidia/6.2.0/535.129.03/gsp_tu10x.bin",
        ],
        sha256 = "6ada90fdfbfa134ab02c588be09441a9c64670a79d0cbb200106d0f3d3f672fe",
        downloaded_file_path = "gsp_tu10x.bin",
    )

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
