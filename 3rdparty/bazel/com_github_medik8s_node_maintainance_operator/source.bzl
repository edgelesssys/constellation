"""A module defining the source of node maintainence operator."""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def node_maintainance_operator_deps():
    http_archive(
        name = "com_github_medik8s_node_maintainance_operator",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/df5ea2f9d982dd78770f2549333fd40aaf40e50a28deec9d7892f83cf9d1bdb2",
            "https://github.com/medik8s/node-maintenance-operator/archive/refs/tags/v0.17.0.tar.gz",
        ],
        strip_prefix = "node-maintenance-operator-0.17.0",
        build_file_content = """
api_v1beta1 = glob(["api/v1beta1/*.go"])
filegroup(
    srcs = api_v1beta1,
    name = "api_v1beta1",
    visibility = ["//visibility:public"],
)
        """,
        type = "tar.gz",
        sha256 = "df5ea2f9d982dd78770f2549333fd40aaf40e50a28deec9d7892f83cf9d1bdb2",
    )
