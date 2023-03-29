"""A module defining the source of node maintainence operator."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def node_maintainance_operator_deps():
    http_archive(
        name = "com_github_medik8s_node_maintainance_operator",
        urls = [
            "https://github.com/medik8s/node-maintenance-operator/archive/refs/tags/v0.14.0.tar.gz",
        ],
        sha256 = "048323ffdb55787df9b93d85be93e4730f4495fba81b440dc6fe195408ec2533",
        strip_prefix = "node-maintenance-operator-0.14.0",
        build_file_content = """
api_v1beta1 = glob(["api/v1beta1/*.go"])
filegroup(
    srcs = api_v1beta1,
    name = "api_v1beta1",
    visibility = ["//visibility:public"],
)
        """,
    )
