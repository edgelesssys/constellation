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
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "6eaec29870e6549d95a93b72ea10715507db84b851c68c0d75e44e4c20f895f2",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "910fd35209f7dc8185e88dddeaccf6158dd63ad9fd469ef3dc81b96840ef28eb",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "b8de20433c68d2fe0ca6625e25f314aba36a9327592db8b1478b97bb50521149",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "8ecd8e96483810d18e04a20cd8ecef46f27bff0fbb54f23e67adb813828b3cec",
    )
