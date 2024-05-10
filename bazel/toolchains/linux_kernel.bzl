"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b1f40fcc493189045ae93124968d185c8b3351433b9e0f49cfdb49998f5efa17",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.90-100.constellation/kernel-6.1.90-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "b1f40fcc493189045ae93124968d185c8b3351433b9e0f49cfdb49998f5efa17",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/124f418f63ee7453e220448d372d5fd9727d42e5236da22d98e33fabc4fcd5dc",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.90-100.constellation/kernel-core-6.1.90-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "124f418f63ee7453e220448d372d5fd9727d42e5236da22d98e33fabc4fcd5dc",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8aef6f1e07ddfc57b0a0e2c9a23eae392b3c628770492c21a762cbab941c6923",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.90-100.constellation/kernel-modules-6.1.90-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "8aef6f1e07ddfc57b0a0e2c9a23eae392b3c628770492c21a762cbab941c6923",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2a8865452cd9680da17834bc3c1b323779f24bb70d01be9cd5df2b02bde815cb",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.90-100.constellation/kernel-modules-core-6.1.90-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "2a8865452cd9680da17834bc3c1b323779f24bb70d01be9cd5df2b02bde815cb",
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
