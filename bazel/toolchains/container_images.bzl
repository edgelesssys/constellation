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
        digest = "sha256:033a9ff81e5273a29e61fa5b780427c31f2a5317545b238128626a10e41599dc",
        image = "ghcr.io/edgelesssys/constellation/libvirtd-base",
    )
