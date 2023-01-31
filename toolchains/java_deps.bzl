"""Java toolchain dependencies for Bazel.

Defines hermetic java toolchains.
"""

load("@bazel_tools//tools/jdk:remote_java_repository.bzl", "remote_java_repository")

HASHES = {
    "linux_aarch64": "9f5ac83b584a297c792cc5feb67c752a2d9fc1259abec3a477e96be8b672f452",
    "linux_x86_64": "6fae6811b0f3aebb14c3e59a5fde14481cff412ef8ca23221993f1ab33269aab",
    "osx_aarch64": "2a3f56af83f9d180dfce5d6e771a292bbbd68a77c7c18ed3bdb607e86d773704",
    "osx_x86_64": "6234ebb7567c416ff28e2f080569e67656ae8fcdb3b601d8348d4d504ca79e68",
}

OS = {
    "linux": "linux",
    "osx": "macosx",
}

ARCH = {
    "aarch64": "aarch64",
    "x86_64": "x64",
}

def java_deps():
    for os in OS:
        for arch in ["x86_64", "aarch64"]:
            _java_repository(
                os = os,
                arch = arch,
                zulu_version = "11.62.17",
                jdk_version = "11.0.18",
            )

def _java_repository(os, arch, zulu_version, jdk_version, jdk_major = 11):
    """Defines a java repository for the given os and architecture.

    Args:
      os: The os of the java repository.
      arch: The architecture of the java repository.
      jdk_major: The major version of the java repository.
      zulu_version: The zulu version of the java repository.
      jdk_version: The jdk version of the java repository.
    """
    path = "zulu-embedded" if os == "linux" and arch == "aarch64" else "zulu"
    remote_java_repository(
        name = "pinned_remotejdk%s_%s_%s" % (jdk_major, os, arch),
        prefix = "pinned_remotejdk",
        sha256 = HASHES["%s_%s" % (os, arch)],
        strip_prefix = "zulu%s-ca-jdk%s-%s_%s" % (zulu_version, jdk_version, OS[os], ARCH[arch]),
        target_compatible_with = [
            "@platforms//os:" + os,
            "@platforms//cpu:" + arch,
        ],
        urls = [
            "https://cdn.azul.com/%s/bin/zulu%s-ca-jdk%s-%s_%s.tar.gz" % (path, zulu_version, jdk_version, OS[os], ARCH[arch]),
        ],
        version = jdk_major,
    )
