"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7b9f496d5ef3bb6943013d1c11f7a8108aa4d6a050f5fcbaee453285d1a473d4",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-6.6.40-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "7b9f496d5ef3bb6943013d1c11f7a8108aa4d6a050f5fcbaee453285d1a473d4",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d67df9f6481da84d858bac6836f373ebbe6ca6e1a1b1bed6e402916558b4b6a",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-core-6.6.40-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "4d67df9f6481da84d858bac6836f373ebbe6ca6e1a1b1bed6e402916558b4b6a",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ce8b7d81373818af4feeaa1aad4f702011f042094125a1dc12103c160a1e574",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-modules-6.6.40-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "4ce8b7d81373818af4feeaa1aad4f702011f042094125a1dc12103c160a1e574",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f27b2c6510f38b940dfd4138147f557fd87238d76c3fe5c5f061e3731d8dae8",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-modules-core-6.6.40-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "9f27b2c6510f38b940dfd4138147f557fd87238d76c3fe5c5f061e3731d8dae8",
    )

    # mainline kernel
    http_file(
        name = "kernel_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6eaec29870e6549d95a93b72ea10715507db84b851c68c0d75e44e4c20f895f2",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "6eaec29870e6549d95a93b72ea10715507db84b851c68c0d75e44e4c20f895f2",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/910fd35209f7dc8185e88dddeaccf6158dd63ad9fd469ef3dc81b96840ef28eb",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "910fd35209f7dc8185e88dddeaccf6158dd63ad9fd469ef3dc81b96840ef28eb",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8de20433c68d2fe0ca6625e25f314aba36a9327592db8b1478b97bb50521149",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "b8de20433c68d2fe0ca6625e25f314aba36a9327592db8b1478b97bb50521149",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8ecd8e96483810d18e04a20cd8ecef46f27bff0fbb54f23e67adb813828b3cec",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "8ecd8e96483810d18e04a20cd8ecef46f27bff0fbb54f23e67adb813828b3cec",
    )
