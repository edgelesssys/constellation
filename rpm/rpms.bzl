""" This file is used to pin / load external RPMs that are used by the project. """

load("@bazeldnf//:deps.bzl", "rpm")

def rpms():
    """ Provides a list of RPMs that are used by the project. """
    rpm(
        name = "alternatives-0__1.21-1.fc37.x86_64",
        sha256 = "90787668e5f26eb2a87ceff11fb0594f87d616b909394f8d68d4564e3f6e4568",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/90787668e5f26eb2a87ceff11fb0594f87d616b909394f8d68d4564e3f6e4568",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/alternatives-1.21-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "audit-libs-0__3.0.9-1.fc37.x86_64",
        sha256 = "23cda2a639c358757ee35ce6270ba3d0c6cd779309ff528e7c08c0239737dffb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/23cda2a639c358757ee35ce6270ba3d0c6cd779309ff528e7c08c0239737dffb",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/a/audit-libs-3.0.9-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "authselect-0__1.4.2-1.fc37.x86_64",
        sha256 = "c356d05e80f2b57ea2598b45b168fff6da189038e3f3ef0305dd90cfdd2a045f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c356d05e80f2b57ea2598b45b168fff6da189038e3f3ef0305dd90cfdd2a045f",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-1.4.2-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-1.4.2-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-1.4.2-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-1.4.2-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "authselect-libs-0__1.4.2-1.fc37.x86_64",
        sha256 = "275c282a240a3b7225e98b540a91af3419a9fa527623c5f152c48f8209779146",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/275c282a240a3b7225e98b540a91af3419a9fa527623c5f152c48f8209779146",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-libs-1.4.2-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-libs-1.4.2-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-libs-1.4.2-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/a/authselect-libs-1.4.2-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "basesystem-0__11-14.fc37.x86_64",
        sha256 = "38d1877d647bb5f4047d22982a51899c95bdfea1d7b2debbff37c66f0fc0ed44",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/38d1877d647bb5f4047d22982a51899c95bdfea1d7b2debbff37c66f0fc0ed44",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/aarch64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/aarch64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/aarch64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/aarch64/os/Packages/b/basesystem-11-14.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "bash-0__5.2.15-1.fc37.x86_64",
        sha256 = "e50ddbdb35ecec1a9bf4e19fd87c6216382be313c3b671704d444053a1cfd183",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e50ddbdb35ecec1a9bf4e19fd87c6216382be313c3b671704d444053a1cfd183",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/b/bash-5.2.15-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "bzip2-libs-0__1.0.8-12.fc37.x86_64",
        sha256 = "6e74a8ed5b472cf811f9bf429a999ed3f362e2c88566a461517a12c058abd401",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6e74a8ed5b472cf811f9bf429a999ed3f362e2c88566a461517a12c058abd401",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/b/bzip2-libs-1.0.8-12.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "ca-certificates-0__2023.2.60-1.0.fc37.x86_64",
        sha256 = "b2dcac3e49cbf75841d41ee1c53f1a91ffa78ba03dab8febb3153dbf76b2c5b2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b2dcac3e49cbf75841d41ee1c53f1a91ffa78ba03dab8febb3153dbf76b2c5b2",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/ca-certificates-2023.2.60-1.0.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "coreutils-0__9.1-7.fc37.x86_64",
        sha256 = "cd4f2bee79ba95edb4dd529a5a8488769c4538e91180495f1d81701ea1a5115d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cd4f2bee79ba95edb4dd529a5a8488769c4538e91180495f1d81701ea1a5115d",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-9.1-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-9.1-7.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-9.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-9.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "coreutils-common-0__9.1-7.fc37.x86_64",
        sha256 = "34e657305d9356b075c0fa58cdbfbb699bbf4b54c9a2c69534a1718faa8717d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/34e657305d9356b075c0fa58cdbfbb699bbf4b54c9a2c69534a1718faa8717d2",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-common-9.1-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-common-9.1-7.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-common-9.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-common-9.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "coreutils-single-0__9.1-7.fc37.x86_64",
        sha256 = "414bda840560471cb3d7380923ab00585ee78ca2db4b0d52155e9319a32151bc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/414bda840560471cb3d7380923ab00585ee78ca2db4b0d52155e9319a32151bc",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/c/coreutils-single-9.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cracklib-0__2.9.7-30.fc37.x86_64",
        sha256 = "3847abdc8ff973aeb0fb7e681bdf7c37b19cd49e5df17e8bf6bc35f34615c88f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3847abdc8ff973aeb0fb7e681bdf7c37b19cd49e5df17e8bf6bc35f34615c88f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cracklib-2.9.7-30.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cracklib-2.9.7-30.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cracklib-2.9.7-30.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cracklib-2.9.7-30.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "crypto-policies-0__20220815-1.gite4ed860.fc37.x86_64",
        sha256 = "486a11feeaad706c68b05de60a906cc57059454cbce436aeba45f88b84578c0c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/486a11feeaad706c68b05de60a906cc57059454cbce436aeba45f88b84578c0c",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/crypto-policies-20220815-1.gite4ed860.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "cryptsetup-devel-0__2.5.0-1.fc37.x86_64",
        sha256 = "dc7a6b834db44483d8d313fe7658a553ca721cebac99bf1db29a8db483e07964",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc7a6b834db44483d8d313fe7658a553ca721cebac99bf1db29a8db483e07964",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-devel-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cryptsetup-libs-0__2.5.0-1.fc37.x86_64",
        sha256 = "9a9c9f908326ce672180964c0dee6a387fefce9f4e49dacbca87f4aa8bf1e31f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9a9c9f908326ce672180964c0dee6a387fefce9f4e49dacbca87f4aa8bf1e31f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cryptsetup-libs-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cyrus-sasl-0__2.1.28-8.fc37.x86_64",
        sha256 = "1ede74bf11c2a8b3539a53176975f76531ceaf5bb525036b4740749e8a309484",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ede74bf11c2a8b3539a53176975f76531ceaf5bb525036b4740749e8a309484",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-2.1.28-8.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-2.1.28-8.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cyrus-sasl-gssapi-0__2.1.28-8.fc37.x86_64",
        sha256 = "b1dd9f0a836c47adf0628eef6dac3dee1059959e8f22e9c857d0a1f0ee3ff415",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b1dd9f0a836c47adf0628eef6dac3dee1059959e8f22e9c857d0a1f0ee3ff415",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-gssapi-2.1.28-8.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-gssapi-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-gssapi-2.1.28-8.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-gssapi-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-gssapi-2.1.28-8.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "cyrus-sasl-lib-0__2.1.28-8.fc37.x86_64",
        sha256 = "4e0e8656faf1f4f5227e4e40cdb4e662a1d78b19e74b90ba2f39f3cdf73e0083",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4e0e8656faf1f4f5227e4e40cdb4e662a1d78b19e74b90ba2f39f3cdf73e0083",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-lib-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-lib-2.1.28-8.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-lib-2.1.28-8.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/c/cyrus-sasl-lib-2.1.28-8.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-1__1.14.4-1.fc37.x86_64",
        sha256 = "2a9382a55160f297e86069e50b85f47df5546d0cacd3421bd1e79a69806b297a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2a9382a55160f297e86069e50b85f47df5546d0cacd3421bd1e79a69806b297a",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-1.14.4-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-broker-0__32-1.fc37.x86_64",
        sha256 = "e5bbfce30b88c0b4f06c4ad0c80645cfec9c23248d7c734d76607d8bc500c43f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e5bbfce30b88c0b4f06c4ad0c80645cfec9c23248d7c734d76607d8bc500c43f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/dbus-broker-32-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "dbus-common-1__1.14.4-1.fc37.x86_64",
        sha256 = "f2fe0d92a66d642759682ef3818d09519a499807bfb6b50b012b4178bf5e58f7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f2fe0d92a66d642759682ef3818d09519a499807bfb6b50b012b4178bf5e58f7",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/d/dbus-common-1.14.4-1.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "device-mapper-0__1.02.175-9.fc37.x86_64",
        sha256 = "c57831b8629e2e31b3c55d4f0064cd25a515d3eb1ac61fc6897ce07421a2e91b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c57831b8629e2e31b3c55d4f0064cd25a515d3eb1ac61fc6897ce07421a2e91b",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-1.02.175-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "device-mapper-libs-0__1.02.175-9.fc37.x86_64",
        sha256 = "7c0f72217eacc9b5caf553c17cb2428de242094dc7e0e1dbd0d21869d909c7d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c0f72217eacc9b5caf553c17cb2428de242094dc7e0e1dbd0d21869d909c7d2",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/d/device-mapper-libs-1.02.175-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "expat-0__2.5.0-1.fc37.x86_64",
        sha256 = "0e49c2393e5507bbaa16ededf0176e731e0196dd3230f6371d67be8b919e3429",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e49c2393e5507bbaa16ededf0176e731e0196dd3230f6371d67be8b919e3429",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/e/expat-2.5.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "fedora-gpg-keys-0__37-2.x86_64",
        sha256 = "47a0fdf0c8d0aecd3d4b2eee160affec5ba0d12b7ac6647b3f12fdef275e9738",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/47a0fdf0c8d0aecd3d4b2eee160affec5ba0d12b7ac6647b3f12fdef275e9738",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-gpg-keys-37-2.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-gpg-keys-37-2.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-gpg-keys-37-2.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-gpg-keys-37-2.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-common-0__37-15.x86_64",
        sha256 = "4a3013afe17b6e1413f8999c977ed4f8bd9c3d735f2f7bb066e7b021840934bb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a3013afe17b6e1413f8999c977ed4f8bd9c3d735f2f7bb066e7b021840934bb",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-common-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-container-0__37-15.x86_64",
        sha256 = "d4f3dccba997b2c72db48cb88fed1341d2aa1dfadd4b662c472da2f269fc4a85",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4f3dccba997b2c72db48cb88fed1341d2aa1dfadd4b662c472da2f269fc4a85",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-container-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-identity-cinnamon-0__37-15.x86_64",
        sha256 = "980a22600d18c680f6204a00ecaad4499fbeb104b997ef44bb4fdbf774226e53",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/980a22600d18c680f6204a00ecaad4499fbeb104b997ef44bb4fdbf774226e53",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-cinnamon-37-15.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-cinnamon-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-cinnamon-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-cinnamon-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-release-identity-snappy-0__37-15.x86_64",
        sha256 = "50db987aa99256532fb7826329f0dd468e879111c5baea6e528402d80f832b88",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/50db987aa99256532fb7826329f0dd468e879111c5baea6e528402d80f832b88",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-snappy-37-15.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-snappy-37-15.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-snappy-37-15.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-release-identity-snappy-37-15.noarch.rpm",
        ],
    )

    rpm(
        name = "fedora-repos-0__37-2.x86_64",
        sha256 = "f43a00322ae512135f695e9378eadcb3f8a8314bd4e290ea40c7c576621297f6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f43a00322ae512135f695e9378eadcb3f8a8314bd4e290ea40c7c576621297f6",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-repos-37-2.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-repos-37-2.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-repos-37-2.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/f/fedora-repos-37-2.noarch.rpm",
        ],
    )

    rpm(
        name = "filesystem-0__3.18-2.fc37.x86_64",
        sha256 = "1c28f722e7f3e48dba7ebf4f763ebebc6688b9e0fd58b55ba4fcd884c8180ef4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1c28f722e7f3e48dba7ebf4f763ebebc6688b9e0fd58b55ba4fcd884c8180ef4",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/f/filesystem-3.18-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "gawk-0__5.1.1-4.fc37.x86_64",
        sha256 = "6caea2f79e9fadf96e6cd55eac3f8625137b12f6a2ca75fb5e36b453dfe54edd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6caea2f79e9fadf96e6cd55eac3f8625137b12f6a2ca75fb5e36b453dfe54edd",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gawk-5.1.1-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gawk-5.1.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gawk-5.1.1-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gawk-5.1.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "gdbm-libs-1__1.23-2.fc37.x86_64",
        sha256 = "32ab362365afcf96144ba3e65c461cf6f8d495651d0c99fb4eeb970fc2b838e5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32ab362365afcf96144ba3e65c461cf6f8d495651d0c99fb4eeb970fc2b838e5",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gdbm-libs-1.23-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gdbm-libs-1.23-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gdbm-libs-1.23-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gdbm-libs-1.23-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glib2-0__2.74.1-2.fc37.x86_64",
        sha256 = "a61c7404e5d27bfbf0f1c0921b8cec3e6e30ad7342bf9c470800421e510aab77",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a61c7404e5d27bfbf0f1c0921b8cec3e6e30ad7342bf9c470800421e510aab77",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/glib2-2.74.1-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/glib2-2.74.1-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/glib2-2.74.1-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/glib2-2.74.1-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-0__2.36-9.fc37.x86_64",
        sha256 = "8c8463cd9f194f03ea1607670399e2fbf068857f566c43dd07d351228c25f187",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c8463cd9f194f03ea1607670399e2fbf068857f566c43dd07d351228c25f187",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-common-0__2.36-9.fc37.x86_64",
        sha256 = "4237c10e5edacc5d5a9ea88e9fc5fef37249d459b13d4a0715c7836374a8da7a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4237c10e5edacc5d5a9ea88e9fc5fef37249d459b13d4a0715c7836374a8da7a",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-common-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-langpack-az-0__2.36-9.fc37.x86_64",
        sha256 = "fe0a55b77cccc1b954fbcb2f3ba92aaf7a297cbbeb357f55c5c6d5aa90847a15",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe0a55b77cccc1b954fbcb2f3ba92aaf7a297cbbeb357f55c5c6d5aa90847a15",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-az-2.36-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-az-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-az-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-az-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "glibc-langpack-tcy-0__2.36-9.fc37.x86_64",
        sha256 = "19f0368ba03cc1d656d0087cede4f4681dcfae3549bc83be43c2c1c559053bb6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/19f0368ba03cc1d656d0087cede4f4681dcfae3549bc83be43c2c1c559053bb6",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-tcy-2.36-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-tcy-2.36-9.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-tcy-2.36-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/glibc-langpack-tcy-2.36-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "gmp-1__6.2.1-3.fc37.x86_64",
        sha256 = "42c8a66f1efcdffaf611e70395e16311f6c56ef795ee2a43c2a48c55eef77734",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42c8a66f1efcdffaf611e70395e16311f6c56ef795ee2a43c2a48c55eef77734",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gmp-6.2.1-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gmp-6.2.1-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gmp-6.2.1-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gmp-6.2.1-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "gnutls-0__3.7.8-3.fc37.x86_64",
        sha256 = "bf67dc97c68b287312baadaf02e80d88c42357cd8e89f8d090733ffd9e2fd5ad",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf67dc97c68b287312baadaf02e80d88c42357cd8e89f8d090733ffd9e2fd5ad",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/gnutls-3.7.8-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/g/gnutls-3.7.8-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/gnutls-3.7.8-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/g/gnutls-3.7.8-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "grep-0__3.7-4.fc37.x86_64",
        sha256 = "d997786e71f2c7b4a9ed1323b8684ec1802e49a866fb0c1b69101531440cb464",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d997786e71f2c7b4a9ed1323b8684ec1802e49a866fb0c1b69101531440cb464",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/grep-3.7-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "gzip-0__1.12-2.fc37.x86_64",
        sha256 = "3ef9e1b938dd19c5268004e370d90f8a8ae0dbc664715457a371ce900ee7736c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3ef9e1b938dd19c5268004e370d90f8a8ae0dbc664715457a371ce900ee7736c",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gzip-1.12-2.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gzip-1.12-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gzip-1.12-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gzip-1.12-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/g/gzip-1.12-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "json-c-0__0.16-3.fc37.x86_64",
        sha256 = "e7c83a9058c7e7e05e4c7ba97a363414eb973343ea8f00a1140fbdafe6ca67e2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7c83a9058c7e7e05e4c7ba97a363414eb973343ea8f00a1140fbdafe6ca67e2",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/j/json-c-0.16-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "keyutils-libs-0__1.6.1-5.fc37.x86_64",
        sha256 = "e3fd19c3020e55d80b8a24edb68506d2adbb07b2db29eecbde91facae1cca59d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3fd19c3020e55d80b8a24edb68506d2adbb07b2db29eecbde91facae1cca59d",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/keyutils-libs-1.6.1-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/keyutils-libs-1.6.1-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/keyutils-libs-1.6.1-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/keyutils-libs-1.6.1-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "kmod-libs-0__30-2.fc37.x86_64",
        sha256 = "73a1a0f041819c1d50501a699945f0121a3b6e1f54df40cd0bf8f94b1b261ef5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/73a1a0f041819c1d50501a699945f0121a3b6e1f54df40cd0bf8f94b1b261ef5",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/k/kmod-libs-30-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "krb5-libs-0__1.19.2-13.fc37.x86_64",
        sha256 = "5f2ffaa4084cb8918d3990ef352dbfdd9ac28d30c2ed2693c1011641199bb369",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5f2ffaa4084cb8918d3990ef352dbfdd9ac28d30c2ed2693c1011641199bb369",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/k/krb5-libs-1.19.2-13.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/k/krb5-libs-1.19.2-13.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/k/krb5-libs-1.19.2-13.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/k/krb5-libs-1.19.2-13.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libacl-0__2.3.1-4.fc37.x86_64",
        sha256 = "15224cb92199b8011fe47dc12e0bbcdbee0c93e0f29553b3b07ae41768b48ce3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15224cb92199b8011fe47dc12e0bbcdbee0c93e0f29553b3b07ae41768b48ce3",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libacl-2.3.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libargon2-0__20190702-1.fc37.x86_64",
        sha256 = "bf280bf9e59891bfcb4a987d5df22d6a6d9f60589dd00b790b5a3047a727a40b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf280bf9e59891bfcb4a987d5df22d6a6d9f60589dd00b790b5a3047a727a40b",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libargon2-20190702-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libattr-0__2.5.1-5.fc37.x86_64",
        sha256 = "3a423be562953538eaa0d1e78ef35890396cdf1ad89561c619aa72d3a59bfb82",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3a423be562953538eaa0d1e78ef35890396cdf1ad89561c619aa72d3a59bfb82",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libattr-2.5.1-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libblkid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b0388d1a529bf6b54ca648e91529b1e7790e6aaa42e0ac2b7be6640e4f24a21d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0388d1a529bf6b54ca648e91529b1e7790e6aaa42e0ac2b7be6640e4f24a21d",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libblkid-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcap-0__2.48-5.fc37.x86_64",
        sha256 = "aa22373907b6ff9fa3d2f7d9e33a9bdefc9ac50486f2dac5251ac4e206a8a61d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aa22373907b6ff9fa3d2f7d9e33a9bdefc9ac50486f2dac5251ac4e206a8a61d",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-2.48-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcap-ng-0__0.8.3-3.fc37.x86_64",
        sha256 = "bcca8a17ae16f9f1c8664f9f54e8f2178f028821f6802ebf33cdcd2d4289bf7f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bcca8a17ae16f9f1c8664f9f54e8f2178f028821f6802ebf33cdcd2d4289bf7f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcap-ng-0.8.3-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcom_err-0__1.46.5-3.fc37.x86_64",
        sha256 = "e98643b3299e5a5b9b1e85a0763b567035f1d83164b3b9a4629fd23467667464",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e98643b3299e5a5b9b1e85a0763b567035f1d83164b3b9a4629fd23467667464",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcom_err-1.46.5-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcom_err-1.46.5-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcom_err-1.46.5-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libcom_err-1.46.5-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libcurl-minimal-0__7.85.0-5.fc37.x86_64",
        sha256 = "07ff529788102d54d33c2c0f3dd423ee76647bb6f39eb55583939900a68fb819",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/07ff529788102d54d33c2c0f3dd423ee76647bb6f39eb55583939900a68fb819",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libcurl-minimal-7.85.0-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libcurl-minimal-7.85.0-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libcurl-minimal-7.85.0-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libcurl-minimal-7.85.0-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libdb-0__5.3.28-53.fc37.x86_64",
        sha256 = "e89a4a620d5531f30b895694134a982fa37615b3f61c59a21ede6e64a096c5cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e89a4a620d5531f30b895694134a982fa37615b3f61c59a21ede6e64a096c5cd",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libdb-5.3.28-53.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libdb-5.3.28-53.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libdb-5.3.28-53.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libdb-5.3.28-53.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libeconf-0__0.4.0-4.fc37.x86_64",
        sha256 = "f0cc1addee779f09aade289e3be4e9bd103a274a6bdf11f8331878686f432653",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f0cc1addee779f09aade289e3be4e9bd103a274a6bdf11f8331878686f432653",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libeconf-0.4.0-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libevent-0__2.1.12-7.fc37.x86_64",
        sha256 = "eac9405b6177c4778d772b61ef03a5cd571e2ce6ea337929a1e8a10e80422ba7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eac9405b6177c4778d772b61ef03a5cd571e2ce6ea337929a1e8a10e80422ba7",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libevent-2.1.12-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libevent-2.1.12-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libevent-2.1.12-7.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libevent-2.1.12-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libfdisk-0__2.38.1-1.fc37.x86_64",
        sha256 = "7a4bd1f4975a52fc201c9bc978f155dcb97212cb970210525d903b03644a713d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7a4bd1f4975a52fc201c9bc978f155dcb97212cb970210525d903b03644a713d",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libfdisk-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libffi-0__3.4.2-9.fc37.x86_64",
        sha256 = "dc5f7ca1ce86cd9380525b383624bdc1afa52d98db624cfdece7b08086f829d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc5f7ca1ce86cd9380525b383624bdc1afa52d98db624cfdece7b08086f829d6",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libffi-3.4.2-9.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libgcc-0__12.2.1-4.fc37.x86_64",
        sha256 = "25299b673e7488f538c6d0433ea7fe0ffc8311e41dd7115b5985145e493e4b05",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/25299b673e7488f538c6d0433ea7fe0ffc8311e41dd7115b5985145e493e4b05",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgcc-12.2.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libgcrypt-0__1.10.1-4.fc37.x86_64",
        sha256 = "ca802ad5d10b2728ba10bf98bb16796585d69ec775f5452b3a43718e07c4667a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ca802ad5d10b2728ba10bf98bb16796585d69ec775f5452b3a43718e07c4667a",
            "https://mirror.dogado.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libgcrypt-1.10.1-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libgcrypt-1.10.1-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libgcrypt-1.10.1-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libgcrypt-1.10.1-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libgpg-error-0__1.46-1.fc37.x86_64",
        sha256 = "bfa65a9946b2547110994855d168e4434313ad26280cb935c19bb88d2af283d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bfa65a9946b2547110994855d168e4434313ad26280cb935c19bb88d2af283d2",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgpg-error-1.46-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgpg-error-1.46-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgpg-error-1.46-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libgpg-error-1.46-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libidn2-0__2.3.4-1.fc37.x86_64",
        sha256 = "e32e2ab71cfb0bedb84611251987db7acdf665917864be335d0786ea6bbd02b4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e32e2ab71cfb0bedb84611251987db7acdf665917864be335d0786ea6bbd02b4",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libidn2-2.3.4-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libidn2-2.3.4-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libidn2-2.3.4-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libidn2-2.3.4-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libmount-0__2.38.1-1.fc37.x86_64",
        sha256 = "50c304faa94d7959e5cbc0642b3c77539ad000042e6617ea5da4789c8105496f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/50c304faa94d7959e5cbc0642b3c77539ad000042e6617ea5da4789c8105496f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libmount-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libnghttp2-0__1.51.0-1.fc37.x86_64",
        sha256 = "42fbaaacbeb241755d8448dd5672bbbcc48cbe9548c095ce0efef4140bc12520",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42fbaaacbeb241755d8448dd5672bbbcc48cbe9548c095ce0efef4140bc12520",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libnghttp2-1.51.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libnghttp2-1.51.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libnghttp2-1.51.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libnghttp2-1.51.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libnl3-0__3.7.0-2.fc37.x86_64",
        sha256 = "4543c991e6f536468d9d47527a201b58b9bc049364a6bdfe15a2f910a02e68f6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4543c991e6f536468d9d47527a201b58b9bc049364a6bdfe15a2f910a02e68f6",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnl3-3.7.0-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnl3-3.7.0-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnl3-3.7.0-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnl3-3.7.0-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libnsl2-0__2.0.0-4.fc37.x86_64",
        sha256 = "a1e9428515b0df1c2a423ad3c35bcdf93333172fe346169bb3018a882e27be5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a1e9428515b0df1c2a423ad3c35bcdf93333172fe346169bb3018a882e27be5f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnsl2-2.0.0-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnsl2-2.0.0-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnsl2-2.0.0-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libnsl2-2.0.0-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libpkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "ecd52fd3f3065606ba5164249b29c837cbd172643d13a00a1a72fc657b115af7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ecd52fd3f3065606ba5164249b29c837cbd172643d13a00a1a72fc657b115af7",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libpkgconf-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libpwquality-0__1.4.5-3.fc37.x86_64",
        sha256 = "a9019a471496fdada529757331ec004397db7a0c4347531bd639c127bbaf8300",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9019a471496fdada529757331ec004397db7a0c4347531bd639c127bbaf8300",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/l/libpwquality-1.4.5-3.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libpwquality-1.4.5-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libpwquality-1.4.5-3.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libpwquality-1.4.5-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libpwquality-1.4.5-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libseccomp-0__2.5.3-3.fc37.x86_64",
        sha256 = "017877a97c8222fc7eca7fab77600a3a1fcdec92f9dd39d8df6e64726909fcbe",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/017877a97c8222fc7eca7fab77600a3a1fcdec92f9dd39d8df6e64726909fcbe",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libseccomp-2.5.3-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libselinux-0__3.4-5.fc37.x86_64",
        sha256 = "2a5b4e2e1dd388c3c13d79af971fb8efd522f5c2ba8d257875f02c16b4858214",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2a5b4e2e1dd388c3c13d79af971fb8efd522f5c2ba8d257875f02c16b4858214",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libselinux-3.4-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsemanage-0__3.4-5.fc37.x86_64",
        sha256 = "66305d7a3ca92165f1c17e14cc29ea70280fa1c1fd3bf223b5b1d4f7d1ce0dd8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66305d7a3ca92165f1c17e14cc29ea70280fa1c1fd3bf223b5b1d4f7d1ce0dd8",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsemanage-3.4-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsepol-0__3.4-3.fc37.x86_64",
        sha256 = "97e918bc5b11c8abfec9343e1b0bd88087b792e85c604427d5cdc32733f70b3f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97e918bc5b11c8abfec9343e1b0bd88087b792e85c604427d5cdc32733f70b3f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsepol-3.4-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsigsegv-0__2.14-3.fc37.x86_64",
        sha256 = "0f038b70d155dae3df4824776c5a135f02c423c688b9486d4f84eb6a16a90494",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f038b70d155dae3df4824776c5a135f02c423c688b9486d4f84eb6a16a90494",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsigsegv-2.14-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsigsegv-2.14-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsigsegv-2.14-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsigsegv-2.14-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libsmartcols-0__2.38.1-1.fc37.x86_64",
        sha256 = "93246c002aefec27bb398aa3397ae555bcc3035b10aebb4937c4bea9268bacf1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93246c002aefec27bb398aa3397ae555bcc3035b10aebb4937c4bea9268bacf1",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libsmartcols-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libssh-0__0.10.4-2.fc37.x86_64",
        sha256 = "cdee8c9676d686a0df90d27b4863f15e871dc58363eb2f11f5e69fe3e9a23c85",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cdee8c9676d686a0df90d27b4863f15e871dc58363eb2f11f5e69fe3e9a23c85",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-0.10.4-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-0.10.4-2.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-0.10.4-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-0.10.4-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libssh-config-0__0.10.4-2.fc37.x86_64",
        sha256 = "d17d16ca2e2a42035778094bca077ba675e440911c5546f99a274278eb32e0d7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d17d16ca2e2a42035778094bca077ba675e440911c5546f99a274278eb32e0d7",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-config-0.10.4-2.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-config-0.10.4-2.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-config-0.10.4-2.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libssh-config-0.10.4-2.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "libssh2-0__1.10.0-5.fc37.x86_64",
        sha256 = "8a6e1ce9b08e054746ef1ee0eb51a70bc06ee1bea3ae24bf5942a191ea12ab3a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8a6e1ce9b08e054746ef1ee0eb51a70bc06ee1bea3ae24bf5942a191ea12ab3a",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libssh2-1.10.0-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libssh2-1.10.0-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libssh2-1.10.0-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libssh2-1.10.0-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libtasn1-0__4.19.0-1.fc37.x86_64",
        sha256 = "35b51a0796af6930b2a8a511df8c51938006cfcfdf74ddfe6482eb9febd87dfa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/35b51a0796af6930b2a8a511df8c51938006cfcfdf74ddfe6482eb9febd87dfa",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libtasn1-4.19.0-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libtirpc-0__1.3.3-0.fc37.x86_64",
        sha256 = "76dcdfd95452e176f64d6008d114e9415cd8384c5c0d3300fe644c137b6917fa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76dcdfd95452e176f64d6008d114e9415cd8384c5c0d3300fe644c137b6917fa",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libtirpc-1.3.3-0.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libtirpc-1.3.3-0.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libtirpc-1.3.3-0.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libtirpc-1.3.3-0.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libunistring-0__1.0-2.fc37.x86_64",
        sha256 = "acb031577655bba5a41c1fb0ec954bb84e207f9e2d08b2cdb3d4e2b7806b0670",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/acb031577655bba5a41c1fb0ec954bb84e207f9e2d08b2cdb3d4e2b7806b0670",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libunistring-1.0-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libunistring-1.0-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libunistring-1.0-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libunistring-1.0-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libutempter-0__1.2.1-7.fc37.x86_64",
        sha256 = "8fc30b0742e939954d6aebd45364dcd1dbb8b9c85e75c799301c3507e22ea56a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fc30b0742e939954d6aebd45364dcd1dbb8b9c85e75c799301c3507e22ea56a",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libutempter-1.2.1-7.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libutempter-1.2.1-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libutempter-1.2.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libutempter-1.2.1-7.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libutempter-1.2.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libuuid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b054577d98aa9615fe459abec31be46b19ad72e0da620d8d251b4449a6db020d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b054577d98aa9615fe459abec31be46b19ad72e0da620d8d251b4449a6db020d",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libuuid-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libverto-0__0.3.2-4.fc37.x86_64",
        sha256 = "ca47b52e1ecd8a2ac6eda368d985390816fbb447f43135ec0ba105165997817f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ca47b52e1ecd8a2ac6eda368d985390816fbb447f43135ec0ba105165997817f",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libverto-0.3.2-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libverto-0.3.2-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libverto-0.3.2-4.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libverto-0.3.2-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libvirt-devel-0__8.6.0-5.fc37.x86_64",
        sha256 = "3cf11a2799e69f36ee6ee9015549b2a956a9f7f434a015a42ba06835cfad5b83",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3cf11a2799e69f36ee6ee9015549b2a956a9f7f434a015a42ba06835cfad5b83",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-devel-8.6.0-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-devel-8.6.0-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-devel-8.6.0-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-devel-8.6.0-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libvirt-libs-0__8.6.0-5.fc37.x86_64",
        sha256 = "035260aa3ad6e33ace0cfcf075d7f946bef0406484d5fd88a40209c160ca27e5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/035260aa3ad6e33ace0cfcf075d7f946bef0406484d5fd88a40209c160ca27e5",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-libs-8.6.0-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-libs-8.6.0-5.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-libs-8.6.0-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libvirt-libs-8.6.0-5.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libwsman1-0__2.7.1-7.fc37.x86_64",
        sha256 = "41d9f13f5a7020a70f565e326fc5dd9167be30f3298ce5360081028cf2efbcab",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/41d9f13f5a7020a70f565e326fc5dd9167be30f3298ce5360081028cf2efbcab",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libwsman1-2.7.1-7.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libwsman1-2.7.1-7.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libwsman1-2.7.1-7.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libwsman1-2.7.1-7.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libxcrypt-0__4.4.33-4.fc37.x86_64",
        sha256 = "547b9cffb0211abc4445d159e944f4fb59606b2eddfc14813b8c068859294ba6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/547b9cffb0211abc4445d159e944f4fb59606b2eddfc14813b8c068859294ba6",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxcrypt-4.4.33-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libxml2-0__2.10.3-2.fc37.x86_64",
        sha256 = "105e8b221029cc4595682cd837dd80c1124685477efbec280fef2e2bb4974d2d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/105e8b221029cc4595682cd837dd80c1124685477efbec280fef2e2bb4974d2d",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxml2-2.10.3-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxml2-2.10.3-2.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxml2-2.10.3-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/libxml2-2.10.3-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "libzstd-0__1.5.2-3.fc37.x86_64",
        sha256 = "c105f54738fba9793dad9b6ab2e88b4ae05cc47b9ea062cd7215b69d11ce0e1c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c105f54738fba9793dad9b6ab2e88b4ae05cc47b9ea062cd7215b69d11ce0e1c",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/l/libzstd-1.5.2-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "lz4-libs-0__1.9.4-1.fc37.x86_64",
        sha256 = "f39b8b018fcb2b55477cdbfa4af7c9db9b660c85000a4a42e880b1a951efbe5a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f39b8b018fcb2b55477cdbfa4af7c9db9b660c85000a4a42e880b1a951efbe5a",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/l/lz4-libs-1.9.4-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "mpfr-0__4.1.0-10.fc37.x86_64",
        sha256 = "3be8cf104424fb5e148846a1df4a9c193527f55ee866bff0963e788450483566",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3be8cf104424fb5e148846a1df4a9c193527f55ee866bff0963e788450483566",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/m/mpfr-4.1.0-10.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/m/mpfr-4.1.0-10.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/m/mpfr-4.1.0-10.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/m/mpfr-4.1.0-10.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "ncurses-base-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "000164a9a82458fbb69b3433801dcc0d0e2437e21d7f7d4fd45f63a42a0bc26f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/000164a9a82458fbb69b3433801dcc0d0e2437e21d7f7d4fd45f63a42a0bc26f",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-base-6.3-4.20220501.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "ncurses-libs-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "75e51eebcd3fe150b421ec5b1c9a6e918caa5b3c0f243f2b70d445fd434488bb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/75e51eebcd3fe150b421ec5b1c9a6e918caa5b3c0f243f2b70d445fd434488bb",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/n/ncurses-libs-6.3-4.20220501.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "nettle-0__3.8-2.fc37.x86_64",
        sha256 = "8fe2d98578b0c4454536faacbaafd66d1754b8439bb6332d7576a741f4c72208",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fe2d98578b0c4454536faacbaafd66d1754b8439bb6332d7576a741f4c72208",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/nettle-3.8-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/nettle-3.8-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/nettle-3.8-2.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/nettle-3.8-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "numactl-libs-0__2.0.14-6.fc37.x86_64",
        sha256 = "8f2e423d8f64f3abf33f8660df718d69f785a673a57eb188258a9f79af8f678f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8f2e423d8f64f3abf33f8660df718d69f785a673a57eb188258a9f79af8f678f",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/numactl-libs-2.0.14-6.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/numactl-libs-2.0.14-6.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/numactl-libs-2.0.14-6.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/numactl-libs-2.0.14-6.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/n/numactl-libs-2.0.14-6.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "openldap-0__2.6.3-1.fc37.x86_64",
        sha256 = "c42642046bf068a2d0cc32f38cdc56d0c92af48eb131fc8dce55a997142f2e88",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c42642046bf068a2d0cc32f38cdc56d0c92af48eb131fc8dce55a997142f2e88",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openldap-2.6.3-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openldap-2.6.3-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openldap-2.6.3-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openldap-2.6.3-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "openssl-libs-1__3.0.5-3.fc37.x86_64",
        sha256 = "76fdbe6d7d4cd898d73da5a36fcbbfd2330a4855a2dc3e023480e23caf4eac8c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76fdbe6d7d4cd898d73da5a36fcbbfd2330a4855a2dc3e023480e23caf4eac8c",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/o/openssl-libs-3.0.5-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "p11-kit-0__0.24.1-3.fc37.x86_64",
        sha256 = "4dad6ac54eb7708cbfc8522d372f2a196cf711e97e279cbddba8cc8b92970dd7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4dad6ac54eb7708cbfc8522d372f2a196cf711e97e279cbddba8cc8b92970dd7",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-0.24.1-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "p11-kit-trust-0__0.24.1-3.fc37.x86_64",
        sha256 = "0fd85eb1ce27615fea745721b18648b4a4585ad4b11a482c1b77fc1785cd5194",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0fd85eb1ce27615fea745721b18648b4a4585ad4b11a482c1b77fc1785cd5194",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/p11-kit-trust-0.24.1-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pam-0__1.5.2-14.fc37.x86_64",
        sha256 = "a66ee1c9f9155c97e77cbd18658ce5129638f7d6e208c01c172c4dd1dfdbbe6d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a66ee1c9f9155c97e77cbd18658ce5129638f7d6e208c01c172c4dd1dfdbbe6d",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-1.5.2-14.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-1.5.2-14.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-1.5.2-14.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/releases/development/37/Everything/x86_64/os/Packages/p/pam-1.5.2-14.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pam-libs-0__1.5.2-14.fc37.x86_64",
        sha256 = "ee34422adc6451da744bd16a8cd66c9912a822c4e55227c23ff56960c32980f5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee34422adc6451da744bd16a8cd66c9912a822c4e55227c23ff56960c32980f5",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pam-libs-1.5.2-14.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre-0__8.45-1.fc37.2.x86_64",
        sha256 = "86a648e3b88f581b15ca2eda6b441be7c5c3810a9eae25ca940c767029e4e923",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86a648e3b88f581b15ca2eda6b441be7c5c3810a9eae25ca940c767029e4e923",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre-8.45-1.fc37.2.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre2-0__10.40-1.fc37.1.x86_64",
        sha256 = "422de947ec1a7aafcd212a51e64257b64d5b0a02808104a33e7c3cd9ef629148",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/422de947ec1a7aafcd212a51e64257b64d5b0a02808104a33e7c3cd9ef629148",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-10.40-1.fc37.1.x86_64.rpm",
        ],
    )

    rpm(
        name = "pcre2-syntax-0__10.40-1.fc37.1.x86_64",
        sha256 = "585f339942a0bf4b0eab638ddf825544793485cbcb9f1eaee079b9956d90aafa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/585f339942a0bf4b0eab638ddf825544793485cbcb9f1eaee079b9956d90aafa",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pcre2-syntax-10.40-1.fc37.1.noarch.rpm",
        ],
    )

    rpm(
        name = "pkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "778018594ab5bddc4432e53985b80e6c5a1a1ec1700d38b438848d485f5b357c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/778018594ab5bddc4432e53985b80e6c5a1a1ec1700d38b438848d485f5b357c",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "pkgconf-m4-0__1.8.0-3.fc37.x86_64",
        sha256 = "dd0356475d0b9106b5a2d577db359aa0290fe6dd9eacea1b6e0cab816ff33566",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dd0356475d0b9106b5a2d577db359aa0290fe6dd9eacea1b6e0cab816ff33566",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/aarch64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-m4-1.8.0-3.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "pkgconf-pkg-config-0__1.8.0-3.fc37.x86_64",
        sha256 = "d238b12c750b58ceebc80e25c2074bd929d3f232c1390677f33a94fdadb68f6a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d238b12c750b58ceebc80e25c2074bd929d3f232c1390677f33a94fdadb68f6a",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/pkgconf-pkg-config-1.8.0-3.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "popt-0__1.19-1.fc37.x86_64",
        sha256 = "e3c9a6a1611d967fbff4321b5b1ae54377fed22454298859108138c1f64b0c63",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3c9a6a1611d967fbff4321b5b1ae54377fed22454298859108138c1f64b0c63",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/popt-1.19-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/popt-1.19-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/popt-1.19-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/p/popt-1.19-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "readline-0__8.2-2.fc37.x86_64",
        sha256 = "0663e23dc42a7ce84f60f5f3154ba640460a0e5b7158459abf9d5d0986d69d06",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0663e23dc42a7ce84f60f5f3154ba640460a0e5b7158459abf9d5d0986d69d06",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/r/readline-8.2-2.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/r/readline-8.2-2.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/r/readline-8.2-2.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/r/readline-8.2-2.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "sed-0__4.8-11.fc37.x86_64",
        sha256 = "231e782077862f4abecf025aa254a9c391a950490ae856261dcfd229863ac80f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/231e782077862f4abecf025aa254a9c391a950490ae856261dcfd229863ac80f",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/sed-4.8-11.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "setup-0__2.14.1-2.fc37.x86_64",
        sha256 = "15d72b2a44f403b3a7ee9138820a8ce7584f954aeafbb43b1251621bca26f785",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15d72b2a44f403b3a7ee9138820a8ce7584f954aeafbb43b1251621bca26f785",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/s/setup-2.14.1-2.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "shadow-utils-2__4.12.3-4.fc37.x86_64",
        sha256 = "8394db7e5385d64c90876a0ee9274c3f53ae91f172c051524a24a30963e18fb8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8394db7e5385d64c90876a0ee9274c3f53ae91f172c051524a24a30963e18fb8",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/shadow-utils-4.12.3-4.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-0__251.10-588.fc37.x86_64",
        sha256 = "f8d45c2c70b06ba15eb2601bc1bc83d030b5df059525964456b98ef45c2f146e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f8d45c2c70b06ba15eb2601bc1bc83d030b5df059525964456b98ef45c2f146e",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-libs-0__251.10-588.fc37.x86_64",
        sha256 = "b383c219341e141d003ae0722fcea5e6949cf1af41adc64d290292940b239bca",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b383c219341e141d003ae0722fcea5e6949cf1af41adc64d290292940b239bca",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-libs-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "systemd-pam-0__251.10-588.fc37.x86_64",
        sha256 = "4d30204cbb169f9fdbbc872a71d6ddea95f442fa3489fb075fe1d06ff9222c48",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d30204cbb169f9fdbbc872a71d6ddea95f442fa3489fb075fe1d06ff9222c48",
            "https://download-ib01.fedoraproject.org/pub/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/s/systemd-pam-251.10-588.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "tzdata-0__2022g-1.fc37.x86_64",
        sha256 = "7ff35c66b3478103fbf3941e933e25f60e41f2b0bfd07d43666b40721211c3bb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7ff35c66b3478103fbf3941e933e25f60e41f2b0bfd07d43666b40721211c3bb",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/t/tzdata-2022g-1.fc37.noarch.rpm",
        ],
    )

    rpm(
        name = "util-linux-0__2.38.1-1.fc37.x86_64",
        sha256 = "23f052850cd509743fae6089181a124ee65c2783d6d15f61ffbae1272f5f67ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/23f052850cd509743fae6089181a124ee65c2783d6d15f61ffbae1272f5f67ef",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "util-linux-core-0__2.38.1-1.fc37.x86_64",
        sha256 = "f87ad8fc18f4da254966cc6f99b533dc8125e1ec0eaefd5f89a6b6398cb13a34",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f87ad8fc18f4da254966cc6f99b533dc8125e1ec0eaefd5f89a6b6398cb13a34",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/u/util-linux-core-2.38.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "xz-libs-0__5.4.1-1.fc37.x86_64",
        sha256 = "8c06eef8dd28d6dc1406e65e4eb8ee3db359cf6624729be4e426f6b01c4117fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c06eef8dd28d6dc1406e65e4eb8ee3db359cf6624729be4e426f6b01c4117fd",
            "https://mirror.dogado.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://ftp.halifax.rwth-aachen.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/updates/37/Everything/x86_64/Packages/x/xz-libs-5.4.1-1.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "yajl-0__2.1.0-19.fc37.x86_64",
        sha256 = "b0ca9c6ed5935cde0094694127c13b99a441207eb084f44fb3aa093669c9957c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0ca9c6ed5935cde0094694127c13b99a441207eb084f44fb3aa093669c9957c",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/y/yajl-2.1.0-19.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/y/yajl-2.1.0-19.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/y/yajl-2.1.0-19.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/y/yajl-2.1.0-19.fc37.x86_64.rpm",
        ],
    )

    rpm(
        name = "zlib-0__1.2.12-5.fc37.x86_64",
        sha256 = "7b0eda1ad9e9a06e61d9fe41e5e4e0fbdc8427bc252f06a7d29cd7ba81a71a70",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7b0eda1ad9e9a06e61d9fe41e5e4e0fbdc8427bc252f06a7d29cd7ba81a71a70",
            "https://ftp.plusline.net/fedora/linux/releases/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://mirror.23m.com/fedora/linux/releases/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://ftp.fau.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
            "https://mirror.netzwerge.de/fedora/linux/releases/37/Everything/x86_64/os/Packages/z/zlib-1.2.12-5.fc37.x86_64.rpm",
        ],
    )
