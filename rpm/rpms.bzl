""" This file is used to pin / load external RPMs that are used by the project. """

load("@bazeldnf//:deps.bzl", "rpm")

def rpms():
    """ Provides a list of RPMs that are used by the project. """
    rpm(
        name = "SDL2-0__2.26.3-1.fc37.x86_64",
        sha256 = "d4e46f65f06a8c29392af58fac78fe4793badcda538abec7f379d2fd74c898dd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4e46f65f06a8c29392af58fac78fe4793badcda538abec7f379d2fd74c898dd",
        ],
    )

    rpm(
        name = "SDL2_image-0__2.6.3-1.fc37.x86_64",
        sha256 = "7d47cfa4e23b019f96299f6eb896848d8e3efe752bc6fcfca2081e34c53b8f3b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7d47cfa4e23b019f96299f6eb896848d8e3efe752bc6fcfca2081e34c53b8f3b",
        ],
    )

    rpm(
        name = "adwaita-cursor-theme-0__43-1.fc37.x86_64",
        sha256 = "0e183c26fbfc38276474bfaa55ac6ec8ce5f06a91c840bd4a4c5137f8f75cabe",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e183c26fbfc38276474bfaa55ac6ec8ce5f06a91c840bd4a4c5137f8f75cabe",
        ],
    )

    rpm(
        name = "adwaita-icon-theme-0__43-1.fc37.x86_64",
        sha256 = "ae157313e647a1561a55a0ef5b5836cf15912c43f921fd6bc0835a7b02737bba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ae157313e647a1561a55a0ef5b5836cf15912c43f921fd6bc0835a7b02737bba",
        ],
    )

    rpm(
        name = "alsa-lib-0__1.2.8-2.fc37.x86_64",
        sha256 = "ad653de273d11bb6459e334dec4ef7c5b094a7b2ef53b54decb6b93e8eab7cb5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ad653de273d11bb6459e334dec4ef7c5b094a7b2ef53b54decb6b93e8eab7cb5",
        ],
    )

    rpm(
        name = "alternatives-0__1.22-1.fc37.x86_64",
        sha256 = "cf161bb87d597d013444180f4aa26c38e4e85b30f998ad77f2adc25143314055",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cf161bb87d597d013444180f4aa26c38e4e85b30f998ad77f2adc25143314055",
        ],
    )

    rpm(
        name = "at-spi2-atk-0__2.38.0-5.fc37.x86_64",
        sha256 = "c328259b305cb69ed5ad83188cd3a29d31ac7fc9b79c14ac37c1637418c043dc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c328259b305cb69ed5ad83188cd3a29d31ac7fc9b79c14ac37c1637418c043dc",
        ],
    )

    rpm(
        name = "at-spi2-core-0__2.44.1-2.fc37.x86_64",
        sha256 = "0d197507fe9f985b520341e22455664c52755adde5e6af0ba1b5c44716511a5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0d197507fe9f985b520341e22455664c52755adde5e6af0ba1b5c44716511a5f",
        ],
    )

    rpm(
        name = "atk-0__2.38.0-2.fc37.x86_64",
        sha256 = "8e64b25736a8875fe2f0832cf558d4a4c50065889c9eec89e6ca8d421a2940d3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8e64b25736a8875fe2f0832cf558d4a4c50065889c9eec89e6ca8d421a2940d3",
        ],
    )

    rpm(
        name = "attr-0__2.5.1-5.fc37.x86_64",
        sha256 = "c5efe6b45e60fe8ca0fc84796865463355da8ffa66f43c277c4dab5934c395f0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c5efe6b45e60fe8ca0fc84796865463355da8ffa66f43c277c4dab5934c395f0",
        ],
    )

    rpm(
        name = "audit-libs-0__3.1-2.fc37.x86_64",
        sha256 = "c58f2c9982f16cc492bba42e7618bbc932a0521b27179e17f6828fcf28c266aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c58f2c9982f16cc492bba42e7618bbc932a0521b27179e17f6828fcf28c266aa",
        ],
    )

    rpm(
        name = "authselect-0__1.4.2-1.fc37.x86_64",
        sha256 = "c356d05e80f2b57ea2598b45b168fff6da189038e3f3ef0305dd90cfdd2a045f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c356d05e80f2b57ea2598b45b168fff6da189038e3f3ef0305dd90cfdd2a045f",
        ],
    )

    rpm(
        name = "authselect-libs-0__1.4.2-1.fc37.x86_64",
        sha256 = "275c282a240a3b7225e98b540a91af3419a9fa527623c5f152c48f8209779146",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/275c282a240a3b7225e98b540a91af3419a9fa527623c5f152c48f8209779146",
        ],
    )

    rpm(
        name = "avahi-libs-0__0.8-18.fc37.x86_64",
        sha256 = "a5e954bc4aab9a5d44fbeec0379313fc89abfde2dab2bb5594a7ea6e6e6106c5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a5e954bc4aab9a5d44fbeec0379313fc89abfde2dab2bb5594a7ea6e6e6106c5",
        ],
    )

    rpm(
        name = "basesystem-0__11-14.fc37.x86_64",
        sha256 = "38d1877d647bb5f4047d22982a51899c95bdfea1d7b2debbff37c66f0fc0ed44",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/38d1877d647bb5f4047d22982a51899c95bdfea1d7b2debbff37c66f0fc0ed44",
        ],
    )

    rpm(
        name = "bash-0__5.2.15-1.fc37.x86_64",
        sha256 = "e50ddbdb35ecec1a9bf4e19fd87c6216382be313c3b671704d444053a1cfd183",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e50ddbdb35ecec1a9bf4e19fd87c6216382be313c3b671704d444053a1cfd183",
        ],
    )

    rpm(
        name = "boost-iostreams-0__1.78.0-9.fc37.x86_64",
        sha256 = "af80d2358cd9636316e8c5c2ea5d382306c89b38081e1940cc16192dd54b7e2a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af80d2358cd9636316e8c5c2ea5d382306c89b38081e1940cc16192dd54b7e2a",
        ],
    )

    rpm(
        name = "boost-system-0__1.78.0-9.fc37.x86_64",
        sha256 = "b50218f5f53a98dbaccb0fbb58f0d3086bad8b8b62a9ed19a2f5fcdb6d0b4d91",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b50218f5f53a98dbaccb0fbb58f0d3086bad8b8b62a9ed19a2f5fcdb6d0b4d91",
        ],
    )

    rpm(
        name = "boost-thread-0__1.78.0-9.fc37.x86_64",
        sha256 = "9568d5e462be0a6dcb97580f97432e60eacd2b076df93ed298db3688e5ed5df0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9568d5e462be0a6dcb97580f97432e60eacd2b076df93ed298db3688e5ed5df0",
        ],
    )

    rpm(
        name = "brlapi-0__0.8.4-7.fc37.x86_64",
        sha256 = "1a9905e955a3e1556a97ff9a99d630491a7095f33c232a7a3be439b4f23fc23f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1a9905e955a3e1556a97ff9a99d630491a7095f33c232a7a3be439b4f23fc23f",
        ],
    )

    rpm(
        name = "bzip2-0__1.0.8-12.fc37.x86_64",
        sha256 = "da9c23df152e8e0ab0aa1cbf9dae111575ba1d796a799557a435902acd947dc3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/da9c23df152e8e0ab0aa1cbf9dae111575ba1d796a799557a435902acd947dc3",
        ],
    )

    rpm(
        name = "bzip2-libs-0__1.0.8-12.fc37.x86_64",
        sha256 = "6e74a8ed5b472cf811f9bf429a999ed3f362e2c88566a461517a12c058abd401",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6e74a8ed5b472cf811f9bf429a999ed3f362e2c88566a461517a12c058abd401",
        ],
    )

    rpm(
        name = "ca-certificates-0__2023.2.60-1.0.fc37.x86_64",
        sha256 = "b2dcac3e49cbf75841d41ee1c53f1a91ffa78ba03dab8febb3153dbf76b2c5b2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b2dcac3e49cbf75841d41ee1c53f1a91ffa78ba03dab8febb3153dbf76b2c5b2",
        ],
    )

    rpm(
        name = "cairo-0__1.17.6-2.fc37.x86_64",
        sha256 = "98d99c9027f969940939a1ae5fea4bd4b6aa1139cdc9af6d3523dd8bc03aa135",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/98d99c9027f969940939a1ae5fea4bd4b6aa1139cdc9af6d3523dd8bc03aa135",
        ],
    )

    rpm(
        name = "cairo-gobject-0__1.17.6-2.fc37.x86_64",
        sha256 = "3bcf98fd61b70b1108581dbd04673099a31092b831352b65ab1db7944500866b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3bcf98fd61b70b1108581dbd04673099a31092b831352b65ab1db7944500866b",
        ],
    )

    rpm(
        name = "capstone-0__4.0.2-11.fc37.x86_64",
        sha256 = "5faaf0c29c0e76456c42b3a8e62b4fdf43150501187653da27974e6384655248",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5faaf0c29c0e76456c42b3a8e62b4fdf43150501187653da27974e6384655248",
        ],
    )

    rpm(
        name = "cdparanoia-libs-0__10.2-40.fc37.x86_64",
        sha256 = "2ab8a872b99ce73dcf6e8890f9fbd8089ec565b3d8be746347ea448686264cef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2ab8a872b99ce73dcf6e8890f9fbd8089ec565b3d8be746347ea448686264cef",
        ],
    )

    rpm(
        name = "checkpolicy-0__3.5-1.fc37.x86_64",
        sha256 = "1bd6036081fb219541cdbf1f23ef05381665356f3f45c4af8bb72ff0295f82ce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1bd6036081fb219541cdbf1f23ef05381665356f3f45c4af8bb72ff0295f82ce",
        ],
    )

    rpm(
        name = "cmake-filesystem-0__3.26.1-1.fc37.x86_64",
        sha256 = "c19584266e4ceb5cf255c73d2e95f41f5c98aff8e6f98bf5da5fb36154c7b874",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c19584266e4ceb5cf255c73d2e95f41f5c98aff8e6f98bf5da5fb36154c7b874",
        ],
    )

    rpm(
        name = "colord-libs-0__1.4.6-2.fc37.x86_64",
        sha256 = "4ee12a45efd84250f10d9dfb4096e37ea9caa6485f88d2705616002a75b7b1a2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ee12a45efd84250f10d9dfb4096e37ea9caa6485f88d2705616002a75b7b1a2",
        ],
    )

    rpm(
        name = "coreutils-0__9.1-7.fc37.x86_64",
        sha256 = "cd4f2bee79ba95edb4dd529a5a8488769c4538e91180495f1d81701ea1a5115d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cd4f2bee79ba95edb4dd529a5a8488769c4538e91180495f1d81701ea1a5115d",
        ],
    )

    rpm(
        name = "coreutils-common-0__9.1-7.fc37.x86_64",
        sha256 = "34e657305d9356b075c0fa58cdbfbb699bbf4b54c9a2c69534a1718faa8717d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/34e657305d9356b075c0fa58cdbfbb699bbf4b54c9a2c69534a1718faa8717d2",
        ],
    )

    rpm(
        name = "coreutils-single-0__9.1-7.fc37.x86_64",
        sha256 = "414bda840560471cb3d7380923ab00585ee78ca2db4b0d52155e9319a32151bc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/414bda840560471cb3d7380923ab00585ee78ca2db4b0d52155e9319a32151bc",
        ],
    )

    rpm(
        name = "corosynclib-0__3.1.7-1.fc37.x86_64",
        sha256 = "ff826d7e1b951c4601cc12ca6c21e549b29edba14d2309ec6452fcb8b73bf0d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ff826d7e1b951c4601cc12ca6c21e549b29edba14d2309ec6452fcb8b73bf0d6",
        ],
    )

    rpm(
        name = "cracklib-0__2.9.7-30.fc37.x86_64",
        sha256 = "3847abdc8ff973aeb0fb7e681bdf7c37b19cd49e5df17e8bf6bc35f34615c88f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3847abdc8ff973aeb0fb7e681bdf7c37b19cd49e5df17e8bf6bc35f34615c88f",
        ],
    )

    rpm(
        name = "crypto-policies-0__20220815-1.gite4ed860.fc37.x86_64",
        sha256 = "486a11feeaad706c68b05de60a906cc57059454cbce436aeba45f88b84578c0c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/486a11feeaad706c68b05de60a906cc57059454cbce436aeba45f88b84578c0c",
        ],
    )

    rpm(
        name = "crypto-policies-scripts-0__20220815-1.gite4ed860.fc37.x86_64",
        sha256 = "108dcc63b7aa387c1fbf2d3153b98447645c313132fea16e855b4549864f96cf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/108dcc63b7aa387c1fbf2d3153b98447645c313132fea16e855b4549864f96cf",
        ],
    )

    rpm(
        name = "cryptsetup-devel-0__2.6.1-1.fc37.x86_64",
        sha256 = "9074d1eefc27c1c551725fff48e77f6ed09a70be9303585161625163b3106fb8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9074d1eefc27c1c551725fff48e77f6ed09a70be9303585161625163b3106fb8",
        ],
    )

    rpm(
        name = "cryptsetup-libs-0__2.6.1-1.fc37.x86_64",
        sha256 = "8ea172dc9edc25482c09a0a09cd87384f001ce9bd80575f580439584d0dc15a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8ea172dc9edc25482c09a0a09cd87384f001ce9bd80575f580439584d0dc15a6",
        ],
    )

    rpm(
        name = "cups-libs-1__2.4.2-10.fc37.x86_64",
        sha256 = "f729c09320df821e46c21e7a57fbc09c15d2ed6426d602955a0a83866ddc1709",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f729c09320df821e46c21e7a57fbc09c15d2ed6426d602955a0a83866ddc1709",
        ],
    )

    rpm(
        name = "curl-minimal-0__7.85.0-8.fc37.x86_64",
        sha256 = "d331d0c957c9b3c6f4f0bf23959f3654415f55968908864bd70d3cc183821295",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d331d0c957c9b3c6f4f0bf23959f3654415f55968908864bd70d3cc183821295",
        ],
    )

    rpm(
        name = "cyrus-sasl-0__2.1.28-8.fc37.x86_64",
        sha256 = "1ede74bf11c2a8b3539a53176975f76531ceaf5bb525036b4740749e8a309484",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ede74bf11c2a8b3539a53176975f76531ceaf5bb525036b4740749e8a309484",
        ],
    )

    rpm(
        name = "cyrus-sasl-gssapi-0__2.1.28-8.fc37.x86_64",
        sha256 = "b1dd9f0a836c47adf0628eef6dac3dee1059959e8f22e9c857d0a1f0ee3ff415",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b1dd9f0a836c47adf0628eef6dac3dee1059959e8f22e9c857d0a1f0ee3ff415",
        ],
    )

    rpm(
        name = "cyrus-sasl-lib-0__2.1.28-8.fc37.x86_64",
        sha256 = "4e0e8656faf1f4f5227e4e40cdb4e662a1d78b19e74b90ba2f39f3cdf73e0083",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4e0e8656faf1f4f5227e4e40cdb4e662a1d78b19e74b90ba2f39f3cdf73e0083",
        ],
    )

    rpm(
        name = "daxctl-libs-0__76.1-1.fc37.x86_64",
        sha256 = "d0d4954b1a540de7a3536a5eba0c33fc5dd539c2f2d2dc547b33fa3508a3b7db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d0d4954b1a540de7a3536a5eba0c33fc5dd539c2f2d2dc547b33fa3508a3b7db",
        ],
    )

    rpm(
        name = "dbus-1__1.14.6-1.fc37.x86_64",
        sha256 = "671e7cb382f4cab02530739fd9c57463f6d2649571834e6874c8050abf556e68",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/671e7cb382f4cab02530739fd9c57463f6d2649571834e6874c8050abf556e68",
        ],
    )

    rpm(
        name = "dbus-broker-0__33-1.fc37.x86_64",
        sha256 = "069f79144219815854e47cda0bf47c5f5e361d48cbfa652405ac68c0d24d29ee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/069f79144219815854e47cda0bf47c5f5e361d48cbfa652405ac68c0d24d29ee",
        ],
    )

    rpm(
        name = "dbus-common-1__1.14.6-1.fc37.x86_64",
        sha256 = "2f5d8c77a752f02e4fc98f5ac53ca7ca2811831c8e805907dfce05007be95027",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2f5d8c77a752f02e4fc98f5ac53ca7ca2811831c8e805907dfce05007be95027",
        ],
    )

    rpm(
        name = "dbus-libs-1__1.14.6-1.fc37.x86_64",
        sha256 = "fa28aafdcc799e059650b44bcad03f2112b3e382877c2030a4f37f39176fc662",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fa28aafdcc799e059650b44bcad03f2112b3e382877c2030a4f37f39176fc662",
        ],
    )

    rpm(
        name = "device-mapper-0__1.02.175-9.fc37.x86_64",
        sha256 = "c57831b8629e2e31b3c55d4f0064cd25a515d3eb1ac61fc6897ce07421a2e91b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c57831b8629e2e31b3c55d4f0064cd25a515d3eb1ac61fc6897ce07421a2e91b",
        ],
    )

    rpm(
        name = "device-mapper-devel-0__1.02.175-9.fc37.x86_64",
        sha256 = "b9e7ecb9a4455460de47332dfed8d5b245f4f06b0ee7e2eb7fab415954a7300e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b9e7ecb9a4455460de47332dfed8d5b245f4f06b0ee7e2eb7fab415954a7300e",
        ],
    )

    rpm(
        name = "device-mapper-event-0__1.02.175-9.fc37.x86_64",
        sha256 = "bf91322129e4c9434e0bc30aacabfac8446287d4c1316d5a5de9d0416bbe49e7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf91322129e4c9434e0bc30aacabfac8446287d4c1316d5a5de9d0416bbe49e7",
        ],
    )

    rpm(
        name = "device-mapper-event-libs-0__1.02.175-9.fc37.x86_64",
        sha256 = "700bb4bcaf12282ca54b83e70bfb296302a1a55d5072c7e858499f23e2442bd6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/700bb4bcaf12282ca54b83e70bfb296302a1a55d5072c7e858499f23e2442bd6",
        ],
    )

    rpm(
        name = "device-mapper-libs-0__1.02.175-9.fc37.x86_64",
        sha256 = "7c0f72217eacc9b5caf553c17cb2428de242094dc7e0e1dbd0d21869d909c7d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c0f72217eacc9b5caf553c17cb2428de242094dc7e0e1dbd0d21869d909c7d2",
        ],
    )

    rpm(
        name = "device-mapper-multipath-libs-0__0.9.0-4.fc37.x86_64",
        sha256 = "e45688002b35af16a162d2ea5c335e54dd5968212ef4c365ff7b3dd31fb4da4b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e45688002b35af16a162d2ea5c335e54dd5968212ef4c365ff7b3dd31fb4da4b",
        ],
    )

    rpm(
        name = "device-mapper-persistent-data-0__0.9.0-8.fc37.x86_64",
        sha256 = "66e967462b711f2b62856dcb115ce53e06d58268ad7e20aa31ef282ea42b0faa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66e967462b711f2b62856dcb115ce53e06d58268ad7e20aa31ef282ea42b0faa",
        ],
    )

    rpm(
        name = "diffutils-0__3.8-3.fc37.x86_64",
        sha256 = "c1374e3372d0d246ecb0e04b36743e23c68ab307c7603c5a267fce654bf05cdd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c1374e3372d0d246ecb0e04b36743e23c68ab307c7603c5a267fce654bf05cdd",
        ],
    )

    rpm(
        name = "dmidecode-1__3.4-2.fc37.x86_64",
        sha256 = "19b7e1385a12e9196622d1549ebee9a2254656300a530a9abb824737950571ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/19b7e1385a12e9196622d1549ebee9a2254656300a530a9abb824737950571ef",
        ],
    )

    rpm(
        name = "dnsmasq-0__2.89-1.fc37.x86_64",
        sha256 = "d0e430451052251fec9810fe47707e894c9b9b568b553d527a98e0d032ce090f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d0e430451052251fec9810fe47707e894c9b9b568b553d527a98e0d032ce090f",
        ],
    )

    rpm(
        name = "e2fsprogs-libs-0__1.46.5-3.fc37.x86_64",
        sha256 = "631c5cdd65015cf905cf9c7b9c0213384a524eeb0b1f15da971aae8cc38ed27e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/631c5cdd65015cf905cf9c7b9c0213384a524eeb0b1f15da971aae8cc38ed27e",
        ],
    )

    rpm(
        name = "ebtables-legacy-0__2.0.11-12.fc37.x86_64",
        sha256 = "22b7c41ae3ac2ff29801bf9e10f7408c945147d39e7da672d5c250cc1b4c1a2e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/22b7c41ae3ac2ff29801bf9e10f7408c945147d39e7da672d5c250cc1b4c1a2e",
        ],
    )

    rpm(
        name = "edk2-ovmf-0__20230301gitf80f052277c8-1.fc37.x86_64",
        sha256 = "49252550260575a19fe48d84981203a3352e89513db43112ac393711a19399ba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49252550260575a19fe48d84981203a3352e89513db43112ac393711a19399ba",
        ],
    )

    rpm(
        name = "elfutils-default-yama-scope-0__0.189-1.fc37.x86_64",
        sha256 = "d47043e1562a37dab1674d7fc09b42797cebdbe2cd545008574e85b65a5d011e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d47043e1562a37dab1674d7fc09b42797cebdbe2cd545008574e85b65a5d011e",
        ],
    )

    rpm(
        name = "elfutils-libelf-0__0.189-1.fc37.x86_64",
        sha256 = "856a761052deed45559ddd5be420d8f29a6771738c3d3a2eef913a8c5c89c22e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/856a761052deed45559ddd5be420d8f29a6771738c3d3a2eef913a8c5c89c22e",
        ],
    )

    rpm(
        name = "elfutils-libs-0__0.189-1.fc37.x86_64",
        sha256 = "cdb6c7b26e4cd92d4c88f8abc76c925d9b509a22062e5c079ddbb8e17453a65a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cdb6c7b26e4cd92d4c88f8abc76c925d9b509a22062e5c079ddbb8e17453a65a",
        ],
    )

    rpm(
        name = "expat-0__2.5.0-1.fc37.x86_64",
        sha256 = "0e49c2393e5507bbaa16ededf0176e731e0196dd3230f6371d67be8b919e3429",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e49c2393e5507bbaa16ededf0176e731e0196dd3230f6371d67be8b919e3429",
        ],
    )

    rpm(
        name = "fedora-gpg-keys-0__37-2.x86_64",
        sha256 = "47a0fdf0c8d0aecd3d4b2eee160affec5ba0d12b7ac6647b3f12fdef275e9738",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/47a0fdf0c8d0aecd3d4b2eee160affec5ba0d12b7ac6647b3f12fdef275e9738",
        ],
    )

    rpm(
        name = "fedora-release-common-0__37-16.x86_64",
        sha256 = "5887ea74e3b3525a31fc0a685e10b8ef0be80afe223a9d327c53a5a3168e36d7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5887ea74e3b3525a31fc0a685e10b8ef0be80afe223a9d327c53a5a3168e36d7",
        ],
    )

    rpm(
        name = "fedora-release-container-0__37-16.x86_64",
        sha256 = "2321ec7a64f24b616f6fef130a97f257aff81b7068b1ede4f81938395e8bab56",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2321ec7a64f24b616f6fef130a97f257aff81b7068b1ede4f81938395e8bab56",
        ],
    )

    rpm(
        name = "fedora-release-identity-cinnamon-0__37-16.x86_64",
        sha256 = "d05ecae860d5617e0312249c37d64f801a7421c151eddb1e351a6b02ddbf08e1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d05ecae860d5617e0312249c37d64f801a7421c151eddb1e351a6b02ddbf08e1",
        ],
    )

    rpm(
        name = "fedora-release-identity-i3-0__37-16.x86_64",
        sha256 = "2d241168621df87b6779a6c9a213a31270f39bb3ceb79ec2007f6110287d4bea",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2d241168621df87b6779a6c9a213a31270f39bb3ceb79ec2007f6110287d4bea",
        ],
    )

    rpm(
        name = "fedora-release-identity-kinoite-0__37-16.x86_64",
        sha256 = "ec85bca6a96d9dbd8c74784c2bb268c658ceca13ef6f1eb209efe72c4a949c5c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ec85bca6a96d9dbd8c74784c2bb268c658ceca13ef6f1eb209efe72c4a949c5c",
        ],
    )

    rpm(
        name = "fedora-release-identity-server-0__37-16.x86_64",
        sha256 = "3baba506d4d32de6e4f267317cbc22241686dd39d963c91927b4ab77b5b633ba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3baba506d4d32de6e4f267317cbc22241686dd39d963c91927b4ab77b5b633ba",
        ],
    )

    rpm(
        name = "fedora-repos-0__37-2.x86_64",
        sha256 = "f43a00322ae512135f695e9378eadcb3f8a8314bd4e290ea40c7c576621297f6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f43a00322ae512135f695e9378eadcb3f8a8314bd4e290ea40c7c576621297f6",
        ],
    )

    rpm(
        name = "filesystem-0__3.18-2.fc37.x86_64",
        sha256 = "1c28f722e7f3e48dba7ebf4f763ebebc6688b9e0fd58b55ba4fcd884c8180ef4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1c28f722e7f3e48dba7ebf4f763ebebc6688b9e0fd58b55ba4fcd884c8180ef4",
        ],
    )

    rpm(
        name = "findutils-1__4.9.0-2.fc37.x86_64",
        sha256 = "25cd555f1a70138b3e81ede1cd375cb620e7a3de05680c9ebaa764f1261d0ce3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/25cd555f1a70138b3e81ede1cd375cb620e7a3de05680c9ebaa764f1261d0ce3",
        ],
    )

    rpm(
        name = "flac-libs-0__1.3.4-2.fc37.x86_64",
        sha256 = "17e5a1f4b60a9b54fc00c0e51c9c7d7a07439bddacf6c3d02bf6faab9627b776",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17e5a1f4b60a9b54fc00c0e51c9c7d7a07439bddacf6c3d02bf6faab9627b776",
        ],
    )

    rpm(
        name = "fmt-0__9.1.0-1.fc37.x86_64",
        sha256 = "0323e555d954bf9ceb4f1b242f246576ac39beb3b9111085597a6f75cab14fdf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0323e555d954bf9ceb4f1b242f246576ac39beb3b9111085597a6f75cab14fdf",
        ],
    )

    rpm(
        name = "fontconfig-0__2.14.1-2.fc37.x86_64",
        sha256 = "b90b5481bf7627b4456c400a87ce70567aeb6a9770de45301c71fef7e6e6e785",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b90b5481bf7627b4456c400a87ce70567aeb6a9770de45301c71fef7e6e6e785",
        ],
    )

    rpm(
        name = "fonts-filesystem-1__2.0.5-9.fc37.x86_64",
        sha256 = "28268a1c876459ab9d7065c809df76cdf5971c40ec6d22e8fd6395b0d6b81db3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/28268a1c876459ab9d7065c809df76cdf5971c40ec6d22e8fd6395b0d6b81db3",
        ],
    )

    rpm(
        name = "freetype-0__2.12.1-3.fc37.x86_64",
        sha256 = "636bd6f0e69d1b44cb1c8424633044832ecade842831d38c475c2dd26c9cfb6a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/636bd6f0e69d1b44cb1c8424633044832ecade842831d38c475c2dd26c9cfb6a",
        ],
    )

    rpm(
        name = "fribidi-0__1.0.12-2.fc37.x86_64",
        sha256 = "4e9dfc311ab9aab380a55ac899ab35648fc9e69f766688456163ce277a64503c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4e9dfc311ab9aab380a55ac899ab35648fc9e69f766688456163ce277a64503c",
        ],
    )

    rpm(
        name = "fuse-0__2.9.9-15.fc37.x86_64",
        sha256 = "de73e6a453152a5e3635d4796d4b92587f7ad9428647cf85a9bc9506839ed1e8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de73e6a453152a5e3635d4796d4b92587f7ad9428647cf85a9bc9506839ed1e8",
        ],
    )

    rpm(
        name = "fuse-common-0__3.10.5-5.fc37.x86_64",
        sha256 = "1f2c3773b6d2a19f5b2a3ea190b210a9f06e186eff763e70662c63fe5c85d394",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1f2c3773b6d2a19f5b2a3ea190b210a9f06e186eff763e70662c63fe5c85d394",
        ],
    )

    rpm(
        name = "fuse-libs-0__2.9.9-15.fc37.x86_64",
        sha256 = "e084290b302fdae060c81e90504c960a49061e0058033903b45fb27b384aa283",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e084290b302fdae060c81e90504c960a49061e0058033903b45fb27b384aa283",
        ],
    )

    rpm(
        name = "fuse3-libs-0__3.10.5-5.fc37.x86_64",
        sha256 = "559626f87751e9b10db9237e0bf05589081f69a5436ee88344352cbb5c4ef7cf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/559626f87751e9b10db9237e0bf05589081f69a5436ee88344352cbb5c4ef7cf",
        ],
    )

    rpm(
        name = "gawk-0__5.1.1-4.fc37.x86_64",
        sha256 = "6caea2f79e9fadf96e6cd55eac3f8625137b12f6a2ca75fb5e36b453dfe54edd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6caea2f79e9fadf96e6cd55eac3f8625137b12f6a2ca75fb5e36b453dfe54edd",
        ],
    )

    rpm(
        name = "gdbm-libs-1__1.23-2.fc37.x86_64",
        sha256 = "32ab362365afcf96144ba3e65c461cf6f8d495651d0c99fb4eeb970fc2b838e5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32ab362365afcf96144ba3e65c461cf6f8d495651d0c99fb4eeb970fc2b838e5",
        ],
    )

    rpm(
        name = "gdk-pixbuf2-0__2.42.10-1.fc37.x86_64",
        sha256 = "16105c47bfd8a6f5c315e6a1503b49959c4a1264de47b45c9806940ee1c9e108",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/16105c47bfd8a6f5c315e6a1503b49959c4a1264de47b45c9806940ee1c9e108",
        ],
    )

    rpm(
        name = "gdk-pixbuf2-modules-0__2.42.10-1.fc37.x86_64",
        sha256 = "aba6bfdfb82062959cf64df631916213604c3bdb12b172dc41b7b8a625e148e0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aba6bfdfb82062959cf64df631916213604c3bdb12b172dc41b7b8a625e148e0",
        ],
    )

    rpm(
        name = "gettext-envsubst-0__0.21.1-1.fc37.x86_64",
        sha256 = "797ba3836bebbf955b6df4258d2db042063e4c7acb9b076012e3e0cf6f47011a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/797ba3836bebbf955b6df4258d2db042063e4c7acb9b076012e3e0cf6f47011a",
        ],
    )

    rpm(
        name = "gettext-libs-0__0.21.1-1.fc37.x86_64",
        sha256 = "971fa1897680062f9f6a90d6717f5db0548b4bd304be346f0ff1e1b197989a18",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/971fa1897680062f9f6a90d6717f5db0548b4bd304be346f0ff1e1b197989a18",
        ],
    )

    rpm(
        name = "gettext-runtime-0__0.21.1-1.fc37.x86_64",
        sha256 = "27aac6ad641e8308216ae2b433982cb5807c9e4bd4b346ada67a0cc22ac3959e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/27aac6ad641e8308216ae2b433982cb5807c9e4bd4b346ada67a0cc22ac3959e",
        ],
    )

    rpm(
        name = "glib2-0__2.74.6-1.fc37.x86_64",
        sha256 = "57f55072a259bdfe0fe1bab6f8ae2808bd19858214975885a16727b46213c33f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/57f55072a259bdfe0fe1bab6f8ae2808bd19858214975885a16727b46213c33f",
        ],
    )

    rpm(
        name = "glibc-0__2.36-9.fc37.x86_64",
        sha256 = "8c8463cd9f194f03ea1607670399e2fbf068857f566c43dd07d351228c25f187",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c8463cd9f194f03ea1607670399e2fbf068857f566c43dd07d351228c25f187",
        ],
    )

    rpm(
        name = "glibc-common-0__2.36-9.fc37.x86_64",
        sha256 = "4237c10e5edacc5d5a9ea88e9fc5fef37249d459b13d4a0715c7836374a8da7a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4237c10e5edacc5d5a9ea88e9fc5fef37249d459b13d4a0715c7836374a8da7a",
        ],
    )

    rpm(
        name = "glibc-langpack-ce-0__2.36-9.fc37.x86_64",
        sha256 = "874dabbe4bfd348bd033cde90935c4d7cc738c595ba6dffaa1ebd490e589427d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/874dabbe4bfd348bd033cde90935c4d7cc738c595ba6dffaa1ebd490e589427d",
        ],
    )

    rpm(
        name = "glibc-langpack-lb-0__2.36-9.fc37.x86_64",
        sha256 = "7842e5f8a934dc732b246458187a7fffb01356aad2baf0318fba61eaf0f02aff",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7842e5f8a934dc732b246458187a7fffb01356aad2baf0318fba61eaf0f02aff",
        ],
    )

    rpm(
        name = "glibc-langpack-ml-0__2.36-9.fc37.x86_64",
        sha256 = "229ac8c343ef69902f7a67c0e1dac8ed552e004787ee3d96715d075a13eb8a9e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/229ac8c343ef69902f7a67c0e1dac8ed552e004787ee3d96715d075a13eb8a9e",
        ],
    )

    rpm(
        name = "glibc-langpack-mt-0__2.36-9.fc37.x86_64",
        sha256 = "9c2887bba0dfa28208aee3aad12faa64ab806a4bdebe810b63a1389113b37b95",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9c2887bba0dfa28208aee3aad12faa64ab806a4bdebe810b63a1389113b37b95",
        ],
    )

    rpm(
        name = "glibmm2.4-0__2.66.5-2.fc37.x86_64",
        sha256 = "39c024b94acdda410fe5f61641d66c198bb325df2d128485b8f26fd835e40ee3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/39c024b94acdda410fe5f61641d66c198bb325df2d128485b8f26fd835e40ee3",
        ],
    )

    rpm(
        name = "glusterfs-0__10.3-1.fc37.x86_64",
        sha256 = "552ed8d8e242cb3dc55e226733f09b21f70f471b1ed7754efa48ddfb8988c9a9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/552ed8d8e242cb3dc55e226733f09b21f70f471b1ed7754efa48ddfb8988c9a9",
        ],
    )

    rpm(
        name = "glusterfs-cli-0__10.3-1.fc37.x86_64",
        sha256 = "fbbfa1ba228d03d5f2c8583510a9dfef89a45a27f273ef276046ad69fb657b42",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fbbfa1ba228d03d5f2c8583510a9dfef89a45a27f273ef276046ad69fb657b42",
        ],
    )

    rpm(
        name = "glusterfs-client-xlators-0__10.3-1.fc37.x86_64",
        sha256 = "773b3d3082f63c06fba5e58c4f1e0b5240c282040c20574ba537a903373d2ed7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/773b3d3082f63c06fba5e58c4f1e0b5240c282040c20574ba537a903373d2ed7",
        ],
    )

    rpm(
        name = "glusterfs-fuse-0__10.3-1.fc37.x86_64",
        sha256 = "e4ffd7a8391ccc1de1a827d4db38f0e2c7f95a21e70e21c227c1831ab07bf0f5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e4ffd7a8391ccc1de1a827d4db38f0e2c7f95a21e70e21c227c1831ab07bf0f5",
        ],
    )

    rpm(
        name = "gmp-1__6.2.1-3.fc37.x86_64",
        sha256 = "42c8a66f1efcdffaf611e70395e16311f6c56ef795ee2a43c2a48c55eef77734",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42c8a66f1efcdffaf611e70395e16311f6c56ef795ee2a43c2a48c55eef77734",
        ],
    )

    rpm(
        name = "gnutls-0__3.8.0-2.fc37.x86_64",
        sha256 = "91b08de00abe5430f61bc5491ea1f11a23712877da5ab9828865e5c17d4841ee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/91b08de00abe5430f61bc5491ea1f11a23712877da5ab9828865e5c17d4841ee",
        ],
    )

    rpm(
        name = "gnutls-dane-0__3.8.0-2.fc37.x86_64",
        sha256 = "1c3e6fa21d05e78d64c3f167a6a20f8a674762cbb5ff10eddb8f934d7058a22e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1c3e6fa21d05e78d64c3f167a6a20f8a674762cbb5ff10eddb8f934d7058a22e",
        ],
    )

    rpm(
        name = "gnutls-utils-0__3.8.0-2.fc37.x86_64",
        sha256 = "3afde268a04428d152408b91541bd61a0e5bb43654df296d198de337029d3b54",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3afde268a04428d152408b91541bd61a0e5bb43654df296d198de337029d3b54",
        ],
    )

    rpm(
        name = "google-noto-fonts-common-0__20201206__caret__1.git0c78c8329-7.fc37.x86_64",
        sha256 = "dee4c1932baf30c73256c0e718523b1db27bc2eaeb0974bef757d3950b22f728",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dee4c1932baf30c73256c0e718523b1db27bc2eaeb0974bef757d3950b22f728",
        ],
    )

    rpm(
        name = "google-noto-sans-vf-fonts-0__20201206__caret__1.git0c78c8329-7.fc37.x86_64",
        sha256 = "dc87fdd4f290064820ba52b124f3a083ec2480ac337087f8b63b26da92e5fab2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc87fdd4f290064820ba52b124f3a083ec2480ac337087f8b63b26da92e5fab2",
        ],
    )

    rpm(
        name = "gperftools-libs-0__2.9.1-4.fc37.x86_64",
        sha256 = "25680f3855b5c25a2097a0ec6f63f69e7bcd328e1cd16cfd6a0b4417adac88a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/25680f3855b5c25a2097a0ec6f63f69e7bcd328e1cd16cfd6a0b4417adac88a6",
        ],
    )

    rpm(
        name = "graphene-0__1.10.6-4.fc37.x86_64",
        sha256 = "e184c1900b2ed7062a0ccb6eae9fbbeb616d9ebdd831f4aeeca13669633fe15e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e184c1900b2ed7062a0ccb6eae9fbbeb616d9ebdd831f4aeeca13669633fe15e",
        ],
    )

    rpm(
        name = "graphite2-0__1.3.14-10.fc37.x86_64",
        sha256 = "9720bfa82e7d2a0e7d20361437ab08f18f83b3d84a2e0bcf9ddc1cd3979ac5c6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9720bfa82e7d2a0e7d20361437ab08f18f83b3d84a2e0bcf9ddc1cd3979ac5c6",
        ],
    )

    rpm(
        name = "grep-0__3.7-4.fc37.x86_64",
        sha256 = "d997786e71f2c7b4a9ed1323b8684ec1802e49a866fb0c1b69101531440cb464",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d997786e71f2c7b4a9ed1323b8684ec1802e49a866fb0c1b69101531440cb464",
        ],
    )

    rpm(
        name = "groff-base-0__1.22.4-10.fc37.x86_64",
        sha256 = "b5b4e759d1c56188fb777926de0d17498c25d3234d2635ce5a8e7b000bfaf7f3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b5b4e759d1c56188fb777926de0d17498c25d3234d2635ce5a8e7b000bfaf7f3",
        ],
    )

    rpm(
        name = "gsm-0__1.0.22-1.fc37.x86_64",
        sha256 = "55b8fef81cf1e8ac5e17a74e40e86270924c92d7d58de0cd09d97ec47ac05de2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/55b8fef81cf1e8ac5e17a74e40e86270924c92d7d58de0cd09d97ec47ac05de2",
        ],
    )

    rpm(
        name = "gssproxy-0__0.9.1-4.fc37.x86_64",
        sha256 = "80008c8966ae7ba4a684fb078dc663122a2257695b71d314939e5404ad1c7b9e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/80008c8966ae7ba4a684fb078dc663122a2257695b71d314939e5404ad1c7b9e",
        ],
    )

    rpm(
        name = "gstreamer1-0__1.20.5-1.fc37.x86_64",
        sha256 = "9b606f12f1b4d5d8dfff2f1b8aab197d12b7ab232d78888179495a97bfb6d854",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9b606f12f1b4d5d8dfff2f1b8aab197d12b7ab232d78888179495a97bfb6d854",
        ],
    )

    rpm(
        name = "gstreamer1-plugins-base-0__1.20.5-1.fc37.x86_64",
        sha256 = "c329cd49c6b7bdce90ca7283277715a14e039b7841998666cdca166d025bc622",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c329cd49c6b7bdce90ca7283277715a14e039b7841998666cdca166d025bc622",
        ],
    )

    rpm(
        name = "gtk-update-icon-cache-0__3.24.37-1.fc37.x86_64",
        sha256 = "550234bbd7bb6165601d49aa43de7b1605f5d5406c8921a7dbca35fb8462fccd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/550234bbd7bb6165601d49aa43de7b1605f5d5406c8921a7dbca35fb8462fccd",
        ],
    )

    rpm(
        name = "gtk3-0__3.24.37-1.fc37.x86_64",
        sha256 = "c984cd191548f29568033e23211308e2fb98263d7cfd17f420c45d341d27c381",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c984cd191548f29568033e23211308e2fb98263d7cfd17f420c45d341d27c381",
        ],
    )

    rpm(
        name = "gzip-0__1.12-2.fc37.x86_64",
        sha256 = "3ef9e1b938dd19c5268004e370d90f8a8ae0dbc664715457a371ce900ee7736c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3ef9e1b938dd19c5268004e370d90f8a8ae0dbc664715457a371ce900ee7736c",
        ],
    )

    rpm(
        name = "harfbuzz-0__5.2.0-1.fc37.x86_64",
        sha256 = "72cbd4281a2a68ee05d191145ec3abd35fd175b2b8a47aa238fac9fabe8dc82c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/72cbd4281a2a68ee05d191145ec3abd35fd175b2b8a47aa238fac9fabe8dc82c",
        ],
    )

    rpm(
        name = "hicolor-icon-theme-0__0.17-14.fc37.x86_64",
        sha256 = "67ee7f6052ac62293c1ef1208ea9064b793e840198fdc1e4ec912fc631416fb0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/67ee7f6052ac62293c1ef1208ea9064b793e840198fdc1e4ec912fc631416fb0",
        ],
    )

    rpm(
        name = "highway-0__1.0.4-1.fc37.x86_64",
        sha256 = "5a06201fcc0341a429cf30661e0267cd9f08d3fa4bec0f0473fc580e8785bb44",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5a06201fcc0341a429cf30661e0267cd9f08d3fa4bec0f0473fc580e8785bb44",
        ],
    )

    rpm(
        name = "hwdata-0__0.368-1.fc37.x86_64",
        sha256 = "e9ae6e555c0d7d0e8094fa3ad033813a67ca66acdfdd217a4f44d5dc3137e39a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e9ae6e555c0d7d0e8094fa3ad033813a67ca66acdfdd217a4f44d5dc3137e39a",
        ],
    )

    rpm(
        name = "iproute-0__5.18.0-2.fc37.x86_64",
        sha256 = "960633a9035f8fe02d49ba6713f240363383d5dc3b3f862d28e9b211fd34fea3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/960633a9035f8fe02d49ba6713f240363383d5dc3b3f862d28e9b211fd34fea3",
        ],
    )

    rpm(
        name = "iproute-tc-0__5.18.0-2.fc37.x86_64",
        sha256 = "926162e7602c59dd1af47b71604fe4bd332a24cdf7853e9f266ea925f771d9be",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/926162e7602c59dd1af47b71604fe4bd332a24cdf7853e9f266ea925f771d9be",
        ],
    )

    rpm(
        name = "iptables-legacy-0__1.8.8-4.fc37.x86_64",
        sha256 = "9eca37e404dc7ee3a943c1e7c8fcd1f73fc2bc03c03916611aec4bad0e8f854e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9eca37e404dc7ee3a943c1e7c8fcd1f73fc2bc03c03916611aec4bad0e8f854e",
        ],
    )

    rpm(
        name = "iptables-legacy-libs-0__1.8.8-4.fc37.x86_64",
        sha256 = "83a5e8a239d9904c2901258afc535f4bbb3beac96742610ed6876c3da2d57d5e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/83a5e8a239d9904c2901258afc535f4bbb3beac96742610ed6876c3da2d57d5e",
        ],
    )

    rpm(
        name = "iptables-libs-0__1.8.8-4.fc37.x86_64",
        sha256 = "f7f3538f7a76060e8698add713bae9a21c7b67fa1a1f1005582881351ff2ef28",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f7f3538f7a76060e8698add713bae9a21c7b67fa1a1f1005582881351ff2ef28",
        ],
    )

    rpm(
        name = "ipxe-roms-qemu-0__20220210-2.git64113751.fc37.x86_64",
        sha256 = "541c4cdbb73deecd93b68ef8a5dc50b8bc32ccd0c9f9480c2a35f318dda38f2e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/541c4cdbb73deecd93b68ef8a5dc50b8bc32ccd0c9f9480c2a35f318dda38f2e",
        ],
    )

    rpm(
        name = "iscsi-initiator-utils-0__6.2.1.4-6.git2a8f9d8.fc37.x86_64",
        sha256 = "c64cc5f44a660062ced56dc143312f87e73bbd3680bb1e1bc2c6463b938cd6cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c64cc5f44a660062ced56dc143312f87e73bbd3680bb1e1bc2c6463b938cd6cd",
        ],
    )

    rpm(
        name = "iscsi-initiator-utils-iscsiuio-0__6.2.1.4-6.git2a8f9d8.fc37.x86_64",
        sha256 = "49d75609083bc45ba147c7611bab6b90d4112f2b2cca6317623c94a1a8445306",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49d75609083bc45ba147c7611bab6b90d4112f2b2cca6317623c94a1a8445306",
        ],
    )

    rpm(
        name = "isns-utils-libs-0__0.101-5.fc37.x86_64",
        sha256 = "e35732720639a673041e686f924e5298afeff67d2bcf0d9b5e7833acec563872",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e35732720639a673041e686f924e5298afeff67d2bcf0d9b5e7833acec563872",
        ],
    )

    rpm(
        name = "iso-codes-0__4.11.0-1.fc37.x86_64",
        sha256 = "bbab7c3e2e0304440f72ec6b2c4cb88c81b710f0f0fd8f71cfbc7eccd0660c06",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bbab7c3e2e0304440f72ec6b2c4cb88c81b710f0f0fd8f71cfbc7eccd0660c06",
        ],
    )

    rpm(
        name = "jack-audio-connection-kit-0__1.9.21-3.fc37.x86_64",
        sha256 = "6b578cbbc8f96799353e8c0c2bd9e19e949f53b09595b958f8e60c5226723eda",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6b578cbbc8f96799353e8c0c2bd9e19e949f53b09595b958f8e60c5226723eda",
        ],
    )

    rpm(
        name = "jbigkit-libs-0__2.1-24.fc37.x86_64",
        sha256 = "3c27e55dcfab1bbdb11fa8f7f38e186d716b65236625580fce5e9a1587a11334",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3c27e55dcfab1bbdb11fa8f7f38e186d716b65236625580fce5e9a1587a11334",
        ],
    )

    rpm(
        name = "json-c-0__0.16-3.fc37.x86_64",
        sha256 = "e7c83a9058c7e7e05e4c7ba97a363414eb973343ea8f00a1140fbdafe6ca67e2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7c83a9058c7e7e05e4c7ba97a363414eb973343ea8f00a1140fbdafe6ca67e2",
        ],
    )

    rpm(
        name = "json-c-devel-0__0.16-3.fc37.x86_64",
        sha256 = "a60557c7450d68d2f4310565f87a830334ca0866d81413bd3a025db3c3f9cde3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a60557c7450d68d2f4310565f87a830334ca0866d81413bd3a025db3c3f9cde3",
        ],
    )

    rpm(
        name = "json-glib-0__1.6.6-3.fc37.x86_64",
        sha256 = "598ddb966e5ac4f664f2cbe9d0087b62c1a4067a1f1af24f290d7f681956e29c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/598ddb966e5ac4f664f2cbe9d0087b62c1a4067a1f1af24f290d7f681956e29c",
        ],
    )

    rpm(
        name = "kbd-0__2.5.1-3.fc37.x86_64",
        sha256 = "7c660eada8cb6e2d2b0c035a9d1696db945761c07e7900b8003b1b405243adef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c660eada8cb6e2d2b0c035a9d1696db945761c07e7900b8003b1b405243adef",
        ],
    )

    rpm(
        name = "kbd-legacy-0__2.5.1-3.fc37.x86_64",
        sha256 = "a6e8f6c0a9973883d33d35381234a88acee5d075ba0a92f811e9728441756d1f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a6e8f6c0a9973883d33d35381234a88acee5d075ba0a92f811e9728441756d1f",
        ],
    )

    rpm(
        name = "kbd-misc-0__2.5.1-3.fc37.x86_64",
        sha256 = "b71602703b63f87199a145cd64cf804b19af7275eede730bd3b82296b64417d7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b71602703b63f87199a145cd64cf804b19af7275eede730bd3b82296b64417d7",
        ],
    )

    rpm(
        name = "keyutils-0__1.6.1-5.fc37.x86_64",
        sha256 = "6559c77e8b828179ff6c55ca47c68ee6d00c5f13f6f6ed1d2cc5b6d0bb48502e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6559c77e8b828179ff6c55ca47c68ee6d00c5f13f6f6ed1d2cc5b6d0bb48502e",
        ],
    )

    rpm(
        name = "keyutils-libs-0__1.6.1-5.fc37.x86_64",
        sha256 = "e3fd19c3020e55d80b8a24edb68506d2adbb07b2db29eecbde91facae1cca59d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3fd19c3020e55d80b8a24edb68506d2adbb07b2db29eecbde91facae1cca59d",
        ],
    )

    rpm(
        name = "kmod-0__30-2.fc37.x86_64",
        sha256 = "b57193efad83c9cdb3acf6ad843e1ef17b8c00382a8395713b1480905e23f786",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b57193efad83c9cdb3acf6ad843e1ef17b8c00382a8395713b1480905e23f786",
        ],
    )

    rpm(
        name = "kmod-libs-0__30-2.fc37.x86_64",
        sha256 = "73a1a0f041819c1d50501a699945f0121a3b6e1f54df40cd0bf8f94b1b261ef5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/73a1a0f041819c1d50501a699945f0121a3b6e1f54df40cd0bf8f94b1b261ef5",
        ],
    )

    rpm(
        name = "krb5-libs-0__1.19.2-13.fc37.x86_64",
        sha256 = "5f2ffaa4084cb8918d3990ef352dbfdd9ac28d30c2ed2693c1011641199bb369",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5f2ffaa4084cb8918d3990ef352dbfdd9ac28d30c2ed2693c1011641199bb369",
        ],
    )

    rpm(
        name = "lame-libs-0__3.100-13.fc37.x86_64",
        sha256 = "6ac2008512a5d295efff80d0d54a481660147decfd86ef5d1ca89e25ce2795f9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6ac2008512a5d295efff80d0d54a481660147decfd86ef5d1ca89e25ce2795f9",
        ],
    )

    rpm(
        name = "langpacks-core-font-en-0__3.0-26.fc37.x86_64",
        sha256 = "a52ff26d4742ac302a10c68270bf5b63e538b6c0ebca6d5211eeb274c033aa21",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a52ff26d4742ac302a10c68270bf5b63e538b6c0ebca6d5211eeb274c033aa21",
        ],
    )

    rpm(
        name = "lcms2-0__2.14-1.fc37.x86_64",
        sha256 = "75764b4452ec28c6f52d8613a455fa01e6fc493f5b897b5fe3829710603d02c5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/75764b4452ec28c6f52d8613a455fa01e6fc493f5b897b5fe3829710603d02c5",
        ],
    )

    rpm(
        name = "libX11-0__1.8.4-1.fc37.x86_64",
        sha256 = "e49e2834b484c6317e551d3ad364f8981c323269465078b511b6cb36fd99d226",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e49e2834b484c6317e551d3ad364f8981c323269465078b511b6cb36fd99d226",
        ],
    )

    rpm(
        name = "libX11-common-0__1.8.4-1.fc37.x86_64",
        sha256 = "9ee693831f05b8fd18dfd86b439d68c184d715762f3feafccca0bd8a0db0ee30",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9ee693831f05b8fd18dfd86b439d68c184d715762f3feafccca0bd8a0db0ee30",
        ],
    )

    rpm(
        name = "libX11-xcb-0__1.8.4-1.fc37.x86_64",
        sha256 = "522aca73d4cfb7cc404ba656558953d0534791ae626984461460a3bae5c84075",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/522aca73d4cfb7cc404ba656558953d0534791ae626984461460a3bae5c84075",
        ],
    )

    rpm(
        name = "libXau-0__1.0.10-1.fc37.x86_64",
        sha256 = "c4000c0086e6db010b9ecbbf7c629970f00b7735e8bc8a731d264fd138d27656",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c4000c0086e6db010b9ecbbf7c629970f00b7735e8bc8a731d264fd138d27656",
        ],
    )

    rpm(
        name = "libXcomposite-0__0.4.5-8.fc37.x86_64",
        sha256 = "199937c38646941ce3cf35915f806399ee7610e789396748bfeae45545d39365",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/199937c38646941ce3cf35915f806399ee7610e789396748bfeae45545d39365",
        ],
    )

    rpm(
        name = "libXcursor-0__1.2.1-2.fc37.x86_64",
        sha256 = "bb19773d681f3291c8efc58149ff4662def311b042366b51118aa8143559c6f7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bb19773d681f3291c8efc58149ff4662def311b042366b51118aa8143559c6f7",
        ],
    )

    rpm(
        name = "libXdamage-0__1.1.5-8.fc37.x86_64",
        sha256 = "7c4cf6920d9004568c03cfc52dcdd7b688ce263977aea06544a9e632e9d9f6b0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c4cf6920d9004568c03cfc52dcdd7b688ce263977aea06544a9e632e9d9f6b0",
        ],
    )

    rpm(
        name = "libXext-0__1.3.4-9.fc37.x86_64",
        sha256 = "66de2c87ae690e8ec21c01749bb0eb223346aedf9f388096441ad5fa21be48db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66de2c87ae690e8ec21c01749bb0eb223346aedf9f388096441ad5fa21be48db",
        ],
    )

    rpm(
        name = "libXfixes-0__6.0.0-4.fc37.x86_64",
        sha256 = "289e2b9f9f42d7ef2f7fec88cfaf6b7983694aa6b45efff059d2081d47e10088",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/289e2b9f9f42d7ef2f7fec88cfaf6b7983694aa6b45efff059d2081d47e10088",
        ],
    )

    rpm(
        name = "libXft-0__2.3.4-3.fc37.x86_64",
        sha256 = "a58d00c3352ef14db5c6c69eb30db12aa7c71db10fa783bf9b49681ada35b5f0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a58d00c3352ef14db5c6c69eb30db12aa7c71db10fa783bf9b49681ada35b5f0",
        ],
    )

    rpm(
        name = "libXi-0__1.8-3.fc37.x86_64",
        sha256 = "f62eee892d036e6607e4e204cc2c913c700a68d9eda5633e49198209fa78ca20",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f62eee892d036e6607e4e204cc2c913c700a68d9eda5633e49198209fa78ca20",
        ],
    )

    rpm(
        name = "libXinerama-0__1.1.4-11.fc37.x86_64",
        sha256 = "68f950e6acca043226def1cc1f7310480fc7c897ed1828e739a941a420a8d9b0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/68f950e6acca043226def1cc1f7310480fc7c897ed1828e739a941a420a8d9b0",
        ],
    )

    rpm(
        name = "libXrandr-0__1.5.2-9.fc37.x86_64",
        sha256 = "d852d8dad4ae7e20d1d7f0ce3424634ccccac44c411b85e943e4a26887213f28",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d852d8dad4ae7e20d1d7f0ce3424634ccccac44c411b85e943e4a26887213f28",
        ],
    )

    rpm(
        name = "libXrender-0__0.9.10-17.fc37.x86_64",
        sha256 = "15237bfa9681ca8566949d0d70e1f21fb53ff32f6c49b3bd117b641438af72ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15237bfa9681ca8566949d0d70e1f21fb53ff32f6c49b3bd117b641438af72ac",
        ],
    )

    rpm(
        name = "libXtst-0__1.2.3-17.fc37.x86_64",
        sha256 = "0e47e0a93b1a142868cfd9fb426262fc61b6061b09ce69cc3a7583a4b6faa8e8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e47e0a93b1a142868cfd9fb426262fc61b6061b09ce69cc3a7583a4b6faa8e8",
        ],
    )

    rpm(
        name = "libXv-0__1.0.11-17.fc37.x86_64",
        sha256 = "b895e3520db9e89098498282e282d011b446555da28c143c52ab2fa000cce0e7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b895e3520db9e89098498282e282d011b446555da28c143c52ab2fa000cce0e7",
        ],
    )

    rpm(
        name = "libXxf86vm-0__1.1.4-19.fc37.x86_64",
        sha256 = "2830c91a4ab4371c0e6659ed219da65fb036c3902f4a060fca83b06fae6795a3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2830c91a4ab4371c0e6659ed219da65fb036c3902f4a060fca83b06fae6795a3",
        ],
    )

    rpm(
        name = "libacl-0__2.3.1-4.fc37.x86_64",
        sha256 = "15224cb92199b8011fe47dc12e0bbcdbee0c93e0f29553b3b07ae41768b48ce3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15224cb92199b8011fe47dc12e0bbcdbee0c93e0f29553b3b07ae41768b48ce3",
        ],
    )

    rpm(
        name = "libaio-0__0.3.111-14.fc37.x86_64",
        sha256 = "d4c8bb3a8bb0c529f49ee7fe6c2100674de6b54837aa29bf0a12e08f08575fdd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4c8bb3a8bb0c529f49ee7fe6c2100674de6b54837aa29bf0a12e08f08575fdd",
        ],
    )

    rpm(
        name = "libarchive-0__3.6.1-3.fc37.x86_64",
        sha256 = "a21c75bf1af2f299b06879592d9eb89a20168d3c5306365438e6403e1d1064ce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a21c75bf1af2f299b06879592d9eb89a20168d3c5306365438e6403e1d1064ce",
        ],
    )

    rpm(
        name = "libargon2-0__20190702-1.fc37.x86_64",
        sha256 = "bf280bf9e59891bfcb4a987d5df22d6a6d9f60589dd00b790b5a3047a727a40b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf280bf9e59891bfcb4a987d5df22d6a6d9f60589dd00b790b5a3047a727a40b",
        ],
    )

    rpm(
        name = "libargon2-devel-0__20190702-1.fc37.x86_64",
        sha256 = "0fb322b72ce9a95c09f05a8f94b0bed06c203c4f8a7ffcfd6faeec23134d09e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0fb322b72ce9a95c09f05a8f94b0bed06c203c4f8a7ffcfd6faeec23134d09e3",
        ],
    )

    rpm(
        name = "libasyncns-0__0.8-23.fc37.x86_64",
        sha256 = "b80826e69554a444a8aa187d0ac9f82c7118f02f8aa280729effb80311615af5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b80826e69554a444a8aa187d0ac9f82c7118f02f8aa280729effb80311615af5",
        ],
    )

    rpm(
        name = "libattr-0__2.5.1-5.fc37.x86_64",
        sha256 = "3a423be562953538eaa0d1e78ef35890396cdf1ad89561c619aa72d3a59bfb82",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3a423be562953538eaa0d1e78ef35890396cdf1ad89561c619aa72d3a59bfb82",
        ],
    )

    rpm(
        name = "libb2-0__0.98.1-7.fc37.x86_64",
        sha256 = "da6c0a039fb7e2ce0b324c758757c6482c2683f2ff7bd7f9b06cd625d0fae17a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/da6c0a039fb7e2ce0b324c758757c6482c2683f2ff7bd7f9b06cd625d0fae17a",
        ],
    )

    rpm(
        name = "libbasicobjects-0__0.1.1-52.fc37.x86_64",
        sha256 = "364011baa5e293baae2e60992fd1241f6c822d7c7e48c259d3e6abb0530036af",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/364011baa5e293baae2e60992fd1241f6c822d7c7e48c259d3e6abb0530036af",
        ],
    )

    rpm(
        name = "libblkid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b0388d1a529bf6b54ca648e91529b1e7790e6aaa42e0ac2b7be6640e4f24a21d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0388d1a529bf6b54ca648e91529b1e7790e6aaa42e0ac2b7be6640e4f24a21d",
        ],
    )

    rpm(
        name = "libblkid-devel-0__2.38.1-1.fc37.x86_64",
        sha256 = "fab8830fa105d51f04582d19c6fc841b7f64e1d29e96041e5ceda1fbe211db40",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fab8830fa105d51f04582d19c6fc841b7f64e1d29e96041e5ceda1fbe211db40",
        ],
    )

    rpm(
        name = "libbpf-2__0.8.0-2.fc37.x86_64",
        sha256 = "3722422d69b3fcfc2d1b0e263051aa94c3fed6b89a709b1f1b4ff6627e114c0a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3722422d69b3fcfc2d1b0e263051aa94c3fed6b89a709b1f1b4ff6627e114c0a",
        ],
    )

    rpm(
        name = "libbrotli-0__1.0.9-9.fc37.x86_64",
        sha256 = "2a8b12086461425c602dd65443a6430b11a73580378f43e207f34346162f3050",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2a8b12086461425c602dd65443a6430b11a73580378f43e207f34346162f3050",
        ],
    )

    rpm(
        name = "libcacard-3__2.8.1-3.fc37.x86_64",
        sha256 = "fd8c61ebb2e64564438a89484fe86d48f485c2bf80dadce20659c659e7b41a69",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fd8c61ebb2e64564438a89484fe86d48f485c2bf80dadce20659c659e7b41a69",
        ],
    )

    rpm(
        name = "libcap-0__2.48-5.fc37.x86_64",
        sha256 = "aa22373907b6ff9fa3d2f7d9e33a9bdefc9ac50486f2dac5251ac4e206a8a61d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aa22373907b6ff9fa3d2f7d9e33a9bdefc9ac50486f2dac5251ac4e206a8a61d",
        ],
    )

    rpm(
        name = "libcap-ng-0__0.8.3-3.fc37.x86_64",
        sha256 = "bcca8a17ae16f9f1c8664f9f54e8f2178f028821f6802ebf33cdcd2d4289bf7f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bcca8a17ae16f9f1c8664f9f54e8f2178f028821f6802ebf33cdcd2d4289bf7f",
        ],
    )

    rpm(
        name = "libcloudproviders-0__0.3.1-6.fc37.x86_64",
        sha256 = "f5c756e13527266933c30c9f0cf04572e6b5415d7213d775200dd78e45173274",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f5c756e13527266933c30c9f0cf04572e6b5415d7213d775200dd78e45173274",
        ],
    )

    rpm(
        name = "libcollection-0__0.7.0-52.fc37.x86_64",
        sha256 = "32e67e6a70e493c52515626820cef9d3df6457f111004483fe6a8657255f4ba7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32e67e6a70e493c52515626820cef9d3df6457f111004483fe6a8657255f4ba7",
        ],
    )

    rpm(
        name = "libcom_err-0__1.46.5-3.fc37.x86_64",
        sha256 = "e98643b3299e5a5b9b1e85a0763b567035f1d83164b3b9a4629fd23467667464",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e98643b3299e5a5b9b1e85a0763b567035f1d83164b3b9a4629fd23467667464",
        ],
    )

    rpm(
        name = "libconfig-0__1.7.3-4.fc37.x86_64",
        sha256 = "fddb1f157125626cdb33e56d53e6c167899681f6ea77224951a0ab94a2e83e45",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fddb1f157125626cdb33e56d53e6c167899681f6ea77224951a0ab94a2e83e45",
        ],
    )

    rpm(
        name = "libcurl-0__7.85.0-8.fc37.x86_64",
        sha256 = "43e8a4e927d64daa9733f6c761a66715849ac3da3a9aed9eefdfcc2228510a5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/43e8a4e927d64daa9733f6c761a66715849ac3da3a9aed9eefdfcc2228510a5f",
        ],
    )

    rpm(
        name = "libcurl-minimal-0__7.85.0-8.fc37.x86_64",
        sha256 = "bb0460b195694c78a58e83ab54268a41cc10f9655ac465d4d0588a5c19a35ab1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bb0460b195694c78a58e83ab54268a41cc10f9655ac465d4d0588a5c19a35ab1",
        ],
    )

    rpm(
        name = "libdatrie-0__0.2.13-4.fc37.x86_64",
        sha256 = "8229f330ea5e220b5ddb7b047b6ec998ee4a4d62002568dbfc9f8c4ac9341ec0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8229f330ea5e220b5ddb7b047b6ec998ee4a4d62002568dbfc9f8c4ac9341ec0",
        ],
    )

    rpm(
        name = "libdb-0__5.3.28-53.fc37.x86_64",
        sha256 = "e89a4a620d5531f30b895694134a982fa37615b3f61c59a21ede6e64a096c5cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e89a4a620d5531f30b895694134a982fa37615b3f61c59a21ede6e64a096c5cd",
        ],
    )

    rpm(
        name = "libdrm-0__2.4.114-1.fc37.x86_64",
        sha256 = "955093d44c37476f34a65d9b13c36aec45bfd3e366ad224581547b13e97bd3a9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/955093d44c37476f34a65d9b13c36aec45bfd3e366ad224581547b13e97bd3a9",
        ],
    )

    rpm(
        name = "libeconf-0__0.4.0-4.fc37.x86_64",
        sha256 = "f0cc1addee779f09aade289e3be4e9bd103a274a6bdf11f8331878686f432653",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f0cc1addee779f09aade289e3be4e9bd103a274a6bdf11f8331878686f432653",
        ],
    )

    rpm(
        name = "libedit-0__3.1-43.20221009cvs.fc37.x86_64",
        sha256 = "7e128e732af0a53585a9cdae8975c423f3079f57ada4374b4d797ef51cae9ce7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e128e732af0a53585a9cdae8975c423f3079f57ada4374b4d797ef51cae9ce7",
        ],
    )

    rpm(
        name = "libepoxy-0__1.5.10-2.fc37.x86_64",
        sha256 = "06e9ae0b5d260cd66c21b663eecfef618c2fdd8aa2de314793d96c822b09a183",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/06e9ae0b5d260cd66c21b663eecfef618c2fdd8aa2de314793d96c822b09a183",
        ],
    )

    rpm(
        name = "libevent-0__2.1.12-7.fc37.x86_64",
        sha256 = "eac9405b6177c4778d772b61ef03a5cd571e2ce6ea337929a1e8a10e80422ba7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eac9405b6177c4778d772b61ef03a5cd571e2ce6ea337929a1e8a10e80422ba7",
        ],
    )

    rpm(
        name = "libfdisk-0__2.38.1-1.fc37.x86_64",
        sha256 = "7a4bd1f4975a52fc201c9bc978f155dcb97212cb970210525d903b03644a713d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7a4bd1f4975a52fc201c9bc978f155dcb97212cb970210525d903b03644a713d",
        ],
    )

    rpm(
        name = "libfdt-0__1.6.1-5.fc37.x86_64",
        sha256 = "4a7f47d967d7884439590b5b1a9145dc68e15a201f648b57e3b5f38ada09ea9c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a7f47d967d7884439590b5b1a9145dc68e15a201f648b57e3b5f38ada09ea9c",
        ],
    )

    rpm(
        name = "libffado-0__2.4.7-1.fc37.x86_64",
        sha256 = "08261c7e212c5758b39f2220175e1b1ee05b7030459d029e5b1fcdd435321025",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/08261c7e212c5758b39f2220175e1b1ee05b7030459d029e5b1fcdd435321025",
        ],
    )

    rpm(
        name = "libffi-0__3.4.4-1.fc37.x86_64",
        sha256 = "66bae5662d9287e769f5d8b7f723d45eb19f2902d912be40bf9e5dd8d5c68067",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66bae5662d9287e769f5d8b7f723d45eb19f2902d912be40bf9e5dd8d5c68067",
        ],
    )

    rpm(
        name = "libgcc-0__12.2.1-4.fc37.x86_64",
        sha256 = "25299b673e7488f538c6d0433ea7fe0ffc8311e41dd7115b5985145e493e4b05",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/25299b673e7488f538c6d0433ea7fe0ffc8311e41dd7115b5985145e493e4b05",
        ],
    )

    rpm(
        name = "libgcrypt-0__1.10.1-4.fc37.x86_64",
        sha256 = "ca802ad5d10b2728ba10bf98bb16796585d69ec775f5452b3a43718e07c4667a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ca802ad5d10b2728ba10bf98bb16796585d69ec775f5452b3a43718e07c4667a",
        ],
    )

    rpm(
        name = "libgfapi0-0__10.3-1.fc37.x86_64",
        sha256 = "ae0f00c055fc9d9da0190ef4f3229f870380cf9cce1a0c78b842b99120b1f020",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ae0f00c055fc9d9da0190ef4f3229f870380cf9cce1a0c78b842b99120b1f020",
        ],
    )

    rpm(
        name = "libgfrpc0-0__10.3-1.fc37.x86_64",
        sha256 = "bf19b3e14fbbe4b93a4c8c852406d6cc58c225f5732432a4998eb01008657ad8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf19b3e14fbbe4b93a4c8c852406d6cc58c225f5732432a4998eb01008657ad8",
        ],
    )

    rpm(
        name = "libgfxdr0-0__10.3-1.fc37.x86_64",
        sha256 = "619c7bfecd84efec749336e35d9479c8deea133dfb56fef31bb4674f80c093e9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/619c7bfecd84efec749336e35d9479c8deea133dfb56fef31bb4674f80c093e9",
        ],
    )

    rpm(
        name = "libglusterd0-0__10.3-1.fc37.x86_64",
        sha256 = "da498d0c5c617df26aab01b06e7b2db213bc7353ed5704600e63e30f165b915f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/da498d0c5c617df26aab01b06e7b2db213bc7353ed5704600e63e30f165b915f",
        ],
    )

    rpm(
        name = "libglusterfs0-0__10.3-1.fc37.x86_64",
        sha256 = "bf043596ba81d4062322119ed93ec3e0aaa1e9eb87b4aedf2525f273d6b9d0c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf043596ba81d4062322119ed93ec3e0aaa1e9eb87b4aedf2525f273d6b9d0c2",
        ],
    )

    rpm(
        name = "libglvnd-1__1.5.0-1.fc37.x86_64",
        sha256 = "f5fdd595b4aad94b00695cbe1fded7306588e1c5407fe8ff048e47ea7bfff819",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f5fdd595b4aad94b00695cbe1fded7306588e1c5407fe8ff048e47ea7bfff819",
        ],
    )

    rpm(
        name = "libglvnd-egl-1__1.5.0-1.fc37.x86_64",
        sha256 = "e6e8511c0b2fbee6036fab88778ea75ca1e3d17e813043a80fbd02a614aa9702",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e6e8511c0b2fbee6036fab88778ea75ca1e3d17e813043a80fbd02a614aa9702",
        ],
    )

    rpm(
        name = "libglvnd-glx-1__1.5.0-1.fc37.x86_64",
        sha256 = "5faee12d35d66511484d563b996eea8cdc539de63b008e0193e99bb199ac8506",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5faee12d35d66511484d563b996eea8cdc539de63b008e0193e99bb199ac8506",
        ],
    )

    rpm(
        name = "libgomp-0__12.2.1-4.fc37.x86_64",
        sha256 = "3f2da924fd5168b4f31f56895eb80691778319bf85e408ff02a5ac6714f02f50",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3f2da924fd5168b4f31f56895eb80691778319bf85e408ff02a5ac6714f02f50",
        ],
    )

    rpm(
        name = "libgpg-error-0__1.46-1.fc37.x86_64",
        sha256 = "bfa65a9946b2547110994855d168e4434313ad26280cb935c19bb88d2af283d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bfa65a9946b2547110994855d168e4434313ad26280cb935c19bb88d2af283d2",
        ],
    )

    rpm(
        name = "libgudev-0__237-3.fc37.x86_64",
        sha256 = "0bd6b3f97c370399de8aad7e37d1195d3be5f3aa66c3ab4c83eeb42e4cde23ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0bd6b3f97c370399de8aad7e37d1195d3be5f3aa66c3ab4c83eeb42e4cde23ac",
        ],
    )

    rpm(
        name = "libgusb-0__0.4.5-1.fc37.x86_64",
        sha256 = "72a3402b893b97af55f8f96668967ea02c3eae5b1910f16d71af4893e1072584",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/72a3402b893b97af55f8f96668967ea02c3eae5b1910f16d71af4893e1072584",
        ],
    )

    rpm(
        name = "libibverbs-0__41.0-1.fc37.x86_64",
        sha256 = "58fc922b01b99cf99809121ca2d3134853a5cc06ec5b8b5f6a0de7eec5c12202",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/58fc922b01b99cf99809121ca2d3134853a5cc06ec5b8b5f6a0de7eec5c12202",
        ],
    )

    rpm(
        name = "libicu-0__71.1-2.fc37.x86_64",
        sha256 = "d1a35de9152869803d6667926c330f7e0bffdb52c0d513ee0f3ecdc25f289aff",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d1a35de9152869803d6667926c330f7e0bffdb52c0d513ee0f3ecdc25f289aff",
        ],
    )

    rpm(
        name = "libidn2-0__2.3.4-1.fc37.x86_64",
        sha256 = "e32e2ab71cfb0bedb84611251987db7acdf665917864be335d0786ea6bbd02b4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e32e2ab71cfb0bedb84611251987db7acdf665917864be335d0786ea6bbd02b4",
        ],
    )

    rpm(
        name = "libiec61883-0__1.2.0-30.fc37.x86_64",
        sha256 = "2d330c8128822a82a121b1e80aa1743a79f56dcebd19da170c158542706febe7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2d330c8128822a82a121b1e80aa1743a79f56dcebd19da170c158542706febe7",
        ],
    )

    rpm(
        name = "libini_config-0__1.3.1-52.fc37.x86_64",
        sha256 = "adf559f274af1e1a4facb1cc93e3840e83ba429513df55bdb79baa6508423e42",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/adf559f274af1e1a4facb1cc93e3840e83ba429513df55bdb79baa6508423e42",
        ],
    )

    rpm(
        name = "libiscsi-0__1.19.0-6.fc37.x86_64",
        sha256 = "dc07d9cc6559510f6e7100894a075b061e99caa9d6bb79532f75f007dc9b4f6d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc07d9cc6559510f6e7100894a075b061e99caa9d6bb79532f75f007dc9b4f6d",
        ],
    )

    rpm(
        name = "libjpeg-turbo-0__2.1.3-2.fc37.x86_64",
        sha256 = "a7934e081a697ca28f8ff83b973b33299c7127cdec8d4102128ab9ea69d172f4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a7934e081a697ca28f8ff83b973b33299c7127cdec8d4102128ab9ea69d172f4",
        ],
    )

    rpm(
        name = "libjxl-1__0.7.0-5.fc37.x86_64",
        sha256 = "61f50efc5bbc076ac1c4a179b6a20905cf911174453eb15a07ec8ffd451c4f20",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/61f50efc5bbc076ac1c4a179b6a20905cf911174453eb15a07ec8ffd451c4f20",
        ],
    )

    rpm(
        name = "libmnl-0__1.0.5-1.fc37.x86_64",
        sha256 = "d6aa832d8cc2c70fe044fec76c5769f282a0c2c92749591e96b7d2e6053393eb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d6aa832d8cc2c70fe044fec76c5769f282a0c2c92749591e96b7d2e6053393eb",
        ],
    )

    rpm(
        name = "libmount-0__2.38.1-1.fc37.x86_64",
        sha256 = "50c304faa94d7959e5cbc0642b3c77539ad000042e6617ea5da4789c8105496f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/50c304faa94d7959e5cbc0642b3c77539ad000042e6617ea5da4789c8105496f",
        ],
    )

    rpm(
        name = "libnetfilter_conntrack-0__1.0.8-5.fc37.x86_64",
        sha256 = "7e5767c70bc9f6429b6e9a66ce2cb75e90eea0d9a48392e1e9eb54480fcabd8b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e5767c70bc9f6429b6e9a66ce2cb75e90eea0d9a48392e1e9eb54480fcabd8b",
        ],
    )

    rpm(
        name = "libnfnetlink-0__1.0.1-22.fc37.x86_64",
        sha256 = "44e396c693edeea0a6db316428af4bd7b2c0df41ce562e2f951b871b70a10bc9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/44e396c693edeea0a6db316428af4bd7b2c0df41ce562e2f951b871b70a10bc9",
        ],
    )

    rpm(
        name = "libnfs-0__4.0.0-7.fc37.x86_64",
        sha256 = "2323edb5a7252d2c3c91182a8f329718e99ebc90f046f1a00960da3bcc87afc4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2323edb5a7252d2c3c91182a8f329718e99ebc90f046f1a00960da3bcc87afc4",
        ],
    )

    rpm(
        name = "libnfsidmap-1__2.6.2-2.rc6.fc37.x86_64",
        sha256 = "ee4bbf5b13c396db622de69fcf90a979366c290b81c3581a97240ca38c718120",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee4bbf5b13c396db622de69fcf90a979366c290b81c3581a97240ca38c718120",
        ],
    )

    rpm(
        name = "libnghttp2-0__1.51.0-1.fc37.x86_64",
        sha256 = "42fbaaacbeb241755d8448dd5672bbbcc48cbe9548c095ce0efef4140bc12520",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42fbaaacbeb241755d8448dd5672bbbcc48cbe9548c095ce0efef4140bc12520",
        ],
    )

    rpm(
        name = "libnl3-0__3.7.0-2.fc37.x86_64",
        sha256 = "4543c991e6f536468d9d47527a201b58b9bc049364a6bdfe15a2f910a02e68f6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4543c991e6f536468d9d47527a201b58b9bc049364a6bdfe15a2f910a02e68f6",
        ],
    )

    rpm(
        name = "libnsl2-0__2.0.0-4.fc37.x86_64",
        sha256 = "a1e9428515b0df1c2a423ad3c35bcdf93333172fe346169bb3018a882e27be5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a1e9428515b0df1c2a423ad3c35bcdf93333172fe346169bb3018a882e27be5f",
        ],
    )

    rpm(
        name = "libogg-2__1.3.5-4.fc37.x86_64",
        sha256 = "4716db0b1017a05751908d7376357867c843341068e6f01000bcdc11d4889df9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4716db0b1017a05751908d7376357867c843341068e6f01000bcdc11d4889df9",
        ],
    )

    rpm(
        name = "libpath_utils-0__0.2.1-52.fc37.x86_64",
        sha256 = "932abcd232d522bb508daa2c9e6f4a3ea79039ad80a6478e6384d517c1d6698c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/932abcd232d522bb508daa2c9e6f4a3ea79039ad80a6478e6384d517c1d6698c",
        ],
    )

    rpm(
        name = "libpcap-14__1.10.3-1.fc37.x86_64",
        sha256 = "6fd955a6637e2998476cf1a9ccde0af350b1a8931ec5b58efed5e86c38af41f8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6fd955a6637e2998476cf1a9ccde0af350b1a8931ec5b58efed5e86c38af41f8",
        ],
    )

    rpm(
        name = "libpciaccess-0__0.16-7.fc37.x86_64",
        sha256 = "a9f1fa38572b5e3ae412dc4e822020e2290c7092cef120a0f08a1fcc62600ce3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9f1fa38572b5e3ae412dc4e822020e2290c7092cef120a0f08a1fcc62600ce3",
        ],
    )

    rpm(
        name = "libpkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "ecd52fd3f3065606ba5164249b29c837cbd172643d13a00a1a72fc657b115af7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ecd52fd3f3065606ba5164249b29c837cbd172643d13a00a1a72fc657b115af7",
        ],
    )

    rpm(
        name = "libpmem-0__1.12.0-1.fc37.x86_64",
        sha256 = "af3c9045089110c849ee3191aa506e16f4a69dc625b77abf09583de237df41df",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af3c9045089110c849ee3191aa506e16f4a69dc625b77abf09583de237df41df",
        ],
    )

    rpm(
        name = "libpmemobj-0__1.12.0-1.fc37.x86_64",
        sha256 = "e2f17f151f862a607d86dc17455b0e7f075a8152470788baad3ae0e449dc86ff",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e2f17f151f862a607d86dc17455b0e7f075a8152470788baad3ae0e449dc86ff",
        ],
    )

    rpm(
        name = "libpng-2__1.6.37-13.fc37.x86_64",
        sha256 = "49a024d34e3c531516562bc51b749dee540db4d34486a95cdd8d85300b7de455",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49a024d34e3c531516562bc51b749dee540db4d34486a95cdd8d85300b7de455",
        ],
    )

    rpm(
        name = "libpsl-0__0.21.1-6.fc37.x86_64",
        sha256 = "90801f2f5ce98f2ba06f659b4676cb55d39f8e597a8f2da3e59dc943abe8f5a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/90801f2f5ce98f2ba06f659b4676cb55d39f8e597a8f2da3e59dc943abe8f5a6",
        ],
    )

    rpm(
        name = "libpwquality-0__1.4.5-3.fc37.x86_64",
        sha256 = "a9019a471496fdada529757331ec004397db7a0c4347531bd639c127bbaf8300",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9019a471496fdada529757331ec004397db7a0c4347531bd639c127bbaf8300",
        ],
    )

    rpm(
        name = "libqb-0__2.0.6-3.fc37.x86_64",
        sha256 = "3e05e7df7504725e4c7c168ab9b44e4e7663c8bcaf56fca3196baa11e622a386",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3e05e7df7504725e4c7c168ab9b44e4e7663c8bcaf56fca3196baa11e622a386",
        ],
    )

    rpm(
        name = "librados2-2__17.2.5-1.fc37.x86_64",
        sha256 = "832b790283724be8f35f238ca26401d6bf5059b4a945994c9836f1658afc936c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/832b790283724be8f35f238ca26401d6bf5059b4a945994c9836f1658afc936c",
        ],
    )

    rpm(
        name = "libraw1394-0__2.1.2-16.fc37.x86_64",
        sha256 = "86625b3a89c2cc3ec6f3781ae426ccc207554cfa35ea19c09e1f77604a0a51ad",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86625b3a89c2cc3ec6f3781ae426ccc207554cfa35ea19c09e1f77604a0a51ad",
        ],
    )

    rpm(
        name = "librbd1-2__17.2.5-1.fc37.x86_64",
        sha256 = "cb18c01c37925cd69adb5fe6ddc72f975a4d90326ffa9bf8def04cf3d48c109d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cb18c01c37925cd69adb5fe6ddc72f975a4d90326ffa9bf8def04cf3d48c109d",
        ],
    )

    rpm(
        name = "librdmacm-0__41.0-1.fc37.x86_64",
        sha256 = "65edde85818fd605392a74c235e0be8d9bb7032b8bdccd89d3fe34ae2e4a1e7b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/65edde85818fd605392a74c235e0be8d9bb7032b8bdccd89d3fe34ae2e4a1e7b",
        ],
    )

    rpm(
        name = "libref_array-0__0.1.5-52.fc37.x86_64",
        sha256 = "69ca0883e3929168bc8eb555463ae051fc38afedae0fa98d566bc22388304a91",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/69ca0883e3929168bc8eb555463ae051fc38afedae0fa98d566bc22388304a91",
        ],
    )

    rpm(
        name = "libsamplerate-0__0.2.2-3.fc37.x86_64",
        sha256 = "42f6179984d7be1ff7019cb0a73511ddad62a1019bb09f14fdc67943187da67e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42f6179984d7be1ff7019cb0a73511ddad62a1019bb09f14fdc67943187da67e",
        ],
    )

    rpm(
        name = "libseccomp-0__2.5.3-3.fc37.x86_64",
        sha256 = "017877a97c8222fc7eca7fab77600a3a1fcdec92f9dd39d8df6e64726909fcbe",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/017877a97c8222fc7eca7fab77600a3a1fcdec92f9dd39d8df6e64726909fcbe",
        ],
    )

    rpm(
        name = "libselinux-0__3.5-1.fc37.x86_64",
        sha256 = "43d73a574c3c0838d213c4d5f038766d41e4eb6930c68b09db53d65c30c2de1d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/43d73a574c3c0838d213c4d5f038766d41e4eb6930c68b09db53d65c30c2de1d",
        ],
    )

    rpm(
        name = "libselinux-devel-0__3.5-1.fc37.x86_64",
        sha256 = "d737a3768ad96bde07a1ed0d5cf01d252ce6a6b56636121ab6edd23cd2345a33",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d737a3768ad96bde07a1ed0d5cf01d252ce6a6b56636121ab6edd23cd2345a33",
        ],
    )

    rpm(
        name = "libselinux-utils-0__3.5-1.fc37.x86_64",
        sha256 = "723efbdc421150c13f6a2fe47e3d2587f83a26bfae8561e3361985793762b05d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/723efbdc421150c13f6a2fe47e3d2587f83a26bfae8561e3361985793762b05d",
        ],
    )

    rpm(
        name = "libsemanage-0__3.5-1.fc37.x86_64",
        sha256 = "aeb55e09d224bd6212a8456160cabbcfac61eb2a792a572c01567dba9529c208",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aeb55e09d224bd6212a8456160cabbcfac61eb2a792a572c01567dba9529c208",
        ],
    )

    rpm(
        name = "libsepol-0__3.5-1.fc37.x86_64",
        sha256 = "2cdfb41068ac6e211652b3e2ed88c16d606e7374f9f52dfd1248981101501299",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2cdfb41068ac6e211652b3e2ed88c16d606e7374f9f52dfd1248981101501299",
        ],
    )

    rpm(
        name = "libsepol-devel-0__3.5-1.fc37.x86_64",
        sha256 = "29d79933c62f1fd70c9e4cddc885d6814049bd29775fad970ed73531a27c7117",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/29d79933c62f1fd70c9e4cddc885d6814049bd29775fad970ed73531a27c7117",
        ],
    )

    rpm(
        name = "libsigc__plus____plus__20-0__2.10.8-2.fc37.x86_64",
        sha256 = "1b87303f104d1de4a3ccd5179dfd7f25916c734aa4f80cff312a9c495d994912",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1b87303f104d1de4a3ccd5179dfd7f25916c734aa4f80cff312a9c495d994912",
        ],
    )

    rpm(
        name = "libsigsegv-0__2.14-3.fc37.x86_64",
        sha256 = "0f038b70d155dae3df4824776c5a135f02c423c688b9486d4f84eb6a16a90494",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f038b70d155dae3df4824776c5a135f02c423c688b9486d4f84eb6a16a90494",
        ],
    )

    rpm(
        name = "libslirp-0__4.7.0-2.fc37.x86_64",
        sha256 = "775c3da0e9e2961262c552292f98dc80f4b05bde52573539148ff4cb51459a51",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/775c3da0e9e2961262c552292f98dc80f4b05bde52573539148ff4cb51459a51",
        ],
    )

    rpm(
        name = "libsmartcols-0__2.38.1-1.fc37.x86_64",
        sha256 = "93246c002aefec27bb398aa3397ae555bcc3035b10aebb4937c4bea9268bacf1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93246c002aefec27bb398aa3397ae555bcc3035b10aebb4937c4bea9268bacf1",
        ],
    )

    rpm(
        name = "libsndfile-0__1.1.0-4.fc37.x86_64",
        sha256 = "73aeba406a3c76a62ced76a918efbc8338bd4304d672ab9db38ed057238b9eca",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/73aeba406a3c76a62ced76a918efbc8338bd4304d672ab9db38ed057238b9eca",
        ],
    )

    rpm(
        name = "libsoup3-0__3.2.2-2.fc37.x86_64",
        sha256 = "9575393975efc3087b98ef56ed5aaa0e771edc99c1571147efe3218f21e711ed",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9575393975efc3087b98ef56ed5aaa0e771edc99c1571147efe3218f21e711ed",
        ],
    )

    rpm(
        name = "libssh-0__0.10.4-2.fc37.x86_64",
        sha256 = "cdee8c9676d686a0df90d27b4863f15e871dc58363eb2f11f5e69fe3e9a23c85",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cdee8c9676d686a0df90d27b4863f15e871dc58363eb2f11f5e69fe3e9a23c85",
        ],
    )

    rpm(
        name = "libssh-config-0__0.10.4-2.fc37.x86_64",
        sha256 = "d17d16ca2e2a42035778094bca077ba675e440911c5546f99a274278eb32e0d7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d17d16ca2e2a42035778094bca077ba675e440911c5546f99a274278eb32e0d7",
        ],
    )

    rpm(
        name = "libssh2-0__1.10.0-5.fc37.x86_64",
        sha256 = "8a6e1ce9b08e054746ef1ee0eb51a70bc06ee1bea3ae24bf5942a191ea12ab3a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8a6e1ce9b08e054746ef1ee0eb51a70bc06ee1bea3ae24bf5942a191ea12ab3a",
        ],
    )

    rpm(
        name = "libstdc__plus____plus__-0__12.2.1-4.fc37.x86_64",
        sha256 = "ba8009388d86fbb92deff293e04eb57ca9c3b3ba41994932b3e4226533ffb575",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ba8009388d86fbb92deff293e04eb57ca9c3b3ba41994932b3e4226533ffb575",
        ],
    )

    rpm(
        name = "libstemmer-0__0-19.585svn.fc37.x86_64",
        sha256 = "93655f2930a2304035794b97a6dc150c8c8eae4e3b8ec08f87a60f39af0b2e1c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93655f2930a2304035794b97a6dc150c8c8eae4e3b8ec08f87a60f39af0b2e1c",
        ],
    )

    rpm(
        name = "libtasn1-0__4.19.0-1.fc37.x86_64",
        sha256 = "35b51a0796af6930b2a8a511df8c51938006cfcfdf74ddfe6482eb9febd87dfa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/35b51a0796af6930b2a8a511df8c51938006cfcfdf74ddfe6482eb9febd87dfa",
        ],
    )

    rpm(
        name = "libthai-0__0.1.29-3.fc37.x86_64",
        sha256 = "ff80496b4e4e6ae3cd731203c1acccc618b6e834219e66eb19809793e828fa90",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ff80496b4e4e6ae3cd731203c1acccc618b6e834219e66eb19809793e828fa90",
        ],
    )

    rpm(
        name = "libtheora-1__1.1.1-32.fc37.x86_64",
        sha256 = "441c94bbcbaa622e6e334bd954e6ac6fbaae840493aeacd1c54fa1b448ab9169",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/441c94bbcbaa622e6e334bd954e6ac6fbaae840493aeacd1c54fa1b448ab9169",
        ],
    )

    rpm(
        name = "libtiff-0__4.4.0-4.fc37.x86_64",
        sha256 = "3d9ddffd9b6665d98f13a51f3a3251b9ca14bad29e8d2414b45bb0785f1d126e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3d9ddffd9b6665d98f13a51f3a3251b9ca14bad29e8d2414b45bb0785f1d126e",
        ],
    )

    rpm(
        name = "libtirpc-0__1.3.3-0.fc37.x86_64",
        sha256 = "76dcdfd95452e176f64d6008d114e9415cd8384c5c0d3300fe644c137b6917fa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76dcdfd95452e176f64d6008d114e9415cd8384c5c0d3300fe644c137b6917fa",
        ],
    )

    rpm(
        name = "libtpms-0__0.9.6-1.fc37.x86_64",
        sha256 = "d61fd47b4126e4d89d425af91a224718493b7c7c4eab148508699d577d191dbe",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d61fd47b4126e4d89d425af91a224718493b7c7c4eab148508699d577d191dbe",
        ],
    )

    rpm(
        name = "libtracker-sparql-0__3.4.2-1.fc37.x86_64",
        sha256 = "da6c84f5df11f731226c479bd4b66ea0b4b264b7810a9c65efc5c6490885e3e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/da6c84f5df11f731226c479bd4b66ea0b4b264b7810a9c65efc5c6490885e3e3",
        ],
    )

    rpm(
        name = "libunistring-0__1.0-2.fc37.x86_64",
        sha256 = "acb031577655bba5a41c1fb0ec954bb84e207f9e2d08b2cdb3d4e2b7806b0670",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/acb031577655bba5a41c1fb0ec954bb84e207f9e2d08b2cdb3d4e2b7806b0670",
        ],
    )

    rpm(
        name = "libunwind-0__1.6.2-5.fc37.x86_64",
        sha256 = "fe7cffb9387748b166437a143c886f92f012aff6c16a9604fc1bab3fcb0be928",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe7cffb9387748b166437a143c886f92f012aff6c16a9604fc1bab3fcb0be928",
        ],
    )

    rpm(
        name = "liburing-0__2.3-1.fc37.x86_64",
        sha256 = "b7ace7323804d803c8b8c33fc699b9346924fb80841a701471f29e29abeb255b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7ace7323804d803c8b8c33fc699b9346924fb80841a701471f29e29abeb255b",
        ],
    )

    rpm(
        name = "libusb1-0__1.0.25-9.fc37.x86_64",
        sha256 = "012a7fbcbffc0c6d9a7101fb29eef81cd7ba18d3bb427b3aa0c9ae3a27ba2b2e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/012a7fbcbffc0c6d9a7101fb29eef81cd7ba18d3bb427b3aa0c9ae3a27ba2b2e",
        ],
    )

    rpm(
        name = "libutempter-0__1.2.1-7.fc37.x86_64",
        sha256 = "8fc30b0742e939954d6aebd45364dcd1dbb8b9c85e75c799301c3507e22ea56a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fc30b0742e939954d6aebd45364dcd1dbb8b9c85e75c799301c3507e22ea56a",
        ],
    )

    rpm(
        name = "libuuid-0__2.38.1-1.fc37.x86_64",
        sha256 = "b054577d98aa9615fe459abec31be46b19ad72e0da620d8d251b4449a6db020d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b054577d98aa9615fe459abec31be46b19ad72e0da620d8d251b4449a6db020d",
        ],
    )

    rpm(
        name = "libuuid-devel-0__2.38.1-1.fc37.x86_64",
        sha256 = "a42450ad26785144969fd5faab10f6a382d13e3db2e7b130a33e8a3b314d5d3f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a42450ad26785144969fd5faab10f6a382d13e3db2e7b130a33e8a3b314d5d3f",
        ],
    )

    rpm(
        name = "libverto-0__0.3.2-4.fc37.x86_64",
        sha256 = "ca47b52e1ecd8a2ac6eda368d985390816fbb447f43135ec0ba105165997817f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ca47b52e1ecd8a2ac6eda368d985390816fbb447f43135ec0ba105165997817f",
        ],
    )

    rpm(
        name = "libverto-libevent-0__0.3.2-4.fc37.x86_64",
        sha256 = "2e8f50494b876f4aad8914a20e757ff583aac7cf838af138c10ab9847cf2dea3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2e8f50494b876f4aad8914a20e757ff583aac7cf838af138c10ab9847cf2dea3",
        ],
    )

    rpm(
        name = "libvirt-client-0__8.6.0-5.fc37.x86_64",
        sha256 = "0b807117b6f75cea49140334bc5d66a9a148581f487c1d9c45dca699562a2b02",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0b807117b6f75cea49140334bc5d66a9a148581f487c1d9c45dca699562a2b02",
        ],
    )

    rpm(
        name = "libvirt-daemon-0__8.6.0-5.fc37.x86_64",
        sha256 = "2d6cea47afc5a2e4de3b20f2c970eb09d4f7c2e4fb66e8059c6ecb9e8d83e67c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2d6cea47afc5a2e4de3b20f2c970eb09d4f7c2e4fb66e8059c6ecb9e8d83e67c",
        ],
    )

    rpm(
        name = "libvirt-daemon-config-network-0__8.6.0-5.fc37.x86_64",
        sha256 = "f8f958fbdf0736e399917b9dcd31aee6092e2a82d5c0a790f02abcdd4042b940",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f8f958fbdf0736e399917b9dcd31aee6092e2a82d5c0a790f02abcdd4042b940",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-interface-0__8.6.0-5.fc37.x86_64",
        sha256 = "4be78cd96af8edc1608f8a8f0ce24708c3a8a2c7f5371788477912183fc482cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4be78cd96af8edc1608f8a8f0ce24708c3a8a2c7f5371788477912183fc482cd",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-network-0__8.6.0-5.fc37.x86_64",
        sha256 = "e0804cc66206635c07f8f6527fef0ccb1b069f503dc4e69fa32816f59cdb2e80",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0804cc66206635c07f8f6527fef0ccb1b069f503dc4e69fa32816f59cdb2e80",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-nodedev-0__8.6.0-5.fc37.x86_64",
        sha256 = "fe3a382b99273cd3637a119acc899ca5cdd05a97f8d6d290d9125796839ed9b0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe3a382b99273cd3637a119acc899ca5cdd05a97f8d6d290d9125796839ed9b0",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-nwfilter-0__8.6.0-5.fc37.x86_64",
        sha256 = "9bc95efdc5db11c865924f279d9b89b2787f16c5932ffa5854d4233cca29401e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9bc95efdc5db11c865924f279d9b89b2787f16c5932ffa5854d4233cca29401e",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-qemu-0__8.6.0-5.fc37.x86_64",
        sha256 = "b9d05b4e86818cad62a60f271127375868799819ad3bdf1e5c94dc6b2b7d2e5d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b9d05b4e86818cad62a60f271127375868799819ad3bdf1e5c94dc6b2b7d2e5d",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-secret-0__8.6.0-5.fc37.x86_64",
        sha256 = "f87b6ddd2c9d33155649bc093355160cebe7a0fdf1e4bc3b11505d8405a07cc8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f87b6ddd2c9d33155649bc093355160cebe7a0fdf1e4bc3b11505d8405a07cc8",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-0__8.6.0-5.fc37.x86_64",
        sha256 = "968f3be3cba3dd27e9dc9c758aa18f65e6d5ba0be992fe160b2bb06a41e59a3e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/968f3be3cba3dd27e9dc9c758aa18f65e6d5ba0be992fe160b2bb06a41e59a3e",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-core-0__8.6.0-5.fc37.x86_64",
        sha256 = "eb0ddab88e7e366cf98b9c2ec1fdd2a47d41e798d232b0ae1a36f3639e91b9cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eb0ddab88e7e366cf98b9c2ec1fdd2a47d41e798d232b0ae1a36f3639e91b9cd",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-disk-0__8.6.0-5.fc37.x86_64",
        sha256 = "32efc75bc453a73b8779decf5ca167098133964791742c0e623003f164371d69",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32efc75bc453a73b8779decf5ca167098133964791742c0e623003f164371d69",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-gluster-0__8.6.0-5.fc37.x86_64",
        sha256 = "68f2be95c38bfe4922beb17185ead94cde5d3814b0f4c98fa9a8293d3d58d586",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/68f2be95c38bfe4922beb17185ead94cde5d3814b0f4c98fa9a8293d3d58d586",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-iscsi-0__8.6.0-5.fc37.x86_64",
        sha256 = "8b86462322418ff96c97b3749021107344364ce2adba1b64827c0b11c7c0936c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8b86462322418ff96c97b3749021107344364ce2adba1b64827c0b11c7c0936c",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-iscsi-direct-0__8.6.0-5.fc37.x86_64",
        sha256 = "87687e115b971719067192f4709be83702caa6898acd07f40b00f195cdb51ef4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/87687e115b971719067192f4709be83702caa6898acd07f40b00f195cdb51ef4",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-logical-0__8.6.0-5.fc37.x86_64",
        sha256 = "c36c2ed8110b34bbbfb0b1fb40a2e20b85622b9ad6ad727c3a4647d0bda814ee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c36c2ed8110b34bbbfb0b1fb40a2e20b85622b9ad6ad727c3a4647d0bda814ee",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-mpath-0__8.6.0-5.fc37.x86_64",
        sha256 = "31f24a03a2d2a1a7a7bf68308b5b550f4a5836f340d4a6ba143101d47ffda9a8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/31f24a03a2d2a1a7a7bf68308b5b550f4a5836f340d4a6ba143101d47ffda9a8",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-rbd-0__8.6.0-5.fc37.x86_64",
        sha256 = "700951896e457b5d3c073b925dced22a604b3cf6ff265d072991a625de957781",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/700951896e457b5d3c073b925dced22a604b3cf6ff265d072991a625de957781",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-scsi-0__8.6.0-5.fc37.x86_64",
        sha256 = "16c7f3af89808e64b3c17d9a8198161a981924dfa4b4d70b7a16e12123bd9eb0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/16c7f3af89808e64b3c17d9a8198161a981924dfa4b4d70b7a16e12123bd9eb0",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-sheepdog-0__8.6.0-5.fc37.x86_64",
        sha256 = "b0625cd8c1a54a758101b2f333906c075d178f5b4e0f21cccf63ea6152b225e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0625cd8c1a54a758101b2f333906c075d178f5b4e0f21cccf63ea6152b225e3",
        ],
    )

    rpm(
        name = "libvirt-daemon-driver-storage-zfs-0__8.6.0-5.fc37.x86_64",
        sha256 = "3fc1bb30830cf16056bccebd38e5991e2da0a1a60653d0cd2a304699aed60f6d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3fc1bb30830cf16056bccebd38e5991e2da0a1a60653d0cd2a304699aed60f6d",
        ],
    )

    rpm(
        name = "libvirt-daemon-kvm-0__8.6.0-5.fc37.x86_64",
        sha256 = "fdd1eadf194d9dbdc06e4a3053dfd1501952afab4f5bcaf11bafb2a93b9bef88",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fdd1eadf194d9dbdc06e4a3053dfd1501952afab4f5bcaf11bafb2a93b9bef88",
        ],
    )

    rpm(
        name = "libvirt-devel-0__8.6.0-5.fc37.x86_64",
        sha256 = "3cf11a2799e69f36ee6ee9015549b2a956a9f7f434a015a42ba06835cfad5b83",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3cf11a2799e69f36ee6ee9015549b2a956a9f7f434a015a42ba06835cfad5b83",
        ],
    )

    rpm(
        name = "libvirt-libs-0__8.6.0-5.fc37.x86_64",
        sha256 = "035260aa3ad6e33ace0cfcf075d7f946bef0406484d5fd88a40209c160ca27e5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/035260aa3ad6e33ace0cfcf075d7f946bef0406484d5fd88a40209c160ca27e5",
        ],
    )

    rpm(
        name = "libvisual-1__0.4.0-36.fc37.x86_64",
        sha256 = "ab732e118bd7fc39c7f1ff467981bde50268d5d6206d81ce01037b78479d245a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ab732e118bd7fc39c7f1ff467981bde50268d5d6206d81ce01037b78479d245a",
        ],
    )

    rpm(
        name = "libvorbis-1__1.3.7-6.fc37.x86_64",
        sha256 = "ff3eec9dddb9293bb8c5d1267067ed5301ee417877125f5e5f2c5f45585a5fe1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ff3eec9dddb9293bb8c5d1267067ed5301ee417877125f5e5f2c5f45585a5fe1",
        ],
    )

    rpm(
        name = "libwayland-client-0__1.21.0-1.fc37.x86_64",
        sha256 = "f47e08a0ccff9ed6302d4a26770ff2a2b3ff6fd719cb0e66de4dff975e3bf8cb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f47e08a0ccff9ed6302d4a26770ff2a2b3ff6fd719cb0e66de4dff975e3bf8cb",
        ],
    )

    rpm(
        name = "libwayland-cursor-0__1.21.0-1.fc37.x86_64",
        sha256 = "c28d01fc2953f77c730444919e470597c2bfa3493a6c48ba8fe6c739152c675b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c28d01fc2953f77c730444919e470597c2bfa3493a6c48ba8fe6c739152c675b",
        ],
    )

    rpm(
        name = "libwayland-egl-0__1.21.0-1.fc37.x86_64",
        sha256 = "852e6c63dbc88917d966da31dcb011e224ec92b730f4ea34fc84add6ec7ef9ec",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/852e6c63dbc88917d966da31dcb011e224ec92b730f4ea34fc84add6ec7ef9ec",
        ],
    )

    rpm(
        name = "libwayland-server-0__1.21.0-1.fc37.x86_64",
        sha256 = "5da70288a025c06113b1768707604a1dc3a37d781a273da84752cfa569570047",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5da70288a025c06113b1768707604a1dc3a37d781a273da84752cfa569570047",
        ],
    )

    rpm(
        name = "libwebp-0__1.3.0-1.fc37.x86_64",
        sha256 = "be6b50cf9bda4246cf53eb628a672998bb37f8c28341d4ac905de346028f4bf2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/be6b50cf9bda4246cf53eb628a672998bb37f8c28341d4ac905de346028f4bf2",
        ],
    )

    rpm(
        name = "libwsman1-0__2.7.1-7.fc37.x86_64",
        sha256 = "41d9f13f5a7020a70f565e326fc5dd9167be30f3298ce5360081028cf2efbcab",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/41d9f13f5a7020a70f565e326fc5dd9167be30f3298ce5360081028cf2efbcab",
        ],
    )

    rpm(
        name = "libxcb-0__1.13.1-10.fc37.x86_64",
        sha256 = "e3226b05b41b3ba62d8910e76302acb3fc3686e834aa9bdabf26f7a28f65cd76",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3226b05b41b3ba62d8910e76302acb3fc3686e834aa9bdabf26f7a28f65cd76",
        ],
    )

    rpm(
        name = "libxcrypt-0__4.4.33-4.fc37.x86_64",
        sha256 = "547b9cffb0211abc4445d159e944f4fb59606b2eddfc14813b8c068859294ba6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/547b9cffb0211abc4445d159e944f4fb59606b2eddfc14813b8c068859294ba6",
        ],
    )

    rpm(
        name = "libxkbcommon-0__1.4.1-2.fc37.x86_64",
        sha256 = "ec9520e18b72651f309f78f4decc2b6571d04be898cf173bbfd6dc7b8d7611d4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ec9520e18b72651f309f78f4decc2b6571d04be898cf173bbfd6dc7b8d7611d4",
        ],
    )

    rpm(
        name = "libxml2-0__2.10.3-2.fc37.x86_64",
        sha256 = "105e8b221029cc4595682cd837dd80c1124685477efbec280fef2e2bb4974d2d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/105e8b221029cc4595682cd837dd80c1124685477efbec280fef2e2bb4974d2d",
        ],
    )

    rpm(
        name = "libxml__plus____plus__-0__2.42.2-1.fc37.x86_64",
        sha256 = "ef43127de1166abaca97466851baf41725e6dd2303d30b386c7177564774faaf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ef43127de1166abaca97466851baf41725e6dd2303d30b386c7177564774faaf",
        ],
    )

    rpm(
        name = "libxshmfence-0__1.3-11.fc37.x86_64",
        sha256 = "a1698c6585e2dd9df98b17e38863e91758c49f13cae76dcc1c8943cfd4faf292",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a1698c6585e2dd9df98b17e38863e91758c49f13cae76dcc1c8943cfd4faf292",
        ],
    )

    rpm(
        name = "libzstd-0__1.5.4-1.fc37.x86_64",
        sha256 = "d9c9de0b8805782ace29c7fbf5a922dc5d34c3e248f4a13b89a350584045d009",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d9c9de0b8805782ace29c7fbf5a922dc5d34c3e248f4a13b89a350584045d009",
        ],
    )

    rpm(
        name = "linux-atm-libs-0__2.5.1-33.fc37.x86_64",
        sha256 = "52a128f75f6cbe3555a44c6928390efe47d24a04ff0fc694033c74bda8a49542",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52a128f75f6cbe3555a44c6928390efe47d24a04ff0fc694033c74bda8a49542",
        ],
    )

    rpm(
        name = "llvm-libs-0__15.0.7-1.fc37.x86_64",
        sha256 = "bb01c1946fccde6933ebe937cef351a8cbbe49921a911f030787396c71d8d77d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bb01c1946fccde6933ebe937cef351a8cbbe49921a911f030787396c71d8d77d",
        ],
    )

    rpm(
        name = "lttng-ust-0__2.13.3-3.fc37.x86_64",
        sha256 = "765e366ec83ea61b782a4dc3b8afa209eb7c2166f53dfe1e619a0c6ba7e4a2f4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/765e366ec83ea61b782a4dc3b8afa209eb7c2166f53dfe1e619a0c6ba7e4a2f4",
        ],
    )

    rpm(
        name = "lua-libs-0__5.4.4-9.fc37.x86_64",
        sha256 = "561ebd5154e2d0d56f6a90283065b27304f81fc57fc881faf485e55d6414fad6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/561ebd5154e2d0d56f6a90283065b27304f81fc57fc881faf485e55d6414fad6",
        ],
    )

    rpm(
        name = "lvm2-0__2.03.11-9.fc37.x86_64",
        sha256 = "b8d2729bc77e301575d2afbe35bf93fac1ae7f95b741a0374bbe110ab880031f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8d2729bc77e301575d2afbe35bf93fac1ae7f95b741a0374bbe110ab880031f",
        ],
    )

    rpm(
        name = "lvm2-libs-0__2.03.11-9.fc37.x86_64",
        sha256 = "dcd2bd77993e452371337a25ca3eefae377a9b1edc105becdf45979b6bdfd7e4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dcd2bd77993e452371337a25ca3eefae377a9b1edc105becdf45979b6bdfd7e4",
        ],
    )

    rpm(
        name = "lz4-libs-0__1.9.4-1.fc37.x86_64",
        sha256 = "f39b8b018fcb2b55477cdbfa4af7c9db9b660c85000a4a42e880b1a951efbe5a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f39b8b018fcb2b55477cdbfa4af7c9db9b660c85000a4a42e880b1a951efbe5a",
        ],
    )

    rpm(
        name = "lzo-0__2.10-7.fc37.x86_64",
        sha256 = "fdde3f48dc7d4f5197d79b765c730aedc86632edc5fcbee3b50e876b2cf39e3e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fdde3f48dc7d4f5197d79b765c730aedc86632edc5fcbee3b50e876b2cf39e3e",
        ],
    )

    rpm(
        name = "lzop-0__1.04-9.fc37.x86_64",
        sha256 = "b93355cd20ca47f13fc56fba444347aa1189993a1b29eb7b3c1302079b6af44b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b93355cd20ca47f13fc56fba444347aa1189993a1b29eb7b3c1302079b6af44b",
        ],
    )

    rpm(
        name = "mdevctl-0__1.2.0-1.fc37.x86_64",
        sha256 = "aa87c9b0dafa770c5b53ecd9238f8e4663632ba41ccb6249f9dc2d105d479ddf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aa87c9b0dafa770c5b53ecd9238f8e4663632ba41ccb6249f9dc2d105d479ddf",
        ],
    )

    rpm(
        name = "mesa-dri-drivers-0__22.3.7-1.fc37.x86_64",
        sha256 = "95e61d0ff50aceda157f75ebed02e7465e86edcce23d86ec57700734f941a1f0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/95e61d0ff50aceda157f75ebed02e7465e86edcce23d86ec57700734f941a1f0",
        ],
    )

    rpm(
        name = "mesa-filesystem-0__22.3.7-1.fc37.x86_64",
        sha256 = "bf7a986940e4dcf0d96252e7e421cc244d1a8d5b3c50b6f86e2cff798bda5b90",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf7a986940e4dcf0d96252e7e421cc244d1a8d5b3c50b6f86e2cff798bda5b90",
        ],
    )

    rpm(
        name = "mesa-libEGL-0__22.3.7-1.fc37.x86_64",
        sha256 = "edf53f0b5568753f2db90f5b5aefc6b93ae8c05c9d23578d0c6575e2a9e24563",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/edf53f0b5568753f2db90f5b5aefc6b93ae8c05c9d23578d0c6575e2a9e24563",
        ],
    )

    rpm(
        name = "mesa-libGL-0__22.3.7-1.fc37.x86_64",
        sha256 = "72c6bd7aedb7b1b5aa745c4daffc323121cbaa700ce65786ec1e55adb62a5b20",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/72c6bd7aedb7b1b5aa745c4daffc323121cbaa700ce65786ec1e55adb62a5b20",
        ],
    )

    rpm(
        name = "mesa-libgbm-0__22.3.7-1.fc37.x86_64",
        sha256 = "b649a19c8b56539a346049c3f819d2e1a84ea3ef930698f3873812503762aafd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b649a19c8b56539a346049c3f819d2e1a84ea3ef930698f3873812503762aafd",
        ],
    )

    rpm(
        name = "mesa-libglapi-0__22.3.7-1.fc37.x86_64",
        sha256 = "7fa53487dc2d5a54c92c2dc4ace9038924a672c1be214fd6b8510e7a5980c7e2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7fa53487dc2d5a54c92c2dc4ace9038924a672c1be214fd6b8510e7a5980c7e2",
        ],
    )

    rpm(
        name = "mozjs102-0__102.9.0-1.fc37.x86_64",
        sha256 = "cc833951b3ff2d5527b618f17d670afe6b58d13155010fc74a9004b48f145e39",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cc833951b3ff2d5527b618f17d670afe6b58d13155010fc74a9004b48f145e39",
        ],
    )

    rpm(
        name = "mpdecimal-0__2.5.1-4.fc37.x86_64",
        sha256 = "45764a6773175638883e02215074f084de209d172d1d07be289e89aa5f4131d3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/45764a6773175638883e02215074f084de209d172d1d07be289e89aa5f4131d3",
        ],
    )

    rpm(
        name = "mpfr-0__4.1.0-10.fc37.x86_64",
        sha256 = "3be8cf104424fb5e148846a1df4a9c193527f55ee866bff0963e788450483566",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3be8cf104424fb5e148846a1df4a9c193527f55ee866bff0963e788450483566",
        ],
    )

    rpm(
        name = "mpg123-libs-0__1.31.3-1.fc37.x86_64",
        sha256 = "5a2688e9b31fcd5e3e31635d6bd254704c56ae2b432078a2b5596c4582bedfa8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5a2688e9b31fcd5e3e31635d6bd254704c56ae2b432078a2b5596c4582bedfa8",
        ],
    )

    rpm(
        name = "ncurses-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "7d90626c613d813fc63a1960985483aabf24ef2ab8b3b8f73cc9d8cac4fa6edd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7d90626c613d813fc63a1960985483aabf24ef2ab8b3b8f73cc9d8cac4fa6edd",
        ],
    )

    rpm(
        name = "ncurses-base-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "000164a9a82458fbb69b3433801dcc0d0e2437e21d7f7d4fd45f63a42a0bc26f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/000164a9a82458fbb69b3433801dcc0d0e2437e21d7f7d4fd45f63a42a0bc26f",
        ],
    )

    rpm(
        name = "ncurses-libs-0__6.3-4.20220501.fc37.x86_64",
        sha256 = "75e51eebcd3fe150b421ec5b1c9a6e918caa5b3c0f243f2b70d445fd434488bb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/75e51eebcd3fe150b421ec5b1c9a6e918caa5b3c0f243f2b70d445fd434488bb",
        ],
    )

    rpm(
        name = "ndctl-libs-0__76.1-1.fc37.x86_64",
        sha256 = "8c1c1d79191ba78c7db091a9dc79bc9db1f6bbad8191f8ddbcce015e7183b1e6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c1c1d79191ba78c7db091a9dc79bc9db1f6bbad8191f8ddbcce015e7183b1e6",
        ],
    )

    rpm(
        name = "nettle-0__3.8-2.fc37.x86_64",
        sha256 = "8fe2d98578b0c4454536faacbaafd66d1754b8439bb6332d7576a741f4c72208",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fe2d98578b0c4454536faacbaafd66d1754b8439bb6332d7576a741f4c72208",
        ],
    )

    rpm(
        name = "nfs-utils-1__2.6.2-2.rc6.fc37.x86_64",
        sha256 = "23f02af2336e2307823369f83fd6b63e7d15e5b66b0d9e43deab208393a34eb4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/23f02af2336e2307823369f83fd6b63e7d15e5b66b0d9e43deab208393a34eb4",
        ],
    )

    rpm(
        name = "nspr-0__4.35.0-5.fc37.x86_64",
        sha256 = "2b5ccbce76b445a5564ced2ae5bc395db5e49485fa6764663846f003542ec1ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2b5ccbce76b445a5564ced2ae5bc395db5e49485fa6764663846f003542ec1ac",
        ],
    )

    rpm(
        name = "nss-0__3.89.0-1.fc37.x86_64",
        sha256 = "eb31ce873dd92eef026f3d031d860b2e87a622f81648df620c09e9caa7661d97",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eb31ce873dd92eef026f3d031d860b2e87a622f81648df620c09e9caa7661d97",
        ],
    )

    rpm(
        name = "nss-softokn-0__3.89.0-1.fc37.x86_64",
        sha256 = "91d5cf0e568ffd5a2ec1de237ffac23bdcc54b42f12623c3b36a7a0b3cd0f789",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/91d5cf0e568ffd5a2ec1de237ffac23bdcc54b42f12623c3b36a7a0b3cd0f789",
        ],
    )

    rpm(
        name = "nss-softokn-freebl-0__3.89.0-1.fc37.x86_64",
        sha256 = "bf528aa71d1c018020f2467f35b1a579921bf10d8d09592ec1de90e1961c5533",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bf528aa71d1c018020f2467f35b1a579921bf10d8d09592ec1de90e1961c5533",
        ],
    )

    rpm(
        name = "nss-sysinit-0__3.89.0-1.fc37.x86_64",
        sha256 = "c9e20aa47b36c9352f0c79b4a6cfc9d49ed503ef4983fe3219dda6f937cfb5d8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c9e20aa47b36c9352f0c79b4a6cfc9d49ed503ef4983fe3219dda6f937cfb5d8",
        ],
    )

    rpm(
        name = "nss-util-0__3.89.0-1.fc37.x86_64",
        sha256 = "f267d65fee6a0198b9ee814e6200b8ddbe722c9d4bbc045181f7ee84831a194d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f267d65fee6a0198b9ee814e6200b8ddbe722c9d4bbc045181f7ee84831a194d",
        ],
    )

    rpm(
        name = "numactl-libs-0__2.0.14-6.fc37.x86_64",
        sha256 = "8f2e423d8f64f3abf33f8660df718d69f785a673a57eb188258a9f79af8f678f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8f2e423d8f64f3abf33f8660df718d69f785a673a57eb188258a9f79af8f678f",
        ],
    )

    rpm(
        name = "numad-0__0.5-37.20150602git.fc37.x86_64",
        sha256 = "de3ce09d93cebeb346d417dd656011e0323c6cfc983d585d093cbd199f2537a0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de3ce09d93cebeb346d417dd656011e0323c6cfc983d585d093cbd199f2537a0",
        ],
    )

    rpm(
        name = "openldap-0__2.6.4-1.fc37.x86_64",
        sha256 = "613788ec7bdccd9d14f3ffa97b06c32d43857a5ade51dc54d36d83a57007333c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/613788ec7bdccd9d14f3ffa97b06c32d43857a5ade51dc54d36d83a57007333c",
        ],
    )

    rpm(
        name = "openssl-devel-1__3.0.8-1.fc37.x86_64",
        sha256 = "28cbab4a2dadfdf33c1510d61f4ef48ef0f33165b22ff9d75233332d7a01df71",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/28cbab4a2dadfdf33c1510d61f4ef48ef0f33165b22ff9d75233332d7a01df71",
        ],
    )

    rpm(
        name = "openssl-libs-1__3.0.8-1.fc37.x86_64",
        sha256 = "f250396bc408a880a50a53535e8038d593107594af1d9d348c01aa27a6348dae",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f250396bc408a880a50a53535e8038d593107594af1d9d348c01aa27a6348dae",
        ],
    )

    rpm(
        name = "opus-0__1.3.1-11.fc37.x86_64",
        sha256 = "79f646709769db7e9f7246ae8e63088d88b1888bc43b12fc6a495af6161c7300",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79f646709769db7e9f7246ae8e63088d88b1888bc43b12fc6a495af6161c7300",
        ],
    )

    rpm(
        name = "orc-0__0.4.31-8.fc37.x86_64",
        sha256 = "2016e01f768a65aa7beac8f8e0017df968f0549797bcf5efaed0713cbaea2d0a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2016e01f768a65aa7beac8f8e0017df968f0549797bcf5efaed0713cbaea2d0a",
        ],
    )

    rpm(
        name = "p11-kit-0__0.24.1-3.fc37.x86_64",
        sha256 = "4dad6ac54eb7708cbfc8522d372f2a196cf711e97e279cbddba8cc8b92970dd7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4dad6ac54eb7708cbfc8522d372f2a196cf711e97e279cbddba8cc8b92970dd7",
        ],
    )

    rpm(
        name = "p11-kit-trust-0__0.24.1-3.fc37.x86_64",
        sha256 = "0fd85eb1ce27615fea745721b18648b4a4585ad4b11a482c1b77fc1785cd5194",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0fd85eb1ce27615fea745721b18648b4a4585ad4b11a482c1b77fc1785cd5194",
        ],
    )

    rpm(
        name = "pam-0__1.5.2-14.fc37.x86_64",
        sha256 = "a66ee1c9f9155c97e77cbd18658ce5129638f7d6e208c01c172c4dd1dfdbbe6d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a66ee1c9f9155c97e77cbd18658ce5129638f7d6e208c01c172c4dd1dfdbbe6d",
        ],
    )

    rpm(
        name = "pam-libs-0__1.5.2-14.fc37.x86_64",
        sha256 = "ee34422adc6451da744bd16a8cd66c9912a822c4e55227c23ff56960c32980f5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee34422adc6451da744bd16a8cd66c9912a822c4e55227c23ff56960c32980f5",
        ],
    )

    rpm(
        name = "pango-0__1.50.14-1.fc37.x86_64",
        sha256 = "4c9168ae1c9e229b4b0cb7d90f7d5a11d5ad8590838811659c1f02547100b660",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4c9168ae1c9e229b4b0cb7d90f7d5a11d5ad8590838811659c1f02547100b660",
        ],
    )

    rpm(
        name = "parted-0__3.5-6.fc37.x86_64",
        sha256 = "cba7fbe4f72489a2fb27429ecfdf8b0177ac8cc4fa30aa7baaa8ce1a0dc4f5ea",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cba7fbe4f72489a2fb27429ecfdf8b0177ac8cc4fa30aa7baaa8ce1a0dc4f5ea",
        ],
    )

    rpm(
        name = "pcre-0__8.45-1.fc37.2.x86_64",
        sha256 = "86a648e3b88f581b15ca2eda6b441be7c5c3810a9eae25ca940c767029e4e923",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86a648e3b88f581b15ca2eda6b441be7c5c3810a9eae25ca940c767029e4e923",
        ],
    )

    rpm(
        name = "pcre2-0__10.40-1.fc37.1.x86_64",
        sha256 = "422de947ec1a7aafcd212a51e64257b64d5b0a02808104a33e7c3cd9ef629148",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/422de947ec1a7aafcd212a51e64257b64d5b0a02808104a33e7c3cd9ef629148",
        ],
    )

    rpm(
        name = "pcre2-devel-0__10.40-1.fc37.1.x86_64",
        sha256 = "a0bf3bcf08ea68f1d65c0e68e574083e40f8e7e60749ed9c4f9f0494ad773d3c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a0bf3bcf08ea68f1d65c0e68e574083e40f8e7e60749ed9c4f9f0494ad773d3c",
        ],
    )

    rpm(
        name = "pcre2-syntax-0__10.40-1.fc37.1.x86_64",
        sha256 = "585f339942a0bf4b0eab638ddf825544793485cbcb9f1eaee079b9956d90aafa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/585f339942a0bf4b0eab638ddf825544793485cbcb9f1eaee079b9956d90aafa",
        ],
    )

    rpm(
        name = "pcre2-utf16-0__10.40-1.fc37.1.x86_64",
        sha256 = "c3c743090c216ff44c0ab41f3632e817fd738b379df098a048f6df5160e71b51",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c3c743090c216ff44c0ab41f3632e817fd738b379df098a048f6df5160e71b51",
        ],
    )

    rpm(
        name = "pcre2-utf32-0__10.40-1.fc37.1.x86_64",
        sha256 = "ed2cc3196b6fc344bd70c033fdcb1a766dafbbf0a5a826059c72640d302ae4ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ed2cc3196b6fc344bd70c033fdcb1a766dafbbf0a5a826059c72640d302ae4ac",
        ],
    )

    rpm(
        name = "pcsc-lite-libs-0__1.9.9-1.fc37.x86_64",
        sha256 = "de4dcfa7271758f759bfa291ac552eccb586047ee0b91471991bb46b2161b2a3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de4dcfa7271758f759bfa291ac552eccb586047ee0b91471991bb46b2161b2a3",
        ],
    )

    rpm(
        name = "perl-Carp-0__1.52-489.fc37.x86_64",
        sha256 = "c5df34198e7dd39f4f09032beacb9db641c8752d045b8e1f8cacd2637559dd1d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c5df34198e7dd39f4f09032beacb9db641c8752d045b8e1f8cacd2637559dd1d",
        ],
    )

    rpm(
        name = "perl-Class-Struct-0__0.66-492.fc37.x86_64",
        sha256 = "625c6cc3d5238fd26369e8190ab57e81b14c5c702e45cf191203b042b3e34807",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/625c6cc3d5238fd26369e8190ab57e81b14c5c702e45cf191203b042b3e34807",
        ],
    )

    rpm(
        name = "perl-DynaLoader-0__1.52-492.fc37.x86_64",
        sha256 = "e7373c9bd3e688edc3a68cd34c92999b3d70d8c26ead19bdbd973971d15308a3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7373c9bd3e688edc3a68cd34c92999b3d70d8c26ead19bdbd973971d15308a3",
        ],
    )

    rpm(
        name = "perl-Encode-4__3.19-492.fc37.x86_64",
        sha256 = "395705071c61bc6faad7cde8eb251d3bff87f543c5096c00139331e9ee2ba856",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/395705071c61bc6faad7cde8eb251d3bff87f543c5096c00139331e9ee2ba856",
        ],
    )

    rpm(
        name = "perl-Errno-0__1.36-492.fc37.x86_64",
        sha256 = "aaaca92e6353f4cc2d4b81e73efe7b01c3712ed285d6faf940e029a07d853eb4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aaaca92e6353f4cc2d4b81e73efe7b01c3712ed285d6faf940e029a07d853eb4",
        ],
    )

    rpm(
        name = "perl-Exporter-0__5.77-489.fc37.x86_64",
        sha256 = "95b26bb47a5b0f52f091cf6a5e6a493b203ed6d1bf8de714ed182c4f78f8b351",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/95b26bb47a5b0f52f091cf6a5e6a493b203ed6d1bf8de714ed182c4f78f8b351",
        ],
    )

    rpm(
        name = "perl-Fcntl-0__1.15-492.fc37.x86_64",
        sha256 = "3e0b22859a921f40ad384d32b212f15ed4dbaecbf41d1077a6a6d1bfc46dc50f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3e0b22859a921f40ad384d32b212f15ed4dbaecbf41d1077a6a6d1bfc46dc50f",
        ],
    )

    rpm(
        name = "perl-File-Basename-0__2.85-492.fc37.x86_64",
        sha256 = "aefb2c6be89b24319aeba4ef9e21623f7dbcc478753219ec32f6f07748f6be5c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aefb2c6be89b24319aeba4ef9e21623f7dbcc478753219ec32f6f07748f6be5c",
        ],
    )

    rpm(
        name = "perl-File-Path-0__2.18-489.fc37.x86_64",
        sha256 = "d73acee5758b6ac85b46416094fcaf9a3e9f54193996f167d353777e5708745a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d73acee5758b6ac85b46416094fcaf9a3e9f54193996f167d353777e5708745a",
        ],
    )

    rpm(
        name = "perl-File-Temp-1__0.231.100-489.fc37.x86_64",
        sha256 = "5724aaa686d2cd278da72939aa01678cb6d9ba2b43237425ce2378197fbcc0d0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5724aaa686d2cd278da72939aa01678cb6d9ba2b43237425ce2378197fbcc0d0",
        ],
    )

    rpm(
        name = "perl-File-stat-0__1.12-492.fc37.x86_64",
        sha256 = "dcadef87fb0da0f5dab7f3e051c8423b7c951dfdedd7b18af3be461e05441945",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dcadef87fb0da0f5dab7f3e051c8423b7c951dfdedd7b18af3be461e05441945",
        ],
    )

    rpm(
        name = "perl-Getopt-Long-1__2.54-1.fc37.x86_64",
        sha256 = "49e397448b20d87418f8a7b7b3ae474928b6e33157dbaf300dc8d8ed842804c8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49e397448b20d87418f8a7b7b3ae474928b6e33157dbaf300dc8d8ed842804c8",
        ],
    )

    rpm(
        name = "perl-Getopt-Std-0__1.13-492.fc37.x86_64",
        sha256 = "2d8846f2950c6b09ae72e5584291ac7d775faabba897b8db760444ebc402146e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2d8846f2950c6b09ae72e5584291ac7d775faabba897b8db760444ebc402146e",
        ],
    )

    rpm(
        name = "perl-HTTP-Tiny-0__0.082-1.fc37.x86_64",
        sha256 = "9e27a04da9f65e27e0eb8bbcad9ec5dfaa986e11731f009700c453db32676455",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9e27a04da9f65e27e0eb8bbcad9ec5dfaa986e11731f009700c453db32676455",
        ],
    )

    rpm(
        name = "perl-IO-0__1.50-492.fc37.x86_64",
        sha256 = "86beaf16f888309ca2c55c240a374f1bd3350770973732d90467261fe7d5620e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86beaf16f888309ca2c55c240a374f1bd3350770973732d90467261fe7d5620e",
        ],
    )

    rpm(
        name = "perl-IPC-Open3-0__1.22-492.fc37.x86_64",
        sha256 = "7c53e7f6080e0aa53236ee673e2b4b20df87fdf3503620e9eef871c217dcd937",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c53e7f6080e0aa53236ee673e2b4b20df87fdf3503620e9eef871c217dcd937",
        ],
    )

    rpm(
        name = "perl-MIME-Base64-0__3.16-489.fc37.x86_64",
        sha256 = "0823ebce69b5d9df94c9d508669ac612bd1df48e8857cf02103dcc2094df246f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0823ebce69b5d9df94c9d508669ac612bd1df48e8857cf02103dcc2094df246f",
        ],
    )

    rpm(
        name = "perl-POSIX-0__2.03-492.fc37.x86_64",
        sha256 = "b4b9f1b5e5fd0b533d88d34c28d98115a6a43759ea7be61d6b6d948c47d4e298",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b4b9f1b5e5fd0b533d88d34c28d98115a6a43759ea7be61d6b6d948c47d4e298",
        ],
    )

    rpm(
        name = "perl-PathTools-0__3.84-489.fc37.x86_64",
        sha256 = "e0a37fde6728e9b662d46a28dcd34399eccda9e5b696858d1447a47a82877de1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0a37fde6728e9b662d46a28dcd34399eccda9e5b696858d1447a47a82877de1",
        ],
    )

    rpm(
        name = "perl-Pod-Escapes-1__1.07-489.fc37.x86_64",
        sha256 = "08bf74ca4b30a562d26026d0a70b508b7b926a3697d700127accf9d052c30da1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/08bf74ca4b30a562d26026d0a70b508b7b926a3697d700127accf9d052c30da1",
        ],
    )

    rpm(
        name = "perl-Pod-Perldoc-0__3.28.01-490.fc37.x86_64",
        sha256 = "0047003615e8d018a4f4a887b414cba325ef294084faab480533f5de78ef58a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0047003615e8d018a4f4a887b414cba325ef294084faab480533f5de78ef58a4",
        ],
    )

    rpm(
        name = "perl-Pod-Simple-1__3.43-490.fc37.x86_64",
        sha256 = "298ba92d8130493373f70271685ba59298377da497a39614d14a6345a44cac8a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/298ba92d8130493373f70271685ba59298377da497a39614d14a6345a44cac8a",
        ],
    )

    rpm(
        name = "perl-Pod-Usage-4__2.03-3.fc37.x86_64",
        sha256 = "c216ffe76ed543dcb546178dabcceeca4a1054c81beee8de67a741ad39af314e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c216ffe76ed543dcb546178dabcceeca4a1054c81beee8de67a741ad39af314e",
        ],
    )

    rpm(
        name = "perl-Scalar-List-Utils-5__1.63-489.fc37.x86_64",
        sha256 = "eed91d8529a1eee7269fd28eccdb636a6fc6dcf30d607bdb7198329f3017a74a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eed91d8529a1eee7269fd28eccdb636a6fc6dcf30d607bdb7198329f3017a74a",
        ],
    )

    rpm(
        name = "perl-SelectSaver-0__1.02-492.fc37.x86_64",
        sha256 = "3b54fe377ca9a901ac80d3bcb70518537c383135e10232a6333135f73325a6e2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b54fe377ca9a901ac80d3bcb70518537c383135e10232a6333135f73325a6e2",
        ],
    )

    rpm(
        name = "perl-Socket-4__2.036-1.fc37.x86_64",
        sha256 = "5cda2738aaff2c73a850f4359b299028560c2106014c3bb02b340f036d81b564",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5cda2738aaff2c73a850f4359b299028560c2106014c3bb02b340f036d81b564",
        ],
    )

    rpm(
        name = "perl-Storable-1__3.26-489.fc37.x86_64",
        sha256 = "22e7e312778e31de59a26759b523e4d0916429c17bd23f51a5ddc550d3b33910",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/22e7e312778e31de59a26759b523e4d0916429c17bd23f51a5ddc550d3b33910",
        ],
    )

    rpm(
        name = "perl-Symbol-0__1.09-492.fc37.x86_64",
        sha256 = "1a52209c5a9bad4c93923470acbeaa8dc4e87a5e03513d01f7bec0147ec9ff7e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1a52209c5a9bad4c93923470acbeaa8dc4e87a5e03513d01f7bec0147ec9ff7e",
        ],
    )

    rpm(
        name = "perl-Term-ANSIColor-0__5.01-490.fc37.x86_64",
        sha256 = "4810f5377abb3ee2cbd1b2572b1c67421145304f54c491afc770101f90d27138",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4810f5377abb3ee2cbd1b2572b1c67421145304f54c491afc770101f90d27138",
        ],
    )

    rpm(
        name = "perl-Term-Cap-0__1.17-489.fc37.x86_64",
        sha256 = "65d62ee56cba18a4fcbba844375b6ed49c024863fa6ea948e2ce699e7ab66298",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/65d62ee56cba18a4fcbba844375b6ed49c024863fa6ea948e2ce699e7ab66298",
        ],
    )

    rpm(
        name = "perl-Text-ParseWords-0__3.31-489.fc37.x86_64",
        sha256 = "c66cc3f03aed1781132e402db92bc1bbcf27ab08283766428b55c368127fee01",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c66cc3f03aed1781132e402db92bc1bbcf27ab08283766428b55c368127fee01",
        ],
    )

    rpm(
        name = "perl-Text-Tabs__plus__Wrap-0__2021.0814-489.fc37.x86_64",
        sha256 = "2bd5a61328042e939c3360586232a8c5579d4284bfa83d92ec89a578c733c253",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2bd5a61328042e939c3360586232a8c5579d4284bfa83d92ec89a578c733c253",
        ],
    )

    rpm(
        name = "perl-Time-Local-2__1.300-489.fc37.x86_64",
        sha256 = "db087f574a6fd314f45daca13a1f4404c76144773fcd6c41df39da7fda862d8d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/db087f574a6fd314f45daca13a1f4404c76144773fcd6c41df39da7fda862d8d",
        ],
    )

    rpm(
        name = "perl-constant-0__1.33-490.fc37.x86_64",
        sha256 = "7898312e5b93625a6fa5f4b60300b06260668dfb12af04602a3354266e8f7850",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7898312e5b93625a6fa5f4b60300b06260668dfb12af04602a3354266e8f7850",
        ],
    )

    rpm(
        name = "perl-if-0__0.61.000-492.fc37.x86_64",
        sha256 = "7d2a698fdc1923e3091359978026b3cbf98f7774c302c8930bf68488ccb83dd2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7d2a698fdc1923e3091359978026b3cbf98f7774c302c8930bf68488ccb83dd2",
        ],
    )

    rpm(
        name = "perl-interpreter-4__5.36.0-492.fc37.x86_64",
        sha256 = "b6c1c4885bc6ca7b7bfbc95e0ebe7bca41042105fd97a264181534907b6be73e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b6c1c4885bc6ca7b7bfbc95e0ebe7bca41042105fd97a264181534907b6be73e",
        ],
    )

    rpm(
        name = "perl-libs-4__5.36.0-492.fc37.x86_64",
        sha256 = "5d41d4194ffc1e78f3816383b83e424984cf81b9c481e76d1453ab37f4298b2a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5d41d4194ffc1e78f3816383b83e424984cf81b9c481e76d1453ab37f4298b2a",
        ],
    )

    rpm(
        name = "perl-mro-0__1.26-492.fc37.x86_64",
        sha256 = "915be0c666707356a59135f0936925aaea35a31df143107188d929163c83051f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/915be0c666707356a59135f0936925aaea35a31df143107188d929163c83051f",
        ],
    )

    rpm(
        name = "perl-overload-0__1.35-492.fc37.x86_64",
        sha256 = "1939f3e239071a62d376b00c0a848ddbd500c2c7c1f5321d794d4d7bf52f94fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1939f3e239071a62d376b00c0a848ddbd500c2c7c1f5321d794d4d7bf52f94fd",
        ],
    )

    rpm(
        name = "perl-overloading-0__0.02-492.fc37.x86_64",
        sha256 = "089efb5cbe1bc3281e0ae6db4bf5a25e64bfe60952f7598231d567a16092db7d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/089efb5cbe1bc3281e0ae6db4bf5a25e64bfe60952f7598231d567a16092db7d",
        ],
    )

    rpm(
        name = "perl-parent-1__0.238-489.fc37.x86_64",
        sha256 = "202ebb9e6bf82022838cbe5814a9a38ec60e3d74864264bd48963a48032177df",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/202ebb9e6bf82022838cbe5814a9a38ec60e3d74864264bd48963a48032177df",
        ],
    )

    rpm(
        name = "perl-podlators-1__4.14-489.fc37.x86_64",
        sha256 = "1dbfed6c41e81aa507a5a723fbed9407ab7a6a8fef7b054c742c1feadf388cba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1dbfed6c41e81aa507a5a723fbed9407ab7a6a8fef7b054c742c1feadf388cba",
        ],
    )

    rpm(
        name = "perl-subs-0__1.04-492.fc37.x86_64",
        sha256 = "c1fc6b645ca97462145c0161e187206c7e216ba3671b141f1b56b4039cf20d20",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c1fc6b645ca97462145c0161e187206c7e216ba3671b141f1b56b4039cf20d20",
        ],
    )

    rpm(
        name = "perl-vars-0__1.05-492.fc37.x86_64",
        sha256 = "b32516b678cd965c3e6cf6c6fae8790384ce89bd491f05cb2d5fc02fc225da7f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b32516b678cd965c3e6cf6c6fae8790384ce89bd491f05cb2d5fc02fc225da7f",
        ],
    )

    rpm(
        name = "pixman-0__0.40.0-6.fc37.x86_64",
        sha256 = "131619876f2f68070ef4e178b5758474f54c577e5a6bf7a88746db54f0d0231f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/131619876f2f68070ef4e178b5758474f54c577e5a6bf7a88746db54f0d0231f",
        ],
    )

    rpm(
        name = "pkgconf-0__1.8.0-3.fc37.x86_64",
        sha256 = "778018594ab5bddc4432e53985b80e6c5a1a1ec1700d38b438848d485f5b357c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/778018594ab5bddc4432e53985b80e6c5a1a1ec1700d38b438848d485f5b357c",
        ],
    )

    rpm(
        name = "pkgconf-m4-0__1.8.0-3.fc37.x86_64",
        sha256 = "dd0356475d0b9106b5a2d577db359aa0290fe6dd9eacea1b6e0cab816ff33566",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dd0356475d0b9106b5a2d577db359aa0290fe6dd9eacea1b6e0cab816ff33566",
        ],
    )

    rpm(
        name = "pkgconf-pkg-config-0__1.8.0-3.fc37.x86_64",
        sha256 = "d238b12c750b58ceebc80e25c2074bd929d3f232c1390677f33a94fdadb68f6a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d238b12c750b58ceebc80e25c2074bd929d3f232c1390677f33a94fdadb68f6a",
        ],
    )

    rpm(
        name = "policycoreutils-0__3.5-1.fc37.x86_64",
        sha256 = "199f28e3ecd24e0650ed7b8bc596f14e7a9cdef79c36989b710da1f557b347d9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/199f28e3ecd24e0650ed7b8bc596f14e7a9cdef79c36989b710da1f557b347d9",
        ],
    )

    rpm(
        name = "policycoreutils-python-utils-0__3.5-1.fc37.x86_64",
        sha256 = "6a9db2d2e789177cad01f4cf52bd9ac47562127869c26eaf091134770a7e20a3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6a9db2d2e789177cad01f4cf52bd9ac47562127869c26eaf091134770a7e20a3",
        ],
    )

    rpm(
        name = "polkit-0__121-4.fc37.x86_64",
        sha256 = "3df7fd45694c994a41077966f80f48d28a5466c21e161c92d05629cc64efe0e6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3df7fd45694c994a41077966f80f48d28a5466c21e161c92d05629cc64efe0e6",
        ],
    )

    rpm(
        name = "polkit-libs-0__121-4.fc37.x86_64",
        sha256 = "e16b0c15a0820ca0795ac8e5cd196c4c8bef3dfebd70eef28a5152a2c69a9386",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e16b0c15a0820ca0795ac8e5cd196c4c8bef3dfebd70eef28a5152a2c69a9386",
        ],
    )

    rpm(
        name = "polkit-pkla-compat-0__0.1-22.fc37.x86_64",
        sha256 = "fd67360c39e9d4bd6eb9c6179ec2fa68382715129a8a43972e9fe63ee870f420",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fd67360c39e9d4bd6eb9c6179ec2fa68382715129a8a43972e9fe63ee870f420",
        ],
    )

    rpm(
        name = "popt-0__1.19-1.fc37.x86_64",
        sha256 = "e3c9a6a1611d967fbff4321b5b1ae54377fed22454298859108138c1f64b0c63",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3c9a6a1611d967fbff4321b5b1ae54377fed22454298859108138c1f64b0c63",
        ],
    )

    rpm(
        name = "protobuf-c-0__1.4.1-2.fc37.x86_64",
        sha256 = "46a9be44b3444815a0197dd85953bf87710d3ea3d8f9fbfff23068ca85885070",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/46a9be44b3444815a0197dd85953bf87710d3ea3d8f9fbfff23068ca85885070",
        ],
    )

    rpm(
        name = "psmisc-0__23.4-4.fc37.x86_64",
        sha256 = "4a50276d2a8bd7f2b16cbdf6592ea365c19fa9f9bd33b5cfa1ab4bbd5ca89c69",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a50276d2a8bd7f2b16cbdf6592ea365c19fa9f9bd33b5cfa1ab4bbd5ca89c69",
        ],
    )

    rpm(
        name = "publicsuffix-list-dafsa-0__20210518-5.fc37.x86_64",
        sha256 = "66f0cb20aae801f5810d2bdd27f0d6b9a70935a231e04269611113f96989132e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66f0cb20aae801f5810d2bdd27f0d6b9a70935a231e04269611113f96989132e",
        ],
    )

    rpm(
        name = "pulseaudio-libs-0__16.1-4.fc37.x86_64",
        sha256 = "678817de986951a9c377ff74b6ec5912fa8e3dc1ad46495de57a5fb7aa9ff905",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/678817de986951a9c377ff74b6ec5912fa8e3dc1ad46495de57a5fb7aa9ff905",
        ],
    )

    rpm(
        name = "python-pip-wheel-0__22.2.2-3.fc37.x86_64",
        sha256 = "f7800b3f5acca7863bf47981258582728b74861c19b2bef38ae47efe3b042eb4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f7800b3f5acca7863bf47981258582728b74861c19b2bef38ae47efe3b042eb4",
        ],
    )

    rpm(
        name = "python-setuptools-wheel-0__62.6.0-2.fc37.x86_64",
        sha256 = "5a9c2a69949d1bd9293d3fd34719e4d01c8e65d80957d8534ebc23b1deb756c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5a9c2a69949d1bd9293d3fd34719e4d01c8e65d80957d8534ebc23b1deb756c2",
        ],
    )

    rpm(
        name = "python3-0__3.11.2-1.fc37.x86_64",
        sha256 = "9eebbf2abbc9791597032fb6136b92e99cadba5fc0b1e673c923b77ff9a5af8c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9eebbf2abbc9791597032fb6136b92e99cadba5fc0b1e673c923b77ff9a5af8c",
        ],
    )

    rpm(
        name = "python3-audit-0__3.1-2.fc37.x86_64",
        sha256 = "53e518304fe3f9d6b3952b29fe7878961af69178d2e0293dff3e2aacf1abe644",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/53e518304fe3f9d6b3952b29fe7878961af69178d2e0293dff3e2aacf1abe644",
        ],
    )

    rpm(
        name = "python3-distro-0__1.7.0-3.fc37.x86_64",
        sha256 = "6427a8b877a51be140e06665b3680a911816c9eb581aa54ecf29c77f51b6e898",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6427a8b877a51be140e06665b3680a911816c9eb581aa54ecf29c77f51b6e898",
        ],
    )

    rpm(
        name = "python3-libs-0__3.11.2-1.fc37.x86_64",
        sha256 = "1681d31085e638e38b2836d0c71bcbf6ddc8b79386824ea403e0886c9fc9f98d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1681d31085e638e38b2836d0c71bcbf6ddc8b79386824ea403e0886c9fc9f98d",
        ],
    )

    rpm(
        name = "python3-libselinux-0__3.5-1.fc37.x86_64",
        sha256 = "d6a8fff22472e9629ae8d61c678448215a4f3fb3314b26d0fde0dee70996b3a0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d6a8fff22472e9629ae8d61c678448215a4f3fb3314b26d0fde0dee70996b3a0",
        ],
    )

    rpm(
        name = "python3-libsemanage-0__3.5-1.fc37.x86_64",
        sha256 = "e7710d9e96771a3becdb771b28d9a2de6bbcb87a3d1a7319e1f4eb9581f242b8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7710d9e96771a3becdb771b28d9a2de6bbcb87a3d1a7319e1f4eb9581f242b8",
        ],
    )

    rpm(
        name = "python3-policycoreutils-0__3.5-1.fc37.x86_64",
        sha256 = "e746e0a43f9a9d57b33964332f595490f73df5df751cc785c66b4b72d4633410",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e746e0a43f9a9d57b33964332f595490f73df5df751cc785c66b4b72d4633410",
        ],
    )

    rpm(
        name = "python3-setools-0__4.4.0-9.fc37.x86_64",
        sha256 = "b872895c9e0ddbebdf970a42f8c7a25d956fb7bbcdbb18aa4cf5515cff962c62",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b872895c9e0ddbebdf970a42f8c7a25d956fb7bbcdbb18aa4cf5515cff962c62",
        ],
    )

    rpm(
        name = "python3-setuptools-0__62.6.0-2.fc37.x86_64",
        sha256 = "f54b9672f6cdc282263610a619ebef76a69d245bf1588fe767ea136c31c3c93b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f54b9672f6cdc282263610a619ebef76a69d245bf1588fe767ea136c31c3c93b",
        ],
    )

    rpm(
        name = "qemu-audio-alsa-2__7.0.0-14.fc37.x86_64",
        sha256 = "b3d0feb0f3f8883526affeafc869caa8d7035feb14dedd5649df6ece4b73f22a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b3d0feb0f3f8883526affeafc869caa8d7035feb14dedd5649df6ece4b73f22a",
        ],
    )

    rpm(
        name = "qemu-audio-dbus-2__7.0.0-14.fc37.x86_64",
        sha256 = "04bfeaeacac6778941a5a047356da9b151a3454633fe6fe755ade1131b225717",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/04bfeaeacac6778941a5a047356da9b151a3454633fe6fe755ade1131b225717",
        ],
    )

    rpm(
        name = "qemu-audio-jack-2__7.0.0-14.fc37.x86_64",
        sha256 = "5e680b791b61a05c75685753876798d52d963aa1ab379088a38cc519362c5168",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5e680b791b61a05c75685753876798d52d963aa1ab379088a38cc519362c5168",
        ],
    )

    rpm(
        name = "qemu-audio-oss-2__7.0.0-14.fc37.x86_64",
        sha256 = "c0f05b0febc62e11efee11df0c83994002e5de27b01526d37f97f45564567efd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c0f05b0febc62e11efee11df0c83994002e5de27b01526d37f97f45564567efd",
        ],
    )

    rpm(
        name = "qemu-audio-pa-2__7.0.0-14.fc37.x86_64",
        sha256 = "42b3a2144cde3683d370ccbaf0a9cdc9ad201eefdc003c6615586bb7091f8188",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42b3a2144cde3683d370ccbaf0a9cdc9ad201eefdc003c6615586bb7091f8188",
        ],
    )

    rpm(
        name = "qemu-audio-sdl-2__7.0.0-14.fc37.x86_64",
        sha256 = "f3cd799e9a0d2f0e7b64af3cbf54d70218453930ef8fcc37651958d689e19e38",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f3cd799e9a0d2f0e7b64af3cbf54d70218453930ef8fcc37651958d689e19e38",
        ],
    )

    rpm(
        name = "qemu-audio-spice-2__7.0.0-14.fc37.x86_64",
        sha256 = "a174c722eca68c4ee4d700f51643ca219cfa5d922de96ff396d90a306b4f8267",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a174c722eca68c4ee4d700f51643ca219cfa5d922de96ff396d90a306b4f8267",
        ],
    )

    rpm(
        name = "qemu-block-curl-2__7.0.0-14.fc37.x86_64",
        sha256 = "e83878568c2cc55d35df1d04b61b13e182c0c75b9a2085a4c953f3920d19d6a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e83878568c2cc55d35df1d04b61b13e182c0c75b9a2085a4c953f3920d19d6a4",
        ],
    )

    rpm(
        name = "qemu-block-dmg-2__7.0.0-14.fc37.x86_64",
        sha256 = "a4da9e237ae3eeb6cb3b00d08db86301b1b6e7100cd16db154a4e9df0cd17eaa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a4da9e237ae3eeb6cb3b00d08db86301b1b6e7100cd16db154a4e9df0cd17eaa",
        ],
    )

    rpm(
        name = "qemu-block-gluster-2__7.0.0-14.fc37.x86_64",
        sha256 = "b30547de393c0b3e20eed6fc5c01b08ef4222c6db9ff37cbc5f918f81366b316",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b30547de393c0b3e20eed6fc5c01b08ef4222c6db9ff37cbc5f918f81366b316",
        ],
    )

    rpm(
        name = "qemu-block-iscsi-2__7.0.0-14.fc37.x86_64",
        sha256 = "8ec06f366acd58ccce9ef00ba7a43ff6ea6649d5223489100a2cf3f2da0aefad",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8ec06f366acd58ccce9ef00ba7a43ff6ea6649d5223489100a2cf3f2da0aefad",
        ],
    )

    rpm(
        name = "qemu-block-nfs-2__7.0.0-14.fc37.x86_64",
        sha256 = "ff8f7ea41fca48a4f111db7134decbe9c0b4bb7b67ffe704718847a5b830e281",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ff8f7ea41fca48a4f111db7134decbe9c0b4bb7b67ffe704718847a5b830e281",
        ],
    )

    rpm(
        name = "qemu-block-rbd-2__7.0.0-14.fc37.x86_64",
        sha256 = "0ca7eba769a95489d7f5b31639564a8fe8bbef12112be19eef822c319d4d2f2d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0ca7eba769a95489d7f5b31639564a8fe8bbef12112be19eef822c319d4d2f2d",
        ],
    )

    rpm(
        name = "qemu-block-ssh-2__7.0.0-14.fc37.x86_64",
        sha256 = "9789e2c7c0fa220b26723c13b6278a5392967271cc12d09ba7086e4096439173",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9789e2c7c0fa220b26723c13b6278a5392967271cc12d09ba7086e4096439173",
        ],
    )

    rpm(
        name = "qemu-char-baum-2__7.0.0-14.fc37.x86_64",
        sha256 = "34e182c1561e602bc8676d5650af0544cef0cb32b48ae5b5d777c73288a9c461",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/34e182c1561e602bc8676d5650af0544cef0cb32b48ae5b5d777c73288a9c461",
        ],
    )

    rpm(
        name = "qemu-char-spice-2__7.0.0-14.fc37.x86_64",
        sha256 = "2479ad005bd8065821e8303193c50e482841408bbc61a94c3ce6ff3f1673234d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2479ad005bd8065821e8303193c50e482841408bbc61a94c3ce6ff3f1673234d",
        ],
    )

    rpm(
        name = "qemu-common-2__7.0.0-14.fc37.x86_64",
        sha256 = "7dce214b472568df84d12f55cc838b0df8db8afbc00b54f145989c2d5c71a63c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7dce214b472568df84d12f55cc838b0df8db8afbc00b54f145989c2d5c71a63c",
        ],
    )

    rpm(
        name = "qemu-device-display-qxl-2__7.0.0-14.fc37.x86_64",
        sha256 = "d224418106c8a0d86c511161e5169f1278ee6c3bc44ea90668039698f76a4ac5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d224418106c8a0d86c511161e5169f1278ee6c3bc44ea90668039698f76a4ac5",
        ],
    )

    rpm(
        name = "qemu-device-display-vhost-user-gpu-2__7.0.0-14.fc37.x86_64",
        sha256 = "7c2bbc6ae6f08bc08069a65d0e99e246ac4bc4c0f4064ed0b8c9a5c3cf454dd2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7c2bbc6ae6f08bc08069a65d0e99e246ac4bc4c0f4064ed0b8c9a5c3cf454dd2",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-gpu-2__7.0.0-14.fc37.x86_64",
        sha256 = "51a0538164f52ab8611d1cf724457e2a8a9879eac421cea84a909cff7e1b9119",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/51a0538164f52ab8611d1cf724457e2a8a9879eac421cea84a909cff7e1b9119",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-gpu-ccw-2__7.0.0-14.fc37.x86_64",
        sha256 = "3eb525b103eb5070708a34d876f4bac3114f1a5d83d11abce96f111536ea6103",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3eb525b103eb5070708a34d876f4bac3114f1a5d83d11abce96f111536ea6103",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-gpu-gl-2__7.0.0-14.fc37.x86_64",
        sha256 = "aaed5e834c89775a221cf770d6fe79e1ac05ca5965ee1844d9f83029d48a8ea2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aaed5e834c89775a221cf770d6fe79e1ac05ca5965ee1844d9f83029d48a8ea2",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-gpu-pci-2__7.0.0-14.fc37.x86_64",
        sha256 = "a8833876b3297b4b062386a1b75055ce50d56b4aa3283befbaeeb3cdefd437ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8833876b3297b4b062386a1b75055ce50d56b4aa3283befbaeeb3cdefd437ef",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-gpu-pci-gl-2__7.0.0-14.fc37.x86_64",
        sha256 = "7e6203bcff6f64652aa2fb09d6045cc6c8d50f1c1415b903cf05bce7cb3835b3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e6203bcff6f64652aa2fb09d6045cc6c8d50f1c1415b903cf05bce7cb3835b3",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-vga-2__7.0.0-14.fc37.x86_64",
        sha256 = "0f05303d3437e51cd4b8b8c4bd75815eda557d7cc819629be1d5ea6f7a78a84e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f05303d3437e51cd4b8b8c4bd75815eda557d7cc819629be1d5ea6f7a78a84e",
        ],
    )

    rpm(
        name = "qemu-device-display-virtio-vga-gl-2__7.0.0-14.fc37.x86_64",
        sha256 = "dfdfd7861b8b4009835262260306a3f79080e5f17029e211e67efd20b8f2bb43",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dfdfd7861b8b4009835262260306a3f79080e5f17029e211e67efd20b8f2bb43",
        ],
    )

    rpm(
        name = "qemu-device-usb-host-2__7.0.0-14.fc37.x86_64",
        sha256 = "2596447ac5c576f2d4ba98a0fcbf6236ca8242d86c4558b5b7fb7bf80362fc2e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2596447ac5c576f2d4ba98a0fcbf6236ca8242d86c4558b5b7fb7bf80362fc2e",
        ],
    )

    rpm(
        name = "qemu-device-usb-redirect-2__7.0.0-14.fc37.x86_64",
        sha256 = "e53ede4498509ae09ba709132f59445dd46fc1d657b599db995a5ceea9b3a771",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e53ede4498509ae09ba709132f59445dd46fc1d657b599db995a5ceea9b3a771",
        ],
    )

    rpm(
        name = "qemu-device-usb-smartcard-2__7.0.0-14.fc37.x86_64",
        sha256 = "82a9c2cdf1eb908d2cbfef06fd2f9458f98f425d698e3da0a9c90c6c08909a24",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/82a9c2cdf1eb908d2cbfef06fd2f9458f98f425d698e3da0a9c90c6c08909a24",
        ],
    )

    rpm(
        name = "qemu-img-2__7.0.0-14.fc37.x86_64",
        sha256 = "bc9ecb2ce512392ad96ee1be760dcbee5f250545c50e7a4541d7f359dc7fb812",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bc9ecb2ce512392ad96ee1be760dcbee5f250545c50e7a4541d7f359dc7fb812",
        ],
    )

    rpm(
        name = "qemu-kvm-2__7.0.0-14.fc37.x86_64",
        sha256 = "365b5295ec9c9c4465c285af7081c7e34203c215357638be2f847146596a2637",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/365b5295ec9c9c4465c285af7081c7e34203c215357638be2f847146596a2637",
        ],
    )

    rpm(
        name = "qemu-pr-helper-2__7.0.0-14.fc37.x86_64",
        sha256 = "26bdc07eacb23fa43ea842437a72fb8bc2003ffdea4bf7a36041284d7247ea40",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/26bdc07eacb23fa43ea842437a72fb8bc2003ffdea4bf7a36041284d7247ea40",
        ],
    )

    rpm(
        name = "qemu-system-x86-2__7.0.0-14.fc37.x86_64",
        sha256 = "5d56bd81c1e9c94101e6f05a27f64855b75848efb23d32f9065b47ebeea751ff",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5d56bd81c1e9c94101e6f05a27f64855b75848efb23d32f9065b47ebeea751ff",
        ],
    )

    rpm(
        name = "qemu-system-x86-core-2__7.0.0-14.fc37.x86_64",
        sha256 = "bca0444ca9dbbd734259e2332dbae93010cacf7bd8ae957783b5195b0a3e39bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bca0444ca9dbbd734259e2332dbae93010cacf7bd8ae957783b5195b0a3e39bf",
        ],
    )

    rpm(
        name = "qemu-ui-curses-2__7.0.0-14.fc37.x86_64",
        sha256 = "faf00ad6f77b2dd9fd486c049004ec4bc30241473b8b114f20c063efe7e14f22",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/faf00ad6f77b2dd9fd486c049004ec4bc30241473b8b114f20c063efe7e14f22",
        ],
    )

    rpm(
        name = "qemu-ui-egl-headless-2__7.0.0-14.fc37.x86_64",
        sha256 = "5ece56341f4fa064f768580578c0a64902fe6d9fe98466e54a24d79f86ed35a0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5ece56341f4fa064f768580578c0a64902fe6d9fe98466e54a24d79f86ed35a0",
        ],
    )

    rpm(
        name = "qemu-ui-gtk-2__7.0.0-14.fc37.x86_64",
        sha256 = "da395925b1e887ee7704a4e91bb5450226adef0cea6dbfb46e316d0926f38545",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/da395925b1e887ee7704a4e91bb5450226adef0cea6dbfb46e316d0926f38545",
        ],
    )

    rpm(
        name = "qemu-ui-opengl-2__7.0.0-14.fc37.x86_64",
        sha256 = "09c8bfe7d3344496b06da4a9b0b0d91ea332c4527ec308bb72f46eaac5f666d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/09c8bfe7d3344496b06da4a9b0b0d91ea332c4527ec308bb72f46eaac5f666d6",
        ],
    )

    rpm(
        name = "qemu-ui-sdl-2__7.0.0-14.fc37.x86_64",
        sha256 = "f03b7eca8c19324cb91c35c3233fbac737a04be0e8fa18d277d97bd8a65991c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f03b7eca8c19324cb91c35c3233fbac737a04be0e8fa18d277d97bd8a65991c2",
        ],
    )

    rpm(
        name = "qemu-ui-spice-app-2__7.0.0-14.fc37.x86_64",
        sha256 = "32f9d2daa9db767b2e5235022f23f00404e77c73c66bf76dae4faf719862b4ee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32f9d2daa9db767b2e5235022f23f00404e77c73c66bf76dae4faf719862b4ee",
        ],
    )

    rpm(
        name = "qemu-ui-spice-core-2__7.0.0-14.fc37.x86_64",
        sha256 = "86bfb12e746902526b80d731e3fb70ce4fa4f6eb824144482d9610b244ee8f61",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86bfb12e746902526b80d731e3fb70ce4fa4f6eb824144482d9610b244ee8f61",
        ],
    )

    rpm(
        name = "qemu-virtiofsd-2__7.0.0-14.fc37.x86_64",
        sha256 = "5fd6a1f9ec7eb3051a607be829967ee60da97d28dbc0f0f1591d6afc692970d8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5fd6a1f9ec7eb3051a607be829967ee60da97d28dbc0f0f1591d6afc692970d8",
        ],
    )

    rpm(
        name = "quota-1__4.06-8.fc37.x86_64",
        sha256 = "12c24b81220187755be28d4f3fecad92a04a7a20544a4674a8b05a34aa384e3b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/12c24b81220187755be28d4f3fecad92a04a7a20544a4674a8b05a34aa384e3b",
        ],
    )

    rpm(
        name = "quota-nls-1__4.06-8.fc37.x86_64",
        sha256 = "d06db6c718161747f576c1cffec00b2cbaaafd62371122ac218abcc89ba05310",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d06db6c718161747f576c1cffec00b2cbaaafd62371122ac218abcc89ba05310",
        ],
    )

    rpm(
        name = "readline-0__8.2-2.fc37.x86_64",
        sha256 = "0663e23dc42a7ce84f60f5f3154ba640460a0e5b7158459abf9d5d0986d69d06",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0663e23dc42a7ce84f60f5f3154ba640460a0e5b7158459abf9d5d0986d69d06",
        ],
    )

    rpm(
        name = "rpcbind-0__1.2.6-3.rc2.fc37.x86_64",
        sha256 = "893132bdf1a1ca0cf2e59bc27086dd6d99ce0714c12e052ad363469877c863d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/893132bdf1a1ca0cf2e59bc27086dd6d99ce0714c12e052ad363469877c863d2",
        ],
    )

    rpm(
        name = "rpm-0__4.18.0-1.fc37.x86_64",
        sha256 = "7eb9468d77618514bf861da405e2c85b2411efe81577ebc586fd9c25e5ae4194",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7eb9468d77618514bf861da405e2c85b2411efe81577ebc586fd9c25e5ae4194",
        ],
    )

    rpm(
        name = "rpm-libs-0__4.18.0-1.fc37.x86_64",
        sha256 = "359602208228e24f4d2b4f0ab057ad7ca604ed3f23b0873e7efe395a0c3df25e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/359602208228e24f4d2b4f0ab057ad7ca604ed3f23b0873e7efe395a0c3df25e",
        ],
    )

    rpm(
        name = "rpm-plugin-selinux-0__4.18.0-1.fc37.x86_64",
        sha256 = "1ab8c75a2f9ee929bbfdb722abc00fb96252e4e76c66f1d292cfe01330f3d56a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ab8c75a2f9ee929bbfdb722abc00fb96252e4e76c66f1d292cfe01330f3d56a",
        ],
    )

    rpm(
        name = "seabios-bin-0__1.16.1-2.fc37.x86_64",
        sha256 = "99e4d8e1cd8a8ed6a3b721873b76b44cffad7723ca6e3560932673da71148f2c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/99e4d8e1cd8a8ed6a3b721873b76b44cffad7723ca6e3560932673da71148f2c",
        ],
    )

    rpm(
        name = "seavgabios-bin-0__1.16.1-2.fc37.x86_64",
        sha256 = "5eb1d2fe62e15cd4f9b92c23a03385c87ce4c0bebfcede8444e98d6c86679e37",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5eb1d2fe62e15cd4f9b92c23a03385c87ce4c0bebfcede8444e98d6c86679e37",
        ],
    )

    rpm(
        name = "sed-0__4.8-11.fc37.x86_64",
        sha256 = "231e782077862f4abecf025aa254a9c391a950490ae856261dcfd229863ac80f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/231e782077862f4abecf025aa254a9c391a950490ae856261dcfd229863ac80f",
        ],
    )

    rpm(
        name = "selinux-policy-0__37.19-1.fc37.x86_64",
        sha256 = "8081e5f42dcf1f55cf328ed7b0aca3793ddac5515fc7cacf89f9b826bbda13ba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8081e5f42dcf1f55cf328ed7b0aca3793ddac5515fc7cacf89f9b826bbda13ba",
        ],
    )

    rpm(
        name = "selinux-policy-targeted-0__37.19-1.fc37.x86_64",
        sha256 = "e0575d33b23d5ffd35e679f3a8d5095b2785b4a2f3c58fd74fea4262128519eb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0575d33b23d5ffd35e679f3a8d5095b2785b4a2f3c58fd74fea4262128519eb",
        ],
    )

    rpm(
        name = "setup-0__2.14.1-2.fc37.x86_64",
        sha256 = "15d72b2a44f403b3a7ee9138820a8ce7584f954aeafbb43b1251621bca26f785",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15d72b2a44f403b3a7ee9138820a8ce7584f954aeafbb43b1251621bca26f785",
        ],
    )

    rpm(
        name = "sgabios-bin-1__0.20180715git-9.fc37.x86_64",
        sha256 = "7b18cb14de4338aa5fa937f19a44ac5accceacb18009e5b33afd22755cb7955a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7b18cb14de4338aa5fa937f19a44ac5accceacb18009e5b33afd22755cb7955a",
        ],
    )

    rpm(
        name = "shadow-utils-2__4.12.3-5.fc37.x86_64",
        sha256 = "6e977ea4ed87ef136fc1aa7b01ac06ceea038a537dc902f94ec9da5081290bfd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6e977ea4ed87ef136fc1aa7b01ac06ceea038a537dc902f94ec9da5081290bfd",
        ],
    )

    rpm(
        name = "shared-mime-info-0__2.2-2.fc37.x86_64",
        sha256 = "c471712a2998682347e1680ef1d5a652b48c12502f19e8d39354c4083af91d8a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c471712a2998682347e1680ef1d5a652b48c12502f19e8d39354c4083af91d8a",
        ],
    )

    rpm(
        name = "sheepdog-0__1.0.1-18.fc37.x86_64",
        sha256 = "edf0f215e9c91d8ce0422b2515ccf1b923ff413c817f07d3337f0ffa8b8287a7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/edf0f215e9c91d8ce0422b2515ccf1b923ff413c817f07d3337f0ffa8b8287a7",
        ],
    )

    rpm(
        name = "snappy-0__1.1.9-5.fc37.x86_64",
        sha256 = "46504f3ad77433138805882361af9245a26a74e2b0984f4b35b3509a3b2f91bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/46504f3ad77433138805882361af9245a26a74e2b0984f4b35b3509a3b2f91bf",
        ],
    )

    rpm(
        name = "spice-server-0__0.15.1-1.fc37.x86_64",
        sha256 = "51e2a98325724d799c69e14bdb7c565bc2edf62e102fa8176b263c0686961f1b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/51e2a98325724d799c69e14bdb7c565bc2edf62e102fa8176b263c0686961f1b",
        ],
    )

    rpm(
        name = "sqlite-libs-0__3.40.0-1.fc37.x86_64",
        sha256 = "4d1603de146f9bbe90810100df0afa2efe32e13cc86ed42e32528bc50b8f03dd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d1603de146f9bbe90810100df0afa2efe32e13cc86ed42e32528bc50b8f03dd",
        ],
    )

    rpm(
        name = "swtpm-0__0.7.3-2.20220427gitf2268ee.fc37.x86_64",
        sha256 = "4dd0ae80effe40033c02e3d2b9c4f4824c4faa7f58d7e3ba8c946316dc578ba5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4dd0ae80effe40033c02e3d2b9c4f4824c4faa7f58d7e3ba8c946316dc578ba5",
        ],
    )

    rpm(
        name = "swtpm-libs-0__0.7.3-2.20220427gitf2268ee.fc37.x86_64",
        sha256 = "3b28d0e464f9aefb3c109c56508740c8958a4475235c75ed996f0b80e8caeb0f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b28d0e464f9aefb3c109c56508740c8958a4475235c75ed996f0b80e8caeb0f",
        ],
    )

    rpm(
        name = "swtpm-tools-0__0.7.3-2.20220427gitf2268ee.fc37.x86_64",
        sha256 = "4e6e001a6c6f8793d4b7abd824396ce7560c9524c636f75346f4721461082d1f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4e6e001a6c6f8793d4b7abd824396ce7560c9524c636f75346f4721461082d1f",
        ],
    )

    rpm(
        name = "systemd-0__251.13-6.fc37.x86_64",
        sha256 = "d1191eb3a0149638e395591d2004cd6a5d852e5712ab06c4beb7cfd77e2e2488",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d1191eb3a0149638e395591d2004cd6a5d852e5712ab06c4beb7cfd77e2e2488",
        ],
    )

    rpm(
        name = "systemd-boot-unsigned-0__251.13-6.fc37.x86_64",
        sha256 = "eefefd07bd1ce853a90864be80129a17eaa3711e6961816e886e7089602fab99",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eefefd07bd1ce853a90864be80129a17eaa3711e6961816e886e7089602fab99",
        ],
    )

    rpm(
        name = "systemd-container-0__251.13-6.fc37.x86_64",
        sha256 = "fe88fec0479464db8effe4181c8c4608ae9fe37c3fefa430312de3094543f7da",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe88fec0479464db8effe4181c8c4608ae9fe37c3fefa430312de3094543f7da",
        ],
    )

    rpm(
        name = "systemd-devel-0__251.13-6.fc37.x86_64",
        sha256 = "670b37fe825d24fa6d47131593485f197366bb5190e321ee78e8e3bd467cb687",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/670b37fe825d24fa6d47131593485f197366bb5190e321ee78e8e3bd467cb687",
        ],
    )

    rpm(
        name = "systemd-libs-0__251.13-6.fc37.x86_64",
        sha256 = "20aa751dfa2c65cf5a3cad75867d863ed562694c8d30a317814d3c6e4a0a3e6a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/20aa751dfa2c65cf5a3cad75867d863ed562694c8d30a317814d3c6e4a0a3e6a",
        ],
    )

    rpm(
        name = "systemd-pam-0__251.13-6.fc37.x86_64",
        sha256 = "788dc48f363aaaf40e004475278db7b6408cbe21e20663ca4ed9540b5070ff2a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/788dc48f363aaaf40e004475278db7b6408cbe21e20663ca4ed9540b5070ff2a",
        ],
    )

    rpm(
        name = "systemd-udev-0__251.13-6.fc37.x86_64",
        sha256 = "3c4948ad1f47b599ec254ecf237889e603923265b01ee7228c38b1d89b6095cc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3c4948ad1f47b599ec254ecf237889e603923265b01ee7228c38b1d89b6095cc",
        ],
    )

    rpm(
        name = "trousers-0__0.3.15-7.fc37.x86_64",
        sha256 = "9ec34885483cd25c7ae39b9e5b0af020f6db54123cdc3e38d898badbafb8ca43",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9ec34885483cd25c7ae39b9e5b0af020f6db54123cdc3e38d898badbafb8ca43",
        ],
    )

    rpm(
        name = "trousers-lib-0__0.3.15-7.fc37.x86_64",
        sha256 = "b33af58d16302786d9b793c4f780aeb3b4d96d944868a998eecdcc37e71cfc50",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b33af58d16302786d9b793c4f780aeb3b4d96d944868a998eecdcc37e71cfc50",
        ],
    )

    rpm(
        name = "tzdata-0__2022g-1.fc37.x86_64",
        sha256 = "7ff35c66b3478103fbf3941e933e25f60e41f2b0bfd07d43666b40721211c3bb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7ff35c66b3478103fbf3941e933e25f60e41f2b0bfd07d43666b40721211c3bb",
        ],
    )

    rpm(
        name = "unbound-libs-0__1.17.1-1.fc37.x86_64",
        sha256 = "13f29a4066dde4c0b48de7676275972bd05d8156270e63dadefe9dc2dac82a43",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/13f29a4066dde4c0b48de7676275972bd05d8156270e63dadefe9dc2dac82a43",
        ],
    )

    rpm(
        name = "usbredir-0__0.13.0-1.fc37.x86_64",
        sha256 = "6eff960ec69cf742b10ef009a05d839cba53f24ef9b705e2586f99216e1c6f7b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6eff960ec69cf742b10ef009a05d839cba53f24ef9b705e2586f99216e1c6f7b",
        ],
    )

    rpm(
        name = "userspace-rcu-0__0.13.0-5.fc37.x86_64",
        sha256 = "b7e7d3793988a51ea308b7e9bf9882fa35d1b44a27da67e64c4b5a6eed34f792",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7e7d3793988a51ea308b7e9bf9882fa35d1b44a27da67e64c4b5a6eed34f792",
        ],
    )

    rpm(
        name = "util-linux-0__2.38.1-1.fc37.x86_64",
        sha256 = "23f052850cd509743fae6089181a124ee65c2783d6d15f61ffbae1272f5f67ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/23f052850cd509743fae6089181a124ee65c2783d6d15f61ffbae1272f5f67ef",
        ],
    )

    rpm(
        name = "util-linux-core-0__2.38.1-1.fc37.x86_64",
        sha256 = "f87ad8fc18f4da254966cc6f99b533dc8125e1ec0eaefd5f89a6b6398cb13a34",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f87ad8fc18f4da254966cc6f99b533dc8125e1ec0eaefd5f89a6b6398cb13a34",
        ],
    )

    rpm(
        name = "virglrenderer-0__0.9.1-4.20210420git36391559.fc37.x86_64",
        sha256 = "c021971224778746ac7ae07fbe2e8665d1170c001ca9d4fa3ec9476cc0840907",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c021971224778746ac7ae07fbe2e8665d1170c001ca9d4fa3ec9476cc0840907",
        ],
    )

    rpm(
        name = "vte-profile-0__0.70.3-1.fc37.x86_64",
        sha256 = "d7a6ed96f35f7e8966b80193ec60fee73b4d9836690bfd773a78311c54486079",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d7a6ed96f35f7e8966b80193ec60fee73b4d9836690bfd773a78311c54486079",
        ],
    )

    rpm(
        name = "vte291-0__0.70.3-1.fc37.x86_64",
        sha256 = "ec31b0cc524681033b6a97bd7f8febfdcd28054acd0efafbeb17f7439407b3bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ec31b0cc524681033b6a97bd7f8febfdcd28054acd0efafbeb17f7439407b3bf",
        ],
    )

    rpm(
        name = "which-0__2.21-39.fc37.x86_64",
        sha256 = "4a014a589f3342898dfe96024a04fc0541ddfb48894774af6668ccd6fde5d483",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4a014a589f3342898dfe96024a04fc0541ddfb48894774af6668ccd6fde5d483",
        ],
    )

    rpm(
        name = "xen-libs-0__4.16.3-4.fc37.x86_64",
        sha256 = "990f530caeb686c8cf6340df4ccec898d6b5e524096736f981161dba680701d8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/990f530caeb686c8cf6340df4ccec898d6b5e524096736f981161dba680701d8",
        ],
    )

    rpm(
        name = "xen-licenses-0__4.16.3-4.fc37.x86_64",
        sha256 = "1813e36375f344d061474204699a9b9d4f2e6632c6880c5df82230d053dfd59d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1813e36375f344d061474204699a9b9d4f2e6632c6880c5df82230d053dfd59d",
        ],
    )

    rpm(
        name = "xkeyboard-config-0__2.36-3.fc37.x86_64",
        sha256 = "ddbe0299565bb8cff68a554ff8401fc9127b9fcecfeb030140951b89a551f0fc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ddbe0299565bb8cff68a554ff8401fc9127b9fcecfeb030140951b89a551f0fc",
        ],
    )

    rpm(
        name = "xml-common-0__0.6.3-59.fc37.x86_64",
        sha256 = "4f89fb724b5926f40b8e51a802db808f273e7cd58505dab7a93797a6070261c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4f89fb724b5926f40b8e51a802db808f273e7cd58505dab7a93797a6070261c2",
        ],
    )

    rpm(
        name = "xz-0__5.4.1-1.fc37.x86_64",
        sha256 = "7af1096450d0d76dcd5666e31736f18ff44de9908f2e87d89be88592b176c643",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7af1096450d0d76dcd5666e31736f18ff44de9908f2e87d89be88592b176c643",
        ],
    )

    rpm(
        name = "xz-libs-0__5.4.1-1.fc37.x86_64",
        sha256 = "8c06eef8dd28d6dc1406e65e4eb8ee3db359cf6624729be4e426f6b01c4117fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c06eef8dd28d6dc1406e65e4eb8ee3db359cf6624729be4e426f6b01c4117fd",
        ],
    )

    rpm(
        name = "yajl-0__2.1.0-19.fc37.x86_64",
        sha256 = "b0ca9c6ed5935cde0094694127c13b99a441207eb084f44fb3aa093669c9957c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0ca9c6ed5935cde0094694127c13b99a441207eb084f44fb3aa093669c9957c",
        ],
    )

    rpm(
        name = "zfs-fuse-0__0.7.2.2-23.fc37.x86_64",
        sha256 = "93ca13a70e8ec37f28f52830b52deda1b664fa14a02e5258581b6b1d916cb3e9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93ca13a70e8ec37f28f52830b52deda1b664fa14a02e5258581b6b1d916cb3e9",
        ],
    )

    rpm(
        name = "zita-alsa-pcmi-0__0.6.1-1.fc37.x86_64",
        sha256 = "ef13ccae6f2a6291566c6817cb31f06bc205863acdb629cda4850e434ec16f1e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ef13ccae6f2a6291566c6817cb31f06bc205863acdb629cda4850e434ec16f1e",
        ],
    )

    rpm(
        name = "zita-resampler-0__1.8.0-5.fc37.x86_64",
        sha256 = "e15ca7eaa404f20ef811213b58969d9b32e8e65db8e1c079f45bc840261e9f3c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e15ca7eaa404f20ef811213b58969d9b32e8e65db8e1c079f45bc840261e9f3c",
        ],
    )

    rpm(
        name = "zlib-0__1.2.12-5.fc37.x86_64",
        sha256 = "7b0eda1ad9e9a06e61d9fe41e5e4e0fbdc8427bc252f06a7d29cd7ba81a71a70",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7b0eda1ad9e9a06e61d9fe41e5e4e0fbdc8427bc252f06a7d29cd7ba81a71a70",
        ],
    )
