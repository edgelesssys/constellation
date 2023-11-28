"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:6706c73aae2afaa8201d63cc3dda48753c09bcd6c300762251065c0f7e602b25",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
    oci_pull(
        name = "libvirtd_base",
        digest = "sha256:f5aca956c8d67059725feb4bf8a7d96da71a51efe84288c74a52fcf6855a13bd",
        image = "ghcr.io/edgelesssys/constellation/libvirtd-base",
    )
