"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c87995e19c04e2f033e6db5e92bfcb845ac015722e776c09a7af4c82c86cd273",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-6.6.30-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "c87995e19c04e2f033e6db5e92bfcb845ac015722e776c09a7af4c82c86cd273",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5692d862b0cc0c442c581e5f3dc9f3c36cabda0c29d3f62e9b6313a6ec88b140",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-core-6.6.30-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "5692d862b0cc0c442c581e5f3dc9f3c36cabda0c29d3f62e9b6313a6ec88b140",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e1b697343b4f8ed8e992cd92860208dc1c28eb8b25a88f42f426326a0bbc307f",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-modules-6.6.30-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "e1b697343b4f8ed8e992cd92860208dc1c28eb8b25a88f42f426326a0bbc307f",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/448c6b10d9ed02aed078ff77223f5e495b2041be12d92eb0e5ca5726a08e0626",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.30-100.constellation/kernel-modules-core-6.6.30-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "448c6b10d9ed02aed078ff77223f5e495b2041be12d92eb0e5ca5726a08e0626",
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
