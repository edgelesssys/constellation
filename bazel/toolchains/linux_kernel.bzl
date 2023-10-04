"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/24ae8a70fa5a81a73cbf17948564d3cc28b1b8fe68233f8dcaa7c3c7067b5425",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.55-100.constellation/kernel-6.1.55-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel.rpm",
        sha256 = "24ae8a70fa5a81a73cbf17948564d3cc28b1b8fe68233f8dcaa7c3c7067b5425",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/909ae1744f46e4e5827ee30742f014ff4290a81c1ba54f90f3c9f053e6e6d3f2",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.55-100.constellation/kernel-core-6.1.55-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core.rpm",
        sha256 = "909ae1744f46e4e5827ee30742f014ff4290a81c1ba54f90f3c9f053e6e6d3f2",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9cee46ad8a8bebebbacb90a788107496def716af0c9e1411d849a78ae9a999e",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.55-100.constellation/kernel-modules-6.1.55-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules.rpm",
        sha256 = "a9cee46ad8a8bebebbacb90a788107496def716af0c9e1411d849a78ae9a999e",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d75c60b0d08f8608456a5447e5939c93f80180a5b016ef4e0b0b2cc6549cbdd0",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.55-100.constellation/kernel-modules-core-6.1.55-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core.rpm",
        sha256 = "d75c60b0d08f8608456a5447e5939c93f80180a5b016ef4e0b0b2cc6549cbdd0",
    )
