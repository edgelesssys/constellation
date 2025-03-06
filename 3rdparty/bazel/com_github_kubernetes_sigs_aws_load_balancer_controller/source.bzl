"""A module defining the source of the AWS load balancer controller."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def aws_load_balancer_controller_deps():
    http_archive(
        name = "com_github_kubernetes_sigs_aws_load_balancer_controller",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0cb78cdff9742945c9968ac12c785164a052b52260d19d218bb28a8bec04a2fd",
            "https://github.com/kubernetes-sigs/aws-load-balancer-controller/archive/refs/tags/v2.11.0.tar.gz",
        ],
        strip_prefix = "aws-load-balancer-controller-2.11.0",
        build_file_content = """
filegroup(
    srcs = ["docs/install/iam_policy.json"],
    name = "lb_policy",
    visibility = ["//visibility:public"],
)
        """,
        type = "tar.gz",
        sha256 = "0cb78cdff9742945c9968ac12c785164a052b52260d19d218bb28a8bec04a2fd",
    )
