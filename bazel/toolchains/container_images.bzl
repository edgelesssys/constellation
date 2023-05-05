"""
This file contains external container images used by the project.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:7198a357ff3a8ef750b041324873960cf2153c11cc50abb9d8d5f8bb089f6b4e",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
