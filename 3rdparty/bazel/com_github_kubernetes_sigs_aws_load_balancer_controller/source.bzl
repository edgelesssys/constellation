"""A module defining the source of the AWS load balancer controller."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aws_load_balancer_controller_deps():
    http_archive(
        name = "com_github_kubernetes_sigs_aws_load_balancer_controller",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/422af7c03ebc73e1be6aea563475ec9ea6396071fa03158b9a3984aa621b8cb1",
            "https://github.com/kubernetes-sigs/aws-load-balancer-controller/archive/refs/tags/v2.13.3.tar.gz",
        ],
        strip_prefix = "aws-load-balancer-controller-2.13.3",
        build_file_content = """
filegroup(
    srcs = ["docs/install/iam_policy.json"],
    name = "lb_policy",
    visibility = ["//visibility:public"],
)
        """,
        type = "tar.gz",
        sha256 = "422af7c03ebc73e1be6aea563475ec9ea6396071fa03158b9a3984aa621b8cb1",
    )
