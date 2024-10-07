"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:69830f29ed7545c762777507426a412f97dad3d8d32bae3e74ad3fb6160917ea",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
    oci_pull(
        name = "libvirtd_base",
        digest = "sha256:99dbf3cf69b3f97cb0158bde152c9bc7c2a96458cf462527ee80b75754f572a7",
        image = "ghcr.io/edgelesssys/constellation/libvirtd-base",
    )
