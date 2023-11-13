"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8d6f7302af40109b031ff97b330238f8b4dec6d5bfb94a1b7a9ced709e6439a",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.62-100.constellation/kernel-6.1.62-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel.rpm",
        sha256 = "a8d6f7302af40109b031ff97b330238f8b4dec6d5bfb94a1b7a9ced709e6439a",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a5c196197abbeb29f04ad14be39916b3d9f773d0ed786a3c7b46fd462daf4297",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.62-100.constellation/kernel-core-6.1.62-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core.rpm",
        sha256 = "a5c196197abbeb29f04ad14be39916b3d9f773d0ed786a3c7b46fd462daf4297",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e2a3df5fb33c020aefb9a209b2e531088a5192d8ecd4824747166f5e0230fb8d",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.62-100.constellation/kernel-modules-6.1.62-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules.rpm",
        sha256 = "e2a3df5fb33c020aefb9a209b2e531088a5192d8ecd4824747166f5e0230fb8d",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/88df5e3afa4beff12fe976339c273886372c6eb841c3407a237aa60c52fb9f00",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.62-100.constellation/kernel-modules-core-6.1.62-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core.rpm",
        sha256 = "88df5e3afa4beff12fe976339c273886372c6eb841c3407a237aa60c52fb9f00",
    )
