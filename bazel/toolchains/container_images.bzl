"""
This file contains container images that are pulled from container registries.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
    oci_pull(
        name = "libvirtd_base",
        digest = "sha256:48ba2401ea66490ca1997b9d3e72b4bef7557ffbcdb1c95651fb3308f32fda58",
        image = "ghcr.io/edgelesssys/constellation/libvirtd-base",
    )
