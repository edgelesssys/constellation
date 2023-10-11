"""
This file contains external container images used by the project.
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
