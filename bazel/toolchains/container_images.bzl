"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:f2ff10a709b0fd153997059b698ada702e4870745b6077eff03a5f4850ca91b6",
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
