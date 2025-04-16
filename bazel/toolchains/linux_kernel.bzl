"""kernel rpms"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kernel_rpms():
    """kernel rpms"""

    # LTS kernel
    http_file(
        name = "kernel_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7834bc4bc7e088c98505956382884bdc670ab9a9283288b7fef04a43df31356e",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.87-100.constellation/kernel-6.6.87-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-lts.rpm",
        sha256 = "7834bc4bc7e088c98505956382884bdc670ab9a9283288b7fef04a43df31356e",
    )
    http_file(
        name = "kernel_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2763c699d1e2f9810421ac7af2e9c94c6f98533e83f2938c26f1d824e3559b97",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.87-100.constellation/kernel-core-6.6.87-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-lts.rpm",
        sha256 = "2763c699d1e2f9810421ac7af2e9c94c6f98533e83f2938c26f1d824e3559b97",
    )
    http_file(
        name = "kernel_modules_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a7604eec263f190db573d809d20336bbf75e46c51f5977f5db95bb88bfec56d3",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.87-100.constellation/kernel-modules-6.6.87-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-lts.rpm",
        sha256 = "a7604eec263f190db573d809d20336bbf75e46c51f5977f5db95bb88bfec56d3",
    )
    http_file(
        name = "kernel_modules_core_lts",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/648fd503d7d54608fbd62ace87c4da098f72abbaac1ab7e343327fc24ccef7f8",
            "https://cdn.confidential.cloud/constellation/kernel/6.6.87-100.constellation/kernel-modules-core-6.6.87-100.constellation.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-lts.rpm",
        sha256 = "648fd503d7d54608fbd62ace87c4da098f72abbaac1ab7e343327fc24ccef7f8",
    )

    # mainline kernel
    http_file(
        name = "kernel_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6eaec29870e6549d95a93b72ea10715507db84b851c68c0d75e44e4c20f895f2",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-mainline.rpm",
        sha256 = "6eaec29870e6549d95a93b72ea10715507db84b851c68c0d75e44e4c20f895f2",
    )
    http_file(
        name = "kernel_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/910fd35209f7dc8185e88dddeaccf6158dd63ad9fd469ef3dc81b96840ef28eb",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-core-mainline.rpm",
        sha256 = "910fd35209f7dc8185e88dddeaccf6158dd63ad9fd469ef3dc81b96840ef28eb",
    )
    http_file(
        name = "kernel_modules_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8de20433c68d2fe0ca6625e25f314aba36a9327592db8b1478b97bb50521149",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-mainline.rpm",
        sha256 = "b8de20433c68d2fe0ca6625e25f314aba36a9327592db8b1478b97bb50521149",
    )
    http_file(
        name = "kernel_modules_core_mainline",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8ecd8e96483810d18e04a20cd8ecef46f27bff0fbb54f23e67adb813828b3cec",
            "https://kojipkgs.fedoraproject.org/packages/kernel/6.8.9/300.fc40/x86_64/kernel-modules-core-6.8.9-300.fc40.x86_64.rpm",
        ],
        downloaded_file_path = "kernel-modules-core-mainline.rpm",
        sha256 = "8ecd8e96483810d18e04a20cd8ecef46f27bff0fbb54f23e67adb813828b3cec",
    )
