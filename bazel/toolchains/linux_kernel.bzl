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
            "https://cdn.confidential.cloud/constellation/cas/sha256/199d3aa46ed1325bb37e163a7f649a4f4dc739421389b7e9d9697af25c589a92",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/100.fc38/x86_64/kernel-6.8.9-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "199d3aa46ed1325bb37e163a7f649a4f4dc739421389b7e9d9697af25c589a92",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aa58f8a6850e67b3e8e93240e0d62a56cee55ddb7a14df6346e9c26ce6fb2bce",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/100.fc38/x86_64/kernel-core-6.8.9-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "aa58f8a6850e67b3e8e93240e0d62a56cee55ddb7a14df6346e9c26ce6fb2bce",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/46ca357ee96f17a37664563b308ee9bee7b21b9b639f3389774e953a9b34ce81",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/100.fc38/x86_64/kernel-modules-6.8.9-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "46ca357ee96f17a37664563b308ee9bee7b21b9b639f3389774e953a9b34ce81",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee983f7784fc5801378d675a4d41a57cafc6525147b601c49a04cea60ac81d79",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/100.fc38/x86_64/kernel-modules-core-6.8.9-100.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "ee983f7784fc5801378d675a4d41a57cafc6525147b601c49a04cea60ac81d79",
    )
