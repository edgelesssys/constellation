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
            "https://cdn.confidential.cloud/constellation/cas/sha256/e1913674bacaa70c067e15649237e1f67d891ba53f367c0a50786b4a274ee047",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.27.1-darwin-amd64.tar.gz",
        ],
        sha256 = "e1913674bacaa70c067e15649237e1f67d891ba53f367c0a50786b4a274ee047",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_darwin_arm64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0422632a2bbb0d4d14d7d8b0f05497a4d041c11d770a07b7a55c44bcc5e8ce66",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.27.1-darwin-arm64.tar.gz",
        ],
        sha256 = "0422632a2bbb0d4d14d7d8b0f05497a4d041c11d770a07b7a55c44bcc5e8ce66",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_linux_amd64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f9699df7b021f71a1ab55329b36b48a798e6ae3a44d2132255fc7e46c6790d4d",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.27.1-linux-amd64.tar.gz",
        ],
        sha256 = "f9699df7b021f71a1ab55329b36b48a798e6ae3a44d2132255fc7e46c6790d4d",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
    http_archive(
        name = "kubebuilder_tools_linux_arm64",
        strip_prefix = "kubebuilder/bin",
        type = "tar.gz",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9d2803e8ca85c465b33c12b06d0b2eba3ddb64b53a468628f741e50b462c46ad",
            "https://kubebuilder-tools.storage.googleapis.com/kubebuilder-tools-1.27.1-linux-arm64.tar.gz",
        ],
        sha256 = "9d2803e8ca85c465b33c12b06d0b2eba3ddb64b53a468628f741e50b462c46ad",
        build_file_content = """exports_files(["etcd", "kubectl", "kube-apiserver"], visibility = ["//visibility:public"])""",
    )
