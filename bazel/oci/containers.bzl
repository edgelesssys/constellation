"""
This module holds the definitions of the containers that are built.
"""

load("@rules_oci//oci:defs.bzl", _oci_push = "oci_push", _oci_tarball = "oci_tarball")
load("//bazel/oci:pin.bzl", "oci_sum")

_default_registry = "ghcr.io"
_default_prefix = "edgelesssys/constellation"

def containers():
    return [
        {
            "identifier": "joinService",
            "image_name": "join-service",
            "name": "joinservice",
            "oci": "//joinservice/cmd:joinservice",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["helm"],
        },
        {
            "identifier": "keyService",
            "image_name": "key-service",
            "name": "keyservice",
            "oci": "//keyservice/cmd:keyservice",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["helm"],
        },
        {
            "identifier": "verificationService",
            "image_name": "verification-service",
            "name": "verificationservice",
            "oci": "//verify/cmd:verificationservice",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["helm"],
        },
        {
            "identifier": "constellationNodeOperator",
            "image_name": "node-operator",
            "name": "nodeoperator",
            "oci": "//operators/constellation-node-operator:node_operator",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["helm"],
        },
        {
            "identifier": "qemuMetadata",
            "image_name": "qemu-metadata-api",
            "name": "qemumetadata",
            "oci": "//hack/qemu-metadata-api:qemumetadata",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["config"],
        },
        {
            "identifier": "libvirt",
            "image_name": "libvirt",
            "name": "libvirt",
            "oci": "//cli/internal/libvirt:constellation_libvirt",
            "prefix": _default_prefix,
            "registry": _default_registry,
            "tag_file": "//bazel/settings:tag",
            "used_by": ["config"],
        },
    ]

def helm_containers():
    return [container for container in containers() if "helm" in container["used_by"]]

def config_containers():
    return [container for container in containers() if "config" in container["used_by"]]

def container_sum(name, oci, registry, prefix, image_name, **kwargs):
    tag = kwargs.get("tag", None)
    tag_file = kwargs.get("tag_file", None)
    oci_sum(
        name = name + "_sum",
        oci = oci,
        registry = registry,
        prefix = prefix,
        image_name = image_name,
        tag = tag,
        tag_file = tag_file,
        visibility = ["//visibility:public"],
    )

def oci_push(name, image, registry, image_name, **kwargs):
    """oci_push pushes an OCI image to a registry.

    Args:
        name: The name of the target.
        image: The OCI image to push.
        registry: The registry to push to.
        image_name: The name of the image.
        **kwargs: Additional arguments to pass to oci_push.
    """
    prefix = kwargs.pop("prefix", None)
    tag = kwargs.pop("tag", None)
    tag_file = kwargs.pop("tag_file", None)
    if prefix == None:
        repository = registry + "/" + image_name
    else:
        repository = registry + "/" + prefix + "/" + image_name
    _oci_push(
        name = name,
        image = image,
        repository = repository,
        tag = tag,
        tag_file = tag_file,
        visibility = ["//visibility:public"],
        **kwargs
    )

# TODO(malt3): allow repotags (registry + tag) to be read from a file.
def oci_tarball(name, image):
    """oci_tarball creates a tarball of an OCI image.

    Args:
        name: The name of the target.
        image: The OCI image to create a tarball of.
    """
    _oci_tarball(
        name = name,
        image = image,
        repotags = [],
        visibility = ["//visibility:public"],
    )
