"""
This file contains external container images used by the project.
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

def containter_image_deps():
    oci_pull(
        name = "distroless_static",
        digest = "sha256:e7e79fb2947f38ce0fab6061733f7e1959c12b843079042fe13f56ca7b9d178c",
        image = "gcr.io/distroless/static",
        platforms = [
            "linux/amd64",
            "linux/arm64",
        ],
    )
