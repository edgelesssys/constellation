"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1923a89d4ab11f61de222a687e06a2a4f522dc6df47713a875d886a03aeb98d5",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.74-100.constellation/kernel-6.1.74-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "1923a89d4ab11f61de222a687e06a2a4f522dc6df47713a875d886a03aeb98d5",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c98a316f1f2eb5345169b9ad46b2c22f729e5d583c39202dbd4cf1f74e74e2a6",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.74-100.constellation/kernel-core-6.1.74-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "c98a316f1f2eb5345169b9ad46b2c22f729e5d583c39202dbd4cf1f74e74e2a6",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ffa10f0bf2e35d45e2dc3bbf77ee21c057449feef3cf5db8855c9890f99bf442",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.74-100.constellation/kernel-modules-6.1.74-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "ffa10f0bf2e35d45e2dc3bbf77ee21c057449feef3cf5db8855c9890f99bf442",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee02c7836f3400b9483bfc934e89195bb059bb6423c807a26215a7e08548b398",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.74-100.constellation/kernel-modules-core-6.1.74-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "ee02c7836f3400b9483bfc934e89195bb059bb6423c807a26215a7e08548b398",
    )

    # mainline kernel
    http_file(
        name = "kernel_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b42a4ee6c486832adbff101447a0f92b61905e43acabffc40e573ebf87799889",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "b42a4ee6c486832adbff101447a0f92b61905e43acabffc40e573ebf87799889",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/88f34a4add6d1c8d9c7cff499843d0d565aa798b1bf365c7b4a0e0c48adab2b4",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "88f34a4add6d1c8d9c7cff499843d0d565aa798b1bf365c7b4a0e0c48adab2b4",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ba6599de2934315fb659b512659e5d96b2812f877e0c2d41625d899d8d440ad",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "4ba6599de2934315fb659b512659e5d96b2812f877e0c2d41625d899d8d440ad",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3d01a6e11fb4110b6c7f2f63c113c7b7c7ea8f5a78d77c4ca355b3039bbcb282",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "3d01a6e11fb4110b6c7f2f63c113c7b7c7ea8f5a78d77c4ca355b3039bbcb282",
    )
