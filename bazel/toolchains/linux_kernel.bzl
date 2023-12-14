"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/958d01bae15cc32b1b884d88717e47f2a374d428bb5efa5799bdbc0842851c32",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.67-100.constellation/kernel-6.1.67-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel.rpm",
        sha256 = "958d01bae15cc32b1b884d88717e47f2a374d428bb5efa5799bdbc0842851c32",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ae5dcef02065b7e8bf7e0cd45a2ee42a8d6f0235facb935edecc076f32277e60",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.67-100.constellation/kernel-core-6.1.67-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core.rpm",
        sha256 = "ae5dcef02065b7e8bf7e0cd45a2ee42a8d6f0235facb935edecc076f32277e60",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f72f83bf0824af1edcc37bcf2bfac0f7e1a14b40d0d7564854a6bea2c17e7a4a",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.67-100.constellation/kernel-modules-6.1.67-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules.rpm",
        sha256 = "f72f83bf0824af1edcc37bcf2bfac0f7e1a14b40d0d7564854a6bea2c17e7a4a",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7b44fc0488603158af811ef3b7b6ea72e2aa0162cd6df20a48c2121d0cf2719",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.67-100.constellation/kernel-modules-core-6.1.67-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core.rpm",
        sha256 = "e7b44fc0488603158af811ef3b7b6ea72e2aa0162cd6df20a48c2121d0cf2719",
    )
