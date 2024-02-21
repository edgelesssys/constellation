"""versioned kubernetes release artifacts (dl.k8s.io)"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def k8s_deps():
    envtest_deps()

def envtest_deps():
    http_archive(
        name = "kubebuilder_tools_darwin_amd64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b62112fc83e503b35eab4fa5f8f5f648fcc4781ec319e6844644b4502eb8e2f1",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.29.1-darwin-amd64.tar.gz",
        ],
        sha256 = "b62112fc83e503b35eab4fa5f8f5f648fcc4781ec319e6844644b4502eb8e2f1",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_darwin_arm64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/61b0eea4a7095cf0c318bd35f8c16b8f5107c5af65a710abf5a95f9ed29fd593",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.29.1-darwin-arm64.tar.gz",
        ],
        sha256 = "61b0eea4a7095cf0c318bd35f8c16b8f5107c5af65a710abf5a95f9ed29fd593",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_linux_amd64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e37c2cda216692e699ce40ac2067dac8d773654f4afb20a90b92e6aafbe1593c",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.29.1-linux-amd64.tar.gz",
        ],
        sha256 = "e37c2cda216692e699ce40ac2067dac8d773654f4afb20a90b92e6aafbe1593c",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_linux_arm64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/956783315fde4371e775b1c7f115c996494b8603b3bf75cb528b0df9b6536382",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.29.1-linux-arm64.tar.gz",
        ],
        sha256 = "956783315fde4371e775b1c7f115c996494b8603b3bf75cb528b0df9b6536382",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
