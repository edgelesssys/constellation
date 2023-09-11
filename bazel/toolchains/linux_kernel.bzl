"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    http_file(
        name = "kernel",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4107ce34e10476036e44afd72a4bf93c4139a9fae811c00e2b4b8e9480157cee",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.46-100.constellation/kernel-6.1.46-100.constellation.fc38.x86_64.rpm",
        ],
        sha256 = "4107ce34e10476036e44afd72a4bf93c4139a9fae811c00e2b4b8e9480157cee",
        downloaded_file_path = "kernel.rpm",
    )
    http_file(
        name = "kernel_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6ba6f6cf0c4ee03044e61a596d9fd20e5383cbea56f7c42b1f003aefe295491a",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.46-100.constellation/kernel-core-6.1.46-100.constellation.fc38.x86_64.rpm",
        ],
        sha256 = "6ba6f6cf0c4ee03044e61a596d9fd20e5383cbea56f7c42b1f003aefe295491a",
        downloaded_file_path = "kernel-core.rpm",
    )
    http_file(
        name = "kernel_modules",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5111c00a8b28074af8354c551e6543c5656332192d2f491149ecf4d1d42e9044",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.46-100.constellation/kernel-modules-6.1.46-100.constellation.fc38.x86_64.rpm",
        ],
        sha256 = "5111c00a8b28074af8354c551e6543c5656332192d2f491149ecf4d1d42e9044",
        downloaded_file_path = "kernel-modules.rpm",
    )
    http_file(
        name = "kernel_modules_core",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52c0a21c57bbc95cbc4ef403806b9dbd9f18cbbbd2d61386a6eb3cc8cb47fa55",
            "https://cdn.confidential.cloud/constellation/kernel/6.1.46-100.constellation/kernel-modules-core-6.1.46-100.constellation.fc38.x86_64.rpm",
        ],
        sha256 = "52c0a21c57bbc95cbc4ef403806b9dbd9f18cbbbd2d61386a6eb3cc8cb47fa55",
        downloaded_file_path = "kernel-modules-core.rpm",
    )
