""" This file is used to pin / load external RPMs that are used by the project. """

load("@bazeldnf//:deps.bzl", "rpm")

def rpms():
    """ Provides a list of RPMs that are used by the project. """
    rpm(
        name = "alternatives-0__1.21-1.fc37.x86_64",
        sha256 = "90787668e5f26eb2a87ceff11fb0594f87d616b909394f8d68d4564e3f6e4568",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "audit-libs-0__3.0.9-1.fc37.x86_64",
        sha256 = "23cda2a639c358757ee35ce6270ba3d0c6cd779309ff528e7c08c0239737dffb",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "basesystem-0__11-14.fc37.x86_64",
        sha256 = "38d1877d647bb5f4047d22982a51899c95bdfea1d7b2debbff37c66f0fc0ed44",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "bash-0__5.2.15-1.fc37.x86_64",
        sha256 = "e50ddbdb35ecec1a9bf4e19fd87c6216382be313c3b671704d444053a1cfd183",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "bzip2-libs-0__1.0.8-12.fc37.x86_64",
        sha256 = "6e74a8ed5b472cf811f9bf429a999ed3f362e2c88566a461517a12c058abd401",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "ca-certificates-0__2023.2.60-1.0.fc37.x86_64",
        sha256 = "b2dcac3e49cbf75841d41ee1c53f1a91ffa78ba03dab8febb3153dbf76b2c5b2",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "coreutils-single-0__9.1-7.fc37.x86_64",
        sha256 = "414bda840560471cb3d7380923ab00585ee78ca2db4b0d52155e9319a32151bc",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "crypto-policies-0__20220815-1.gite4ed860.fc37.x86_64",
        sha256 = "486a11feeaad706c68b05de60a906cc57059454cbce436aeba45f88b84578c0c",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "cryptsetup-devel-0__2.5.0-1.fc37.x86_64",
        sha256 = "dc7a6b834db44483d8d313fe7658a553ca721cebac99bf1db29a8db483e07964",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cryptsetup-libs-0__2.5.0-1.fc37.x86_64",
        sha256 = "9a9c9f908326ce672180964c0dee6a387fefce9f4e49dacbca87f4aa8bf1e31f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-1__1.14.4-1.fc37.x86_64",
        sha256 = "2a9382a55160f297e86069e50b85f47df5546d0cacd3421bd1e79a69806b297a",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-broker-0__32-1.fc37.x86_64",
        sha256 = "e5bbfce30b88c0b4f06c4ad0c80645cfec9c23248d7c734d76607d8bc500c43f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-common-1__1.14.4-1.fc37.x86_64",
        sha256 = "f2fe0d92a66d642759682ef3818d09519a499807bfb6b50b012b4178bf5e58f7",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "device-mapper-0__1.02.175-9.fc37.x86_64",
        sha256 = "c57831b8629e2e31b3c55d4f0064cd25a515d3eb1ac61fc6897ce07421a2e91b",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "device-mapper-libs-0__1.02.175-9.fc37.x86_64",
        sha256 = "7c0f72217eacc9b5caf553c17cb2428de242094dc7e0e1dbd0d21869d909c7d2",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "expat-0__2.5.0-1.fc37.x86_64",
        sha256 = "0e49c2393e5507bbaa16ededf0176e731e0196dd3230f6371d67be8b919e3429",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "fedora-gpg-keys-0__37-1.x86_64",
        sha256 = "bf315e20968e291f76bd566087c26ae3e19a36f3e3f80511058e8253ac8d3352",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-gpg-keys-37-1.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-gpg-keys-37-1.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-gpg-keys-37-1.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-gpg-keys-37-1.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-common-0__37-15.x86_64",
        sha256 = "4a3013afe17b6e1413f8999c977ed4f8bd9c3d735f2f7bb066e7b021840934bb",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-container-0__37-15.x86_64",
        sha256 = "d4f3dccba997b2c72db48cb88fed1341d2aa1dfadd4b662c472da2f269fc4a85",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-identity-kinoite-0__37-15.x86_64",
        sha256 = "6101f000f53e418c9f7e9faabeef0ffd7cf2a39fe44b7784a048f4c94123f3b2",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-kinoite-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-kinoite-37-15.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-kinoite-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-kinoite-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-repos-0__37-1.x86_64",
        sha256 = "97ecefcabcbe0da05f0fc80847f6c4d3334d3698b62098bc7280ef833a213400",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-repos-37-1.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-repos-37-1.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-repos-37-1.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/fedora-repos-37-1.noarch.rpm",
        ],
    )

    rpm(
        name = "filesystem-0__3.18-2.fc37.x86_64",
        sha256 = "1c28f722e7f3e48dba7ebf4f763ebebc6688b9e0fd58b55ba4fcd884c8180ef4",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-0__2.36-9.fc37.x86_64",
        sha256 = "8c8463cd9f194f03ea1607670399e2fbf068857f566c43dd07d351228c25f187",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-common-0__2.36-9.fc37.x86_64",
        sha256 = "4237c10e5edacc5d5a9ea88e9fc5fef37249d459b13d4a0715c7836374a8da7a",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-langpack-ku-0__2.36-9.fc37.x86_64",
        sha256 = "8b2ba291bbe52282989a95cf59406f6436cffabd01dcf60ecf1d934770bb34b1",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-ku-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-ku-2.36-9.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-ku-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-ku-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "grep-0__3.7-4.fc37.x86_64",
        sha256 = "d997786e71f2c7b4a9ed1323b8684ec1802e49a866fb0c1b69101531440cb464",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "json-c-0__0.16-3.fc37.x86_64",
        sha256 = "e7c83a9058c7e7e05e4c7ba97a363414eb973343ea8f00a1140fbdafe6ca67e2",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "kmod-libs-0__30-2.fc37.x86_64",
        sha256 = "73a1a0f041819c1d50501a699945f0121a3b6e1f54df40cd0bf8f94b1b261ef5",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libacl-0__2.3.1-4.fc37.x86_64",
        sha256 = "15224cb92199b8011fe47dc12e0bbcdbee0c93e0f29553b3b07ae41768b48ce3",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libargon2-0__20190702-1.fc37.x86_64",
        sha256 = "bf280bf9e59891bfcb4a987d5df22d6a6d9f60589dd00b790b5a3047a727a40b",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libattr-0__2.5.1-5.fc37.x86_64",
        sha256 = "3a423be562953538eaa0d1e78ef35890396cdf1ad89561c619aa72d3a59bfb82",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libblkid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b0388d1a529bf6b54ca648e91529b1e7790e6aaa42e0ac2b7be6640e4f24a21d",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcap-0__2.48-5.fc37.x86_64",
        sha256 = "aa22373907b6ff9fa3d2f7d9e33a9bdefc9ac50486f2dac5251ac4e206a8a61d",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcap-ng-0__0.8.3-3.fc37.x86_64",
        sha256 = "bcca8a17ae16f9f1c8664f9f54e8f2178f028821f6802ebf33cdcd2d4289bf7f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libeconf-0__0.4.0-4.fc37.x86_64",
        sha256 = "f0cc1addee779f09aade289e3be4e9bd103a274a6bdf11f8331878686f432653",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libfdisk-0__2.38.1-1.fc37.x86_64",
        sha256 = "7a4bd1f4975a52fc201c9bc978f155dcb97212cb970210525d903b03644a713d",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libffi-0__3.4.2-9.fc37.x86_64",
        sha256 = "dc5f7ca1ce86cd9380525b383624bdc1afa52d98db624cfdece7b08086f829d6",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libgcc-0__12.2.1-4.fc37.x86_64",
        sha256 = "25299b673e7488f538c6d0433ea7fe0ffc8311e41dd7115b5985145e493e4b05",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libmount-0__2.38.1-1.fc37.x86_64",
        sha256 = "50c304faa94d7959e5cbc0642b3c77539ad000042e6617ea5da4789c8105496f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libpkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "ecd52fd3f3065606ba5164249b29c837cbd172643d13a00a1a72fc657b115af7",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libseccomp-0__2.5.3-3.fc37.x86_64",
        sha256 = "017877a97c8222fc7eca7fab77600a3a1fcdec92f9dd39d8df6e64726909fcbe",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libselinux-0__3.4-5.fc37.x86_64",
        sha256 = "2a5b4e2e1dd388c3c13d79af971fb8efd522f5c2ba8d257875f02c16b4858214",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsemanage-0__3.4-5.fc37.x86_64",
        sha256 = "66305d7a3ca92165f1c17e14cc29ea70280fa1c1fd3bf223b5b1d4f7d1ce0dd8",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsepol-0__3.4-3.fc37.x86_64",
        sha256 = "97e918bc5b11c8abfec9343e1b0bd88087b792e85c604427d5cdc32733f70b3f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsmartcols-0__2.38.1-1.fc37.x86_64",
        sha256 = "93246c002aefec27bb398aa3397ae555bcc3035b10aebb4937c4bea9268bacf1",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libtasn1-0__4.19.0-1.fc37.x86_64",
        sha256 = "35b51a0796af6930b2a8a511df8c51938006cfcfdf74ddfe6482eb9febd87dfa",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libuuid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b054577d98aa9615fe459abec31be46b19ad72e0da620d8d251b4449a6db020d",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libxcrypt-0__4.4.33-4.fc37.x86_64",
        sha256 = "547b9cffb0211abc4445d159e944f4fb59606b2eddfc14813b8c068859294ba6",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libzstd-0__1.5.2-3.fc37.x86_64",
        sha256 = "c105f54738fba9793dad9b6ab2e88b4ae05cc47b9ea062cd7215b69d11ce0e1c",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "lz4-libs-0__1.9.4-1.fc37.x86_64",
        sha256 = "f39b8b018fcb2b55477cdbfa4af7c9db9b660c85000a4a42e880b1a951efbe5a",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "ncurses-base-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "000164a9a82458fbb69b3433801dcc0d0e2437e21d7f7d4fd45f63a42a0bc26f",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "ncurses-libs-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "75e51eebcd3fe150b421ec5b1c9a6e918caa5b3c0f243f2b70d445fd434488bb",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "openssl-libs-1__3.0.5-3.fc37.x86_64",
        sha256 = "76fdbe6d7d4cd898d73da5a36fcbbfd2330a4855a2dc3e023480e23caf4eac8c",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "p11-kit-0__0.24.1-3.fc37.x86_64",
        sha256 = "4dad6ac54eb7708cbfc8522d372f2a196cf711e97e279cbddba8cc8b92970dd7",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "p11-kit-trust-0__0.24.1-3.fc37.x86_64",
        sha256 = "0fd85eb1ce27615fea745721b18648b4a4585ad4b11a482c1b77fc1785cd5194",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pam-libs-0__1.5.2-14.fc37.x86_64",
        sha256 = "ee34422adc6451da744bd16a8cd66c9912a822c4e55227c23ff56960c32980f5",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre-0__8.45-1.fc37.2.x86_64",
        sha256 = "86a648e3b88f581b15ca2eda6b441be7c5c3810a9eae25ca940c767029e4e923",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre2-0__10.40-1.fc37.1.x86_64",
        sha256 = "422de947ec1a7aafcd212a51e64257b64d5b0a02808104a33e7c3cd9ef629148",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre2-syntax-0__10.40-1.fc37.1.x86_64",
        sha256 = "585f339942a0bf4b0eab638ddf825544793485cbcb9f1eaee079b9956d90aafa",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
        ],
    )

    rpm(
        name = "pkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "778018594ab5bddc4432e53985b80e6c5a1a1ec1700d38b438848d485f5b357c",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pkgconf-m4-0__1.8.0-3.fc37.x86_64",
        sha256 = "dd0356475d0b9106b5a2d577db359aa0290fe6dd9eacea1b6e0cab816ff33566",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "pkgconf-pkg-config-0__1.8.0-3.fc37.x86_64",
        sha256 = "d238b12c750b58ceebc80e25c2074bd929d3f232c1390677f33a94fdadb68f6a",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "sed-0__4.8-11.fc37.x86_64",
        sha256 = "231e782077862f4abecf025aa254a9c391a950490ae856261dcfd229863ac80f",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "setup-0__2.14.1-2.fc37.x86_64",
        sha256 = "15d72b2a44f403b3a7ee9138820a8ce7584f954aeafbb43b1251621bca26f785",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "shadow-utils-2__4.12.3-4.fc37.x86_64",
        sha256 = "8394db7e5385d64c90876a0ee9274c3f53ae91f172c051524a24a30963e18fb8",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-0__251.10-588.fc37.x86_64",
        sha256 = "f8d45c2c70b06ba15eb2601bc1bc83d030b5df059525964456b98ef45c2f146e",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-libs-0__251.10-588.fc37.x86_64",
        sha256 = "b383c219341e141d003ae0722fcea5e6949cf1af41adc64d290292940b239bca",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-pam-0__251.10-588.fc37.x86_64",
        sha256 = "4d30204cbb169f9fdbbc872a71d6ddea95f442fa3489fb075fe1d06ff9222c48",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "tzdata-0__2022g-1.fc37.x86_64",
        sha256 = "7ff35c66b3478103fbf3941e933e25f60e41f2b0bfd07d43666b40721211c3bb",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "util-linux-core-0__2.38.1-1.fc37.x86_64",
        sha256 = "f87ad8fc18f4da254966cc6f99b533dc8125e1ec0eaefd5f89a6b6398cb13a34",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "xz-libs-0__5.4.1-1.fc37.x86_64",
        sha256 = "8c06eef8dd28d6dc1406e65e4eb8ee3db359cf6624729be4e426f6b01c4117fd",
        urls = [
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "zlib-0__1.2.12-5.fc37.x86_64",
        sha256 = "7b0eda1ad9e9a06e61d9fe41e5e4e0fbdc8427bc252f06a7d29cd7ba81a71a70",
        urls = [
            "https://mirror.netzwerge.de/fedora/linux/development/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/development/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/development/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/development/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
        ],
    )
