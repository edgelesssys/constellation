"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9b19efd281645344a54b31e83900652136901c795dca4021385ec1e2946b725d",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.64-100.constellation/kernel-6.1.64-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel.rpm",
        sha256 = "9b19efd281645344a54b31e83900652136901c795dca4021385ec1e2946b725d",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/46a7b1697fce01cad3c2517ae7d7b35f022296ddf5c615d8b73c633990a23313",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.64-100.constellation/kernel-core-6.1.64-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core.rpm",
        sha256 = "46a7b1697fce01cad3c2517ae7d7b35f022296ddf5c615d8b73c633990a23313",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b842238c14df63545922dcdcab018c6dc2077dd92853d4b877732628aafc1abe",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.64-100.constellation/kernel-modules-6.1.64-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules.rpm",
        sha256 = "b842238c14df63545922dcdcab018c6dc2077dd92853d4b877732628aafc1abe",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ab09c4bc8d41676d464475ff29c064a5655985d98554f657cb5b7257358f1b5f",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.64-100.constellation/kernel-modules-core-6.1.64-100.constellation.fc38.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core.rpm",
        sha256 = "ab09c4bc8d41676d464475ff29c064a5655985d98554f657cb5b7257358f1b5f",
    )
