"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4430d2f8076081291d505ccb91bc84e3a763e113348e23775cc01df5a574d684",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel.rpm",
        sha256 = "4430d2f8076081291d505ccb91bc84e3a763e113348e23775cc01df5a574d684",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3f9a42c4c86d56cae98053d3fc099368cbcf6dfa8ed48848e24e2c82ae3b7cc",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-core-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core.rpm",
        sha256 = "e3f9a42c4c86d56cae98053d3fc099368cbcf6dfa8ed48848e24e2c82ae3b7cc",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/072fc8e1b1bb37e1cc40038f60e21a7be374d801f48589146660ffe7028f6b39",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-modules-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules.rpm",
        sha256 = "072fc8e1b1bb37e1cc40038f60e21a7be374d801f48589146660ffe7028f6b39",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97d1e099b874d53b87fad2515c450b33d56770236211bf6a83a52e9e28361be1",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.68-100.constellation/kernel-modules-core-6.1.68-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core.rpm",
        sha256 = "97d1e099b874d53b87fad2515c450b33d56770236211bf6a83a52e9e28361be1",
    )
