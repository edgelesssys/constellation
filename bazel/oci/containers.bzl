"""
This module holds the definitions of the containers that are built.
"""

load("@bazel_skylib//lib:paths.bzl", "paths")
load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")

def containers():
    return [
        {
            "identifier": "joinService",
            "image_name": "join-service",
            "name": "joinservice",
            "oci": "//joinservice/cmd:joinservice",
            "repotag_file": "//bazel/release:joinservice_tag.txt",
            "used_by": ["helm"],
        },
        {
            "identifier": "keyService",
            "image_name": "key-service",
            "name": "keyservice",
            "oci": "//keyservice/cmd:keyservice",
            "repotag_file": "//bazel/release:keyservice_tag.txt",
            "used_by": ["helm"],
        },
        {
            "identifier": "verificationService",
            "image_name": "verification-service",
            "name": "verificationservice",
            "oci": "//verify/cmd:verificationservice",
            "repotag_file": "//bazel/release:verificationservice_tag.txt",
            "used_by": ["helm"],
        },
        {
            "identifier": "constellationNodeOperator",
            "image_name": "node-operator",
            "name": "nodeoperator",
            "oci": "//operators/constellation-node-operator:node_operator",
            "repotag_file": "//bazel/release:nodeoperator_tag.txt",
            "used_by": ["helm"],
        },
        {
            "identifier": "qemuMetadata",
            "image_name": "qemu-metadata-api",
            "name": "qemumetadata",
            "oci": "//hack/qemu-metadata-api:qemumetadata",
            "repotag_file": "//bazel/release:qemumetadata_tag.txt",
            "used_by": ["config"],
        },
        {
            "identifier": "libvirt",
            "image_name": "libvirt",
            "name": "libvirt",
            "oci": "//cli/internal/libvirt:constellation_libvirt",
            "repotag_file": "//bazel/release:libvirt_tag.txt",
            "used_by": ["config"],
        },
        {
            "identifier": "s3proxy",
            "image_name": "s3proxy",
            "name": "s3proxy",
            "oci": "//s3proxy/cmd:s3proxy",
            "repotag_file": "//bazel/release:s3proxy_tag.txt",
            "used_by": ["config"],
        },
    ]

def helm_containers():
    return [container for container in containers() if "helm" in container["used_by"]]

def config_containers():
    return [container for container in containers() if "config" in container["used_by"]]

def _container_reponame_impl(ctx):
    container_prefix = ctx.attr._prefix[BuildSettingInfo].value
    if container_prefix == None:
        fail("container_prefix is not set")

    full_container_tag = paths.join(container_prefix, ctx.attr.container_name)

    output = ctx.actions.declare_file(ctx.attr.container_name + "_container_repotag")
    ctx.actions.write(output = output, content = full_container_tag)
    return [DefaultInfo(files = depset([output]))]

container_reponame = rule(
    implementation = _container_reponame_impl,
    attrs = {
        "container_name": attr.string(),
        "_prefix": attr.label(default = Label("//bazel/settings:container_prefix")),
    },
)
