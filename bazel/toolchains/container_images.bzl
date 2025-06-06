"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:d9f9472a8f4541368192d714a995eb1a99bab1f7071fc8bde261d7eda3b667d8",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
    oci_pull(
        name = "libvirtd_base",
        digest = "sha256:f23e0f587860c841adde25b1b4f0d99aa4fbce1c92b01b5b46ab5fa35980a135",
        image = "ghcr.io/edgelesssys/constellation/libvirtd-base",
    )
