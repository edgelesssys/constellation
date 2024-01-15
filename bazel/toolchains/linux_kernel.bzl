"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4430d2f8076081291d505ccb91bc84e3a763e113348e23775cc01df5a574d684",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "4430d2f8076081291d505ccb91bc84e3a763e113348e23775cc01df5a574d684",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3f9a42c4c86d56cae98053d3fc099368cbcf6dfa8ed48848e24e2c82ae3b7cc",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-core-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "e3f9a42c4c86d56cae98053d3fc099368cbcf6dfa8ed48848e24e2c82ae3b7cc",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/072fc8e1b1bb37e1cc40038f60e21a7be374d801f48589146660ffe7028f6b39",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-modules-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "072fc8e1b1bb37e1cc40038f60e21a7be374d801f48589146660ffe7028f6b39",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97d1e099b874d53b87fad2515c450b33d56770236211bf6a83a52e9e28361be1",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-modules-core-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "97d1e099b874d53b87fad2515c450b33d56770236211bf6a83a52e9e28361be1",
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
