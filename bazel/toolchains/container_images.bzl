"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:b7b9a6953e7bed6baaf37329331051d7bdc1b99c885f6dbeb72d75b1baad54f9",
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
