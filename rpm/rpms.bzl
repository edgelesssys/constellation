""" This file is used to pin / load external RPMs that are used by the project. """

load("@bazeldnf//:deps.bzl", "rpm")

def rpms():
    """ Provides a list of RPMs that are used by the project. """
    rpm(
        name = "SDL2-0__2.26.3-1.fc38.x86_64",
        sha256 = "d18288c462ab008a8a44d21132b29e53facce60f64c0670f2296b56490fc1492",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d18288c462ab008a8a44d21132b29e53facce60f64c0670f2296b56490fc1492",
        ],
    )
    rpm(
        name = "SDL2_image-0__2.6.3-1.fc38.x86_64",
        sha256 = "98e37b9f8b4e75d9049214f525ce1ed870e96a0e4cf5d796a68effa93c5ea19c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/98e37b9f8b4e75d9049214f525ce1ed870e96a0e4cf5d796a68effa93c5ea19c",
        ],
    )
    rpm(
        name = "adwaita-cursor-theme-0__44.0-1.fc38.x86_64",
        sha256 = "32fc0f5270f410ef51096196066412e115284a57c3c425dcaf3234c9181f5e6b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32fc0f5270f410ef51096196066412e115284a57c3c425dcaf3234c9181f5e6b",
        ],
    )
    rpm(
        name = "adwaita-icon-theme-0__44.0-1.fc38.x86_64",
        sha256 = "eb20c37c61812727f3792b675dbbfe3f74c23dd3d0d2379c1c518151095fb48d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eb20c37c61812727f3792b675dbbfe3f74c23dd3d0d2379c1c518151095fb48d",
        ],
    )
    rpm(
        name = "alsa-lib-0__1.2.9-1.fc38.x86_64",
        sha256 = "af25e531f21c532dd306fdce6091a568e057477f6a5e6bbf132a5a7a8cd55aa4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af25e531f21c532dd306fdce6091a568e057477f6a5e6bbf132a5a7a8cd55aa4",
        ],
    )

    rpm(
        name = "alternatives-0__1.24-1.fc38.x86_64",
        sha256 = "96a7e71271d334497b5f108f8aaadb4210e488948947b108375eb8636e7118f5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96a7e71271d334497b5f108f8aaadb4210e488948947b108375eb8636e7118f5",
        ],
    )
    rpm(
        name = "at-spi2-atk-0__2.48.3-1.fc38.x86_64",
        sha256 = "3824b264d786eb406c0562f0cfae13c89edd8f24a7eae0b1c91f05a5fca012da",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3824b264d786eb406c0562f0cfae13c89edd8f24a7eae0b1c91f05a5fca012da",
        ],
    )
    rpm(
        name = "at-spi2-core-0__2.48.3-1.fc38.x86_64",
        sha256 = "7be7133917226dfc59c94f033f3bb5ae737f1e3fb7f490f59d5235a4b5153295",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7be7133917226dfc59c94f033f3bb5ae737f1e3fb7f490f59d5235a4b5153295",
        ],
    )
    rpm(
        name = "atk-0__2.48.3-1.fc38.x86_64",
        sha256 = "38e3aedfca8d5e9dfb89b6b5bdb16f61f6b7554d0bf602e612a5346e95827949",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/38e3aedfca8d5e9dfb89b6b5bdb16f61f6b7554d0bf602e612a5346e95827949",
        ],
    )
    rpm(
        name = "attr-0__2.5.1-6.fc38.x86_64",
        sha256 = "b2e82f496c904742a643cf3a82aa800662d1bb40cec2901fbe4eed5d0c01faed",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b2e82f496c904742a643cf3a82aa800662d1bb40cec2901fbe4eed5d0c01faed",
        ],
    )

    rpm(
        name = "audit-libs-0__3.1.1-1.fc38.x86_64",
        sha256 = "49e693f6812e04450b0c98a108f302a799f5ee21f8000675f59d47e097ad24c7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49e693f6812e04450b0c98a108f302a799f5ee21f8000675f59d47e097ad24c7",
        ],
    )

    rpm(
        name = "authselect-0__1.4.2-2.fc38.x86_64",
        sha256 = "66604d04522a860a1c755a9629b0cb1e25a2c24747945f1f69b41ab8522787b3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/66604d04522a860a1c755a9629b0cb1e25a2c24747945f1f69b41ab8522787b3",
        ],
    )

    rpm(
        name = "authselect-libs-0__1.4.2-2.fc38.x86_64",
        sha256 = "4dc70c4a90ddede4f38d96e8184da5d3113f9404d3cd4b98b45ffd40cc503249",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4dc70c4a90ddede4f38d96e8184da5d3113f9404d3cd4b98b45ffd40cc503249",
        ],
    )
    rpm(
        name = "avahi-libs-0__0.8-22.fc38.x86_64",
        sha256 = "44dd623c85e410086a59340563e083656794d6435e3050288afb5bea5d301df4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/44dd623c85e410086a59340563e083656794d6435e3050288afb5bea5d301df4",
        ],
    )

    rpm(
        name = "basesystem-0__11-15.fc38.x86_64",
        sha256 = "718d95c40b41c2f0ecc8dc2290ebb91b529ba3be7accbad9c30c88e9ce408349",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/718d95c40b41c2f0ecc8dc2290ebb91b529ba3be7accbad9c30c88e9ce408349",
        ],
    )

    rpm(
        name = "bash-0__5.2.15-3.fc38.x86_64",
        sha256 = "961883b6ac18ca54b525209adce0c593f81fd8a7e71bb75bc07724e4ef72bc5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/961883b6ac18ca54b525209adce0c593f81fd8a7e71bb75bc07724e4ef72bc5f",
        ],
    )
    rpm(
        name = "boost-iostreams-0__1.78.0-11.fc38.x86_64",
        sha256 = "5c1782f961be1d0ad4bbb3cdc5f2850b286963567c17d93bc95088d749edc3fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5c1782f961be1d0ad4bbb3cdc5f2850b286963567c17d93bc95088d749edc3fd",
        ],
    )
    rpm(
        name = "boost-system-0__1.78.0-11.fc38.x86_64",
        sha256 = "32a7b56036b3c145a5dc37ffc7ec515f7ad14dd6973d525d96a5d6e9c417ca65",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/32a7b56036b3c145a5dc37ffc7ec515f7ad14dd6973d525d96a5d6e9c417ca65",
        ],
    )
    rpm(
        name = "boost-thread-0__1.78.0-11.fc38.x86_64",
        sha256 = "4430d65ead46bfbde8dda6f30f206971c8f0bd20bfb0f0c578de2a7639f3e77c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4430d65ead46bfbde8dda6f30f206971c8f0bd20bfb0f0c578de2a7639f3e77c",
        ],
    )
    rpm(
        name = "brlapi-0__0.8.4-10.fc38.x86_64",
        sha256 = "42c68010e1c335fe798984db55064176348234a8febf512dba3ed12d8ed51784",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42c68010e1c335fe798984db55064176348234a8febf512dba3ed12d8ed51784",
        ],
    )
    rpm(
        name = "bzip2-0__1.0.8-13.fc38.x86_64",
        sha256 = "8afb9acec0a418447eebab454e992d46369e2379c31ef0d431c15ffb58c3371a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8afb9acec0a418447eebab454e992d46369e2379c31ef0d431c15ffb58c3371a",
        ],
    )

    rpm(
        name = "bzip2-libs-0__1.0.8-13.fc38.x86_64",
        sha256 = "95273426afa05a81e6cf77f941e1156f6a0a3305a7768a02c04a4164280cf876",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/95273426afa05a81e6cf77f941e1156f6a0a3305a7768a02c04a4164280cf876",
        ],
    )

    rpm(
        name = "ca-certificates-0__2023.2.60-2.fc38.x86_64",
        sha256 = "738f573c50a537b556e2fbb8fa4a2c0d22fc2ce4a67344f89d478989cc323ab3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/738f573c50a537b556e2fbb8fa4a2c0d22fc2ce4a67344f89d478989cc323ab3",
        ],
    )
    rpm(
        name = "cairo-0__1.17.8-4.fc38.x86_64",
        sha256 = "684fa6dfb42f2ec09b8763afa246bb24e136619dc2bd6049f3af70ac0a4e40e1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/684fa6dfb42f2ec09b8763afa246bb24e136619dc2bd6049f3af70ac0a4e40e1",
        ],
    )
    rpm(
        name = "cairo-gobject-0__1.17.8-4.fc38.x86_64",
        sha256 = "5b7934b0334aa4745632a3ff784a562313cfa2ebb50def151f852a9e783d6542",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5b7934b0334aa4745632a3ff784a562313cfa2ebb50def151f852a9e783d6542",
        ],
    )
    rpm(
        name = "capstone-0__4.0.2-12.fc38.x86_64",
        sha256 = "69ea882d0a26be6a1ac87586c83791db5d9652c101608984eb02aee624942bee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/69ea882d0a26be6a1ac87586c83791db5d9652c101608984eb02aee624942bee",
        ],
    )
    rpm(
        name = "cdparanoia-libs-0__10.2-41.fc38.x86_64",
        sha256 = "7488183158be4f1a994f3da27ce6205c1b5bc135bda730e6d07e81c2eb6f371c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7488183158be4f1a994f3da27ce6205c1b5bc135bda730e6d07e81c2eb6f371c",
        ],
    )
    rpm(
        name = "checkpolicy-0__3.5-1.fc38.x86_64",
        sha256 = "b55c44ac029b80862c3d3a7d4d106644b4dda59e012abfe821a2cd14fa6d2738",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b55c44ac029b80862c3d3a7d4d106644b4dda59e012abfe821a2cd14fa6d2738",
        ],
    )

    rpm(
        name = "cmake-filesystem-0__3.26.4-4.fc38.x86_64",
        sha256 = "997a5bf1283189176c6debd9700dd3baf3bb2289499bb0c9da7258037a271675",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/997a5bf1283189176c6debd9700dd3baf3bb2289499bb0c9da7258037a271675",
        ],
    )
    rpm(
        name = "colord-libs-0__1.4.6-4.fc38.x86_64",
        sha256 = "f2f2fdbf903c193c34a5dd6432e508af0751859b1e8c9c0fcc1e143296da0bd0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f2f2fdbf903c193c34a5dd6432e508af0751859b1e8c9c0fcc1e143296da0bd0",
        ],
    )

    rpm(
        name = "coreutils-0__9.1-12.fc38.x86_64",
        sha256 = "ea77373517525f6e976fe9c062274b176e86eb378ed82e72a76854c741b2c25e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ea77373517525f6e976fe9c062274b176e86eb378ed82e72a76854c741b2c25e",
        ],
    )

    rpm(
        name = "coreutils-common-0__9.1-12.fc38.x86_64",
        sha256 = "e873cefdeee846f6d1d7e8f00c60bd1714cb2b9d26572d9bf43ea0ce15fd0d75",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e873cefdeee846f6d1d7e8f00c60bd1714cb2b9d26572d9bf43ea0ce15fd0d75",
        ],
    )

    rpm(
        name = "coreutils-single-0__9.1-12.fc38.x86_64",
        sha256 = "1100feca307bb7159a00f336fa5303fd52ded761fc25c77fdc5f095e2add29b3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1100feca307bb7159a00f336fa5303fd52ded761fc25c77fdc5f095e2add29b3",
        ],
    )

    rpm(
        name = "cracklib-0__2.9.7-31.fc38.x86_64",
        sha256 = "1262de076983a540ed42a33160a61cd2723a662c1680935379ccdb0894f60cee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1262de076983a540ed42a33160a61cd2723a662c1680935379ccdb0894f60cee",
        ],
    )

    rpm(
        name = "crypto-policies-0__20230301-1.gita12f7b2.fc38.x86_64",
        sha256 = "6809fe060dc93dfed723bd8faca0f5d65e4f6c62aebd2f9f39f679413b260539",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6809fe060dc93dfed723bd8faca0f5d65e4f6c62aebd2f9f39f679413b260539",
        ],
    )
    rpm(
        name = "crypto-policies-scripts-0__20230301-1.gita12f7b2.fc38.x86_64",
        sha256 = "d1880b8f3e8a5fa943ba033c96dc964fac3e07c34cbd0bb8ac2b0fdcb57abcbc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d1880b8f3e8a5fa943ba033c96dc964fac3e07c34cbd0bb8ac2b0fdcb57abcbc",
        ],
    )

    rpm(
        name = "cryptsetup-devel-0__2.6.1-1.fc38.x86_64",
        sha256 = "b65cb15aeddb8dc5836b491a73692cea5ced80aad375d744b6498ce613c3f07d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b65cb15aeddb8dc5836b491a73692cea5ced80aad375d744b6498ce613c3f07d",
        ],
    )

    rpm(
        name = "cryptsetup-libs-0__2.6.1-1.fc38.x86_64",
        sha256 = "070d86aca4548e9148901881a4ef64d98c5dfd4ea158e9208c650c7f4b225c47",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/070d86aca4548e9148901881a4ef64d98c5dfd4ea158e9208c650c7f4b225c47",
        ],
    )
    rpm(
        name = "cups-libs-1__2.4.4-1.fc38.x86_64",
        sha256 = "4e37c5621a97e13c98087e509bf4b9aa425237efeee9efad25c9bd3def253982",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4e37c5621a97e13c98087e509bf4b9aa425237efeee9efad25c9bd3def253982",
        ],
    )
    rpm(
        name = "curl-minimal-0__8.0.1-2.fc38.x86_64",
        sha256 = "3d163e5d195dbd6b7b5bbb6e2f36289bf8499c5f498a40f8716bc7264e2c05f7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3d163e5d195dbd6b7b5bbb6e2f36289bf8499c5f498a40f8716bc7264e2c05f7",
        ],
    )

    rpm(
        name = "cyrus-sasl-0__2.1.28-9.fc38.x86_64",
        sha256 = "a138c4b2f5083cb7481a9b90c2892af696a062ad89d97c5f493a32ddaa94b26e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a138c4b2f5083cb7481a9b90c2892af696a062ad89d97c5f493a32ddaa94b26e",
        ],
    )

    rpm(
        name = "cyrus-sasl-gssapi-0__2.1.28-9.fc38.x86_64",
        sha256 = "a0ddee89f79caf7f9927b57179ab4f5ddbcd87f5849ad47535c7a20d4882f77b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a0ddee89f79caf7f9927b57179ab4f5ddbcd87f5849ad47535c7a20d4882f77b",
        ],
    )

    rpm(
        name = "cyrus-sasl-lib-0__2.1.28-9.fc38.x86_64",
        sha256 = "b570b4857289cf32ed57d0d84cb861677a649cef7c5cc2498a249d6593eb3614",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b570b4857289cf32ed57d0d84cb861677a649cef7c5cc2498a249d6593eb3614",
        ],
    )
    rpm(
        name = "daxctl-libs-0__77-1.fc38.x86_64",
        sha256 = "fe2a75c8c678eb56d8f5c12ef7da0317c99f16c522d01de090a813f5dafda064",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe2a75c8c678eb56d8f5c12ef7da0317c99f16c522d01de090a813f5dafda064",
        ],
    )

    rpm(
        name = "dbus-1__1.14.8-1.fc38.x86_64",
        sha256 = "c3a909c1fd2267e1d523f1b046e6329dea43c939405c6d23afd3bd89bbc0a71a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c3a909c1fd2267e1d523f1b046e6329dea43c939405c6d23afd3bd89bbc0a71a",
        ],
    )

    rpm(
        name = "dbus-broker-0__33-1.fc38.x86_64",
        sha256 = "6652f5d40acfaeda20555bdd63e964f123e8ab4a52f8c8890e749e06b6bae3e0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6652f5d40acfaeda20555bdd63e964f123e8ab4a52f8c8890e749e06b6bae3e0",
        ],
    )

    rpm(
        name = "dbus-common-1__1.14.8-1.fc38.x86_64",
        sha256 = "2697f6469e610a979b46af75a504d96d3bc4eaae2f3169961433bc312462cddb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2697f6469e610a979b46af75a504d96d3bc4eaae2f3169961433bc312462cddb",
        ],
    )
    rpm(
        name = "dbus-libs-1__1.14.8-1.fc38.x86_64",
        sha256 = "53232b41a42b8f990cc94eb487e5fdc95f1672f48edcd048ae9d27c784e6a6de",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/53232b41a42b8f990cc94eb487e5fdc95f1672f48edcd048ae9d27c784e6a6de",
        ],
    )

    rpm(
        name = "device-mapper-0__1.02.189-2.fc38.x86_64",
        sha256 = "695aab11fda4b6a396c3ca141fb3ccc48af757a70f36ea3a1be2292e90c40d6f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/695aab11fda4b6a396c3ca141fb3ccc48af757a70f36ea3a1be2292e90c40d6f",
        ],
    )
    rpm(
        name = "device-mapper-devel-0__1.02.189-2.fc38.x86_64",
        sha256 = "52da51931c2547d95a727f13dad05e640cba30f7f1518115eb66a97f705646ea",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52da51931c2547d95a727f13dad05e640cba30f7f1518115eb66a97f705646ea",
        ],
    )
    rpm(
        name = "device-mapper-event-0__1.02.189-2.fc38.x86_64",
        sha256 = "e8c1626c4179f11e3f597ad8e73b2ac80537d51de0181da96c0836614fdb8c0a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e8c1626c4179f11e3f597ad8e73b2ac80537d51de0181da96c0836614fdb8c0a",
        ],
    )
    rpm(
        name = "device-mapper-event-libs-0__1.02.189-2.fc38.x86_64",
        sha256 = "0e01de6950ca71089aa2240a6b9d762b7def229e5f95d58695c7cf8670e43b1e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e01de6950ca71089aa2240a6b9d762b7def229e5f95d58695c7cf8670e43b1e",
        ],
    )

    rpm(
        name = "device-mapper-libs-0__1.02.189-2.fc38.x86_64",
        sha256 = "42e8600a8d7e7109de32df6cb6a619b9805f9335bf467a085009f6b82dda6e22",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/42e8600a8d7e7109de32df6cb6a619b9805f9335bf467a085009f6b82dda6e22",
        ],
    )
    rpm(
        name = "device-mapper-multipath-libs-0__0.9.4-2.fc38.x86_64",
        sha256 = "3b5994996497f7a65403b67cd034327243caa8d89750f5e5ff6aaf287f9ed6ec",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b5994996497f7a65403b67cd034327243caa8d89750f5e5ff6aaf287f9ed6ec",
        ],
    )
    rpm(
        name = "device-mapper-persistent-data-0__0.9.0-10.fc38.x86_64",
        sha256 = "fc593550a01333bc07a6b35002ec85b4de40198887d3237984c4e203526db8f2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fc593550a01333bc07a6b35002ec85b4de40198887d3237984c4e203526db8f2",
        ],
    )
    rpm(
        name = "diffutils-0__3.9-1.fc38.x86_64",
        sha256 = "38c134f1c688e9c33ced36cf623e6ce1b0802ff0623b42c69b1cdffb5bea1b9d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/38c134f1c688e9c33ced36cf623e6ce1b0802ff0623b42c69b1cdffb5bea1b9d",
        ],
    )
    rpm(
        name = "dmidecode-1__3.4-3.fc38.x86_64",
        sha256 = "96ce0a0a45030c369297ee608d57f6025f94d10d8b5fe2c2b721b8af64b540c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96ce0a0a45030c369297ee608d57f6025f94d10d8b5fe2c2b721b8af64b540c2",
        ],
    )
    rpm(
        name = "dnsmasq-0__2.89-5.fc38.x86_64",
        sha256 = "02285e6e40f6c1a4f973a2173f6a1786ddea76903f6cda6b0dbf961bf0191a8e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/02285e6e40f6c1a4f973a2173f6a1786ddea76903f6cda6b0dbf961bf0191a8e",
        ],
    )
    rpm(
        name = "duktape-0__2.7.0-2.fc38.x86_64",
        sha256 = "8bcdb9b5ce9b5e19c4f772f52c2c40712491d2a953496e415c380a21a792b050",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8bcdb9b5ce9b5e19c4f772f52c2c40712491d2a953496e415c380a21a792b050",
        ],
    )
    rpm(
        name = "e2fsprogs-libs-0__1.46.5-4.fc38.x86_64",
        sha256 = "1ca049bff926a8ec9b6e0e69f23662934eab96c35d8eda1cf429a2a31f186045",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1ca049bff926a8ec9b6e0e69f23662934eab96c35d8eda1cf429a2a31f186045",
        ],
    )
    rpm(
        name = "ebtables-legacy-0__2.0.11-13.fc38.x86_64",
        sha256 = "59d3a3bef7e21e1282d8db4eabf3c01052a362859d823d5e744bcd23a839c32f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/59d3a3bef7e21e1282d8db4eabf3c01052a362859d823d5e744bcd23a839c32f",
        ],
    )
    rpm(
        name = "edk2-ovmf-0__20230301gitf80f052277c8-26.fc38.x86_64",
        sha256 = "0e10b9d20932885b526256794cd72a23f7ef6919cb081c528d77cd0007a51a8b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e10b9d20932885b526256794cd72a23f7ef6919cb081c528d77cd0007a51a8b",
        ],
    )
    rpm(
        name = "elfutils-default-yama-scope-0__0.189-2.fc38.x86_64",
        sha256 = "4cf1366fcdd86b911fdfab09fd2e87b9334854b666a47199d81db5e3f9caa635",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4cf1366fcdd86b911fdfab09fd2e87b9334854b666a47199d81db5e3f9caa635",
        ],
    )
    rpm(
        name = "elfutils-libelf-0__0.189-2.fc38.x86_64",
        sha256 = "98427b7229a862d099d43f798547ae713f041698c1b274481bf2377afef8fafd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/98427b7229a862d099d43f798547ae713f041698c1b274481bf2377afef8fafd",
        ],
    )
    rpm(
        name = "elfutils-libs-0__0.189-2.fc38.x86_64",
        sha256 = "dc5bc7ddd4b7dcd305f96bf0c160e78d181629405d97f3c7505c815767b51862",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc5bc7ddd4b7dcd305f96bf0c160e78d181629405d97f3c7505c815767b51862",
        ],
    )

    rpm(
        name = "expat-0__2.5.0-2.fc38.x86_64",
        sha256 = "6c64fea958acfb77da5ee23ec1e8d0916c7809ce39987f219927e8f94e5f2755",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6c64fea958acfb77da5ee23ec1e8d0916c7809ce39987f219927e8f94e5f2755",
        ],
    )

    rpm(
        name = "fedora-gpg-keys-0__38-1.x86_64",
        sha256 = "7f7c78f598f7ff131bbe77913b9fc6b7b49d1fce30f2d8505b2d8a85458f519a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7f7c78f598f7ff131bbe77913b9fc6b7b49d1fce30f2d8505b2d8a85458f519a",
        ],
    )

    rpm(
        name = "fedora-release-common-0__38-36.x86_64",
        sha256 = "ac9ede79357b33f0d0c9087333b0dd3e3fd1cf5ccab5c36310b0ec446390e0c7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ac9ede79357b33f0d0c9087333b0dd3e3fd1cf5ccab5c36310b0ec446390e0c7",
        ],
    )

    rpm(
        name = "fedora-release-container-0__38-36.x86_64",
        sha256 = "a2e29da94814409cab5f36afecd9fc14ae85227fbe140d8b10dfc4019726963b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a2e29da94814409cab5f36afecd9fc14ae85227fbe140d8b10dfc4019726963b",
        ],
    )
    rpm(
        name = "fedora-release-identity-basic-0__38-36.x86_64",
        sha256 = "bacf386d747343cb10a2c3847c426d3e044b7de1742b4918e3b1b78cc1b54bc4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bacf386d747343cb10a2c3847c426d3e044b7de1742b4918e3b1b78cc1b54bc4",
        ],
    )

    rpm(
        name = "fedora-release-identity-container-0__38-36.x86_64",
        sha256 = "a8e18d52117d8b079689d99d45ec5442fa4ff5784461508b8be7f1d693e92226",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8e18d52117d8b079689d99d45ec5442fa4ff5784461508b8be7f1d693e92226",
        ],
    )

    rpm(
        name = "fedora-repos-0__38-1.x86_64",
        sha256 = "916b75b58e9a2afe5d53cb73fdabea4d0cd8b1eba9f1213754384d0ccd531e57",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/916b75b58e9a2afe5d53cb73fdabea4d0cd8b1eba9f1213754384d0ccd531e57",
        ],
    )

    rpm(
        name = "filesystem-0__3.18-3.fc38.x86_64",
        sha256 = "b0fc6c55f5989aebf6e71279541206070b32b3b28b708a249bd3bdeaa6c088a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0fc6c55f5989aebf6e71279541206070b32b3b28b708a249bd3bdeaa6c088a4",
        ],
    )
    rpm(
        name = "findutils-1__4.9.0-3.fc38.x86_64",
        sha256 = "79986f917ef1bae7ca2378b16515ba44c19160f5a5eae4f6b697eda160bc26c1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79986f917ef1bae7ca2378b16515ba44c19160f5a5eae4f6b697eda160bc26c1",
        ],
    )
    rpm(
        name = "flac-libs-0__1.4.2-2.fc38.x86_64",
        sha256 = "4b9334a24f064d914b244390636a6ab74b3f595fe653e2cc320594b2c113f0ed",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4b9334a24f064d914b244390636a6ab74b3f595fe653e2cc320594b2c113f0ed",
        ],
    )
    rpm(
        name = "fmt-0__9.1.0-2.fc38.x86_64",
        sha256 = "35fb76f679f5d204b890eeea7b3c2a204aaf1d1c4e989c457917d9252f885d87",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/35fb76f679f5d204b890eeea7b3c2a204aaf1d1c4e989c457917d9252f885d87",
        ],
    )
    rpm(
        name = "fontconfig-0__2.14.2-1.fc38.x86_64",
        sha256 = "b38a6a0f2f9581d169e349e582b0dac3858aaa3d338b6cf1b14c6213d1f33c09",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b38a6a0f2f9581d169e349e582b0dac3858aaa3d338b6cf1b14c6213d1f33c09",
        ],
    )
    rpm(
        name = "fonts-filesystem-1__2.0.5-11.fc38.x86_64",
        sha256 = "c2987d34ab8511b86a565af0491510f1a370175bd6fb99819c1afd50a6851425",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c2987d34ab8511b86a565af0491510f1a370175bd6fb99819c1afd50a6851425",
        ],
    )
    rpm(
        name = "freetype-0__2.13.0-2.fc38.x86_64",
        sha256 = "3881f1c6b5274eaf85732ed106019159805fcdb8c25031200e67a159945e85d4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3881f1c6b5274eaf85732ed106019159805fcdb8c25031200e67a159945e85d4",
        ],
    )
    rpm(
        name = "fribidi-0__1.0.12-3.fc38.x86_64",
        sha256 = "f3ba30c249c30315e695961efcae0a04d6b8cd8fc7e4256e98078e0556e9f3a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f3ba30c249c30315e695961efcae0a04d6b8cd8fc7e4256e98078e0556e9f3a6",
        ],
    )
    rpm(
        name = "fuse-0__2.9.9-16.fc38.x86_64",
        sha256 = "8c70948b04d05290d6441359d461d9f7ec45439b3c97430c41b8a862d59e7b60",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8c70948b04d05290d6441359d461d9f7ec45439b3c97430c41b8a862d59e7b60",
        ],
    )
    rpm(
        name = "fuse-common-0__3.14.1-1.fc38.x86_64",
        sha256 = "d5ae6a7d99826a17d163d9846c2705442b5792a7ccacc5169e4986cdf4b6bae2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d5ae6a7d99826a17d163d9846c2705442b5792a7ccacc5169e4986cdf4b6bae2",
        ],
    )
    rpm(
        name = "fuse-libs-0__2.9.9-16.fc38.x86_64",
        sha256 = "56df47937646df892dad25c6b9ae63d111328febfe86eb93096b8b0a11700b60",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/56df47937646df892dad25c6b9ae63d111328febfe86eb93096b8b0a11700b60",
        ],
    )
    rpm(
        name = "fuse3-libs-0__3.14.1-1.fc38.x86_64",
        sha256 = "f54340fec047cc359a6a164a1ce88d0d7ffcd8f7d6334b50dc5b3d234e3a19ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f54340fec047cc359a6a164a1ce88d0d7ffcd8f7d6334b50dc5b3d234e3a19ac",
        ],
    )

    rpm(
        name = "gawk-0__5.1.1-5.fc38.x86_64",
        sha256 = "e607df61803999da46a199d23d4acadb45b290f29b5b644e583c5526d8081178",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e607df61803999da46a199d23d4acadb45b290f29b5b644e583c5526d8081178",
        ],
    )

    rpm(
        name = "gdbm-libs-1__1.23-3.fc38.x86_64",
        sha256 = "eb376264750aae673aa2a3218c291756023dea980640b30a3efe0f2199ff3889",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/eb376264750aae673aa2a3218c291756023dea980640b30a3efe0f2199ff3889",
        ],
    )
    rpm(
        name = "gdk-pixbuf2-0__2.42.10-2.fc38.x86_64",
        sha256 = "039e8e551cf9678a6bac6fd738847f07a69ff31763dabca690d3ccf4d74f9c22",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/039e8e551cf9678a6bac6fd738847f07a69ff31763dabca690d3ccf4d74f9c22",
        ],
    )
    rpm(
        name = "gdk-pixbuf2-modules-0__2.42.10-2.fc38.x86_64",
        sha256 = "b6163cafc6e60561eb7f7a6105771be449c1d9c86724355d80adf1e6641b7f4f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b6163cafc6e60561eb7f7a6105771be449c1d9c86724355d80adf1e6641b7f4f",
        ],
    )
    rpm(
        name = "gettext-envsubst-0__0.21.1-2.fc38.x86_64",
        sha256 = "276c7af4b13262c0307cf2528c6d79c859ed348db68c0d2780869ea4b179dd02",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/276c7af4b13262c0307cf2528c6d79c859ed348db68c0d2780869ea4b179dd02",
        ],
    )
    rpm(
        name = "gettext-libs-0__0.21.1-2.fc38.x86_64",
        sha256 = "4fb6fcf7eef64a48666ff9fe5a46344087979d9e8fd87be4d58a17cf9c3ef108",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4fb6fcf7eef64a48666ff9fe5a46344087979d9e8fd87be4d58a17cf9c3ef108",
        ],
    )
    rpm(
        name = "gettext-runtime-0__0.21.1-2.fc38.x86_64",
        sha256 = "3837cbe450ceb59e1f9e7469aeb6ec98e08150773b83463725acfb2ebb77a98a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3837cbe450ceb59e1f9e7469aeb6ec98e08150773b83463725acfb2ebb77a98a",
        ],
    )

    rpm(
        name = "glib2-0__2.76.3-1.fc38.x86_64",
        sha256 = "05571a68f7dabff1ba50c42145f51ed4011c70d77ea9b4879f6759ffe837e215",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/05571a68f7dabff1ba50c42145f51ed4011c70d77ea9b4879f6759ffe837e215",
        ],
    )

    rpm(
        name = "glibc-0__2.37-4.fc38.x86_64",
        sha256 = "1d17f327a83ccb5ec42f412c78ae6aaf787bd46f10645b65af43f6f5ae9d8c15",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1d17f327a83ccb5ec42f412c78ae6aaf787bd46f10645b65af43f6f5ae9d8c15",
        ],
    )

    rpm(
        name = "glibc-common-0__2.37-4.fc38.x86_64",
        sha256 = "82f7c74c4f31dad16dec9fc52556f76051979b8dea69877b71aa0ac3060c25d4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/82f7c74c4f31dad16dec9fc52556f76051979b8dea69877b71aa0ac3060c25d4",
        ],
    )
    rpm(
        name = "glibc-langpack-fa-0__2.37-4.fc38.x86_64",
        sha256 = "99497914cc9d0cdca757b3c668b469abbbfb1fc6aa7e3f450fefb7a01ca98329",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/99497914cc9d0cdca757b3c668b469abbbfb1fc6aa7e3f450fefb7a01ca98329",
        ],
    )

    rpm(
        name = "glibc-langpack-ga-0__2.37-4.fc38.x86_64",
        sha256 = "5cc4d02a027c7bb2be9201cdaf813757e589c93a9efbe994c8886421fc70603f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5cc4d02a027c7bb2be9201cdaf813757e589c93a9efbe994c8886421fc70603f",
        ],
    )

    rpm(
        name = "glibc-langpack-sl-0__2.37-4.fc38.x86_64",
        sha256 = "628fee563ed17109ea2398148cfa86e24ea27d6863a189c201c8612063fa7693",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/628fee563ed17109ea2398148cfa86e24ea27d6863a189c201c8612063fa7693",
        ],
    )

    rpm(
        name = "glibc-minimal-langpack-0__2.37-4.fc38.x86_64",
        sha256 = "e3418df9795e4d82bbdd906c94dc4a86bb67e92760875eda33d21dd0379c66fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3418df9795e4d82bbdd906c94dc4a86bb67e92760875eda33d21dd0379c66fd",
        ],
    )
    rpm(
        name = "glibmm2.4-0__2.66.6-1.fc38.x86_64",
        sha256 = "dd085de03dc4e03ec58456ab576ef19ebeec8ed0e9cb54f6de04f2fd4cd608a1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dd085de03dc4e03ec58456ab576ef19ebeec8ed0e9cb54f6de04f2fd4cd608a1",
        ],
    )
    rpm(
        name = "glusterfs-0__11.0-2.fc38.x86_64",
        sha256 = "01a4dca2d0bee76ebd2e33115db45202e1d90f4336f4c0fc6ec7c3ff14eecd11",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/01a4dca2d0bee76ebd2e33115db45202e1d90f4336f4c0fc6ec7c3ff14eecd11",
        ],
    )
    rpm(
        name = "glusterfs-cli-0__11.0-2.fc38.x86_64",
        sha256 = "a6fb716e75c8751e8be87c6fbae863a509a6ec52b624e19d1dc7edc5118c5ade",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a6fb716e75c8751e8be87c6fbae863a509a6ec52b624e19d1dc7edc5118c5ade",
        ],
    )
    rpm(
        name = "glusterfs-client-xlators-0__11.0-2.fc38.x86_64",
        sha256 = "cd99878ca6d7609d3fea047390f799b77c7b708d8da2e05d9c61fcc36e5fb866",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cd99878ca6d7609d3fea047390f799b77c7b708d8da2e05d9c61fcc36e5fb866",
        ],
    )
    rpm(
        name = "glusterfs-fuse-0__11.0-2.fc38.x86_64",
        sha256 = "93e5c85854fa6cb299ad93923bb855617022b45333be0f15364f01f5ca07a19e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/93e5c85854fa6cb299ad93923bb855617022b45333be0f15364f01f5ca07a19e",
        ],
    )

    rpm(
        name = "gmp-1__6.2.1-4.fc38.x86_64",
        sha256 = "69e48c73d962e3798fbc84150dfd79258b32a4f9250ee801d8cdb9540edfc21a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/69e48c73d962e3798fbc84150dfd79258b32a4f9250ee801d8cdb9540edfc21a",
        ],
    )

    rpm(
        name = "gnutls-0__3.8.0-2.fc38.x86_64",
        sha256 = "6170bb84006ad1d74202125a21308b15be0b36ec95e9dbb6552aa35026e966ea",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6170bb84006ad1d74202125a21308b15be0b36ec95e9dbb6552aa35026e966ea",
        ],
    )
    rpm(
        name = "gnutls-dane-0__3.8.0-2.fc38.x86_64",
        sha256 = "e92a9e5213cabb224bdfeada1fcafc002496d5f348003ad38f9635fcb30f3587",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e92a9e5213cabb224bdfeada1fcafc002496d5f348003ad38f9635fcb30f3587",
        ],
    )
    rpm(
        name = "gnutls-utils-0__3.8.0-2.fc38.x86_64",
        sha256 = "19566416719e54a2555d48aa6463e491544a0ffb4ab36e6e472d3ac770b08d30",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/19566416719e54a2555d48aa6463e491544a0ffb4ab36e6e472d3ac770b08d30",
        ],
    )
    rpm(
        name = "google-noto-fonts-common-0__20230201-1.fc38.x86_64",
        sha256 = "1300fabb5a042997633ae8f1afc7828a2f58273392eda5b99d79ed7d6a94c8e9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1300fabb5a042997633ae8f1afc7828a2f58273392eda5b99d79ed7d6a94c8e9",
        ],
    )
    rpm(
        name = "google-noto-sans-vf-fonts-0__20230201-1.fc38.x86_64",
        sha256 = "fbd7e2ca35798a60402b2b6c0387694be511a8e043abb7659063ab963025d917",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fbd7e2ca35798a60402b2b6c0387694be511a8e043abb7659063ab963025d917",
        ],
    )
    rpm(
        name = "gperftools-libs-0__2.9.1-5.fc38.x86_64",
        sha256 = "a37a7d57863e2e6e6117dca3be28af6987d1e77d50d0e97ca865bfc31e54705f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a37a7d57863e2e6e6117dca3be28af6987d1e77d50d0e97ca865bfc31e54705f",
        ],
    )
    rpm(
        name = "graphene-0__1.10.6-5.fc38.x86_64",
        sha256 = "497af48244128200592a1aaa257f74d31bc5a4021d9f37f68487277d6c388a31",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/497af48244128200592a1aaa257f74d31bc5a4021d9f37f68487277d6c388a31",
        ],
    )
    rpm(
        name = "graphite2-0__1.3.14-11.fc38.x86_64",
        sha256 = "55ca9073756a092ccb78f3cc3d8b31f1b5d64b3b93b9cc135ce20b273ad46361",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/55ca9073756a092ccb78f3cc3d8b31f1b5d64b3b93b9cc135ce20b273ad46361",
        ],
    )

    rpm(
        name = "grep-0__3.8-3.fc38.x86_64",
        sha256 = "60ed241ec381a23d03fac733a72132dbdc4ba04c412add78bfc67f1b9f1b4daa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/60ed241ec381a23d03fac733a72132dbdc4ba04c412add78bfc67f1b9f1b4daa",
        ],
    )
    rpm(
        name = "groff-base-0__1.22.4-11.fc38.x86_64",
        sha256 = "8a88f3defca55fc8bdb0b0993482e4edaf010427b2ec5f3e87d4a033fba6f744",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8a88f3defca55fc8bdb0b0993482e4edaf010427b2ec5f3e87d4a033fba6f744",
        ],
    )
    rpm(
        name = "gsm-0__1.0.22-2.fc38.x86_64",
        sha256 = "cdfffc40e2bd5a6332ca5ae655f77f1c157724102b323fbeba7c17eced6bfe0f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cdfffc40e2bd5a6332ca5ae655f77f1c157724102b323fbeba7c17eced6bfe0f",
        ],
    )
    rpm(
        name = "gssproxy-0__0.9.1-5.fc38.x86_64",
        sha256 = "5f45f6f4cc6afbad4e9e42c51c5b97d17663fa4df7a89d2429eb874c730b6f90",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5f45f6f4cc6afbad4e9e42c51c5b97d17663fa4df7a89d2429eb874c730b6f90",
        ],
    )
    rpm(
        name = "gstreamer1-0__1.22.3-2.fc38.x86_64",
        sha256 = "00f4c7e334df3456d79df8bdc9ec4cf9acb87365c8224c1f308b23031c1b33fc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/00f4c7e334df3456d79df8bdc9ec4cf9acb87365c8224c1f308b23031c1b33fc",
        ],
    )
    rpm(
        name = "gstreamer1-plugins-base-0__1.22.3-1.fc38.x86_64",
        sha256 = "24a9db0c8993c7f6d40335b08a71ea2ccdf051baff2dc8d2fa9db1a772705b4f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/24a9db0c8993c7f6d40335b08a71ea2ccdf051baff2dc8d2fa9db1a772705b4f",
        ],
    )
    rpm(
        name = "gtk-update-icon-cache-0__3.24.38-1.fc38.x86_64",
        sha256 = "81f7464b1b92d0f0b9d4772a88d2544038f20e3ff850455162151ae4ddfad919",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/81f7464b1b92d0f0b9d4772a88d2544038f20e3ff850455162151ae4ddfad919",
        ],
    )
    rpm(
        name = "gtk3-0__3.24.38-1.fc38.x86_64",
        sha256 = "c1f92bdec7b22deb94349e8a1aee31bb049161a1ddf788b6d311dc2a071a48e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c1f92bdec7b22deb94349e8a1aee31bb049161a1ddf788b6d311dc2a071a48e3",
        ],
    )

    rpm(
        name = "gzip-0__1.12-3.fc38.x86_64",
        sha256 = "166e842798813a72e4d075dcd9b6e814d7ad3fc8d1b5860281cffe7a68784b25",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/166e842798813a72e4d075dcd9b6e814d7ad3fc8d1b5860281cffe7a68784b25",
        ],
    )
    rpm(
        name = "harfbuzz-0__7.1.0-1.fc38.x86_64",
        sha256 = "7edbfc57da668fe01a84b8eb3739034c323b3d63fb95e9ec1e0fb8d0e542e0b5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7edbfc57da668fe01a84b8eb3739034c323b3d63fb95e9ec1e0fb8d0e542e0b5",
        ],
    )
    rpm(
        name = "hicolor-icon-theme-0__0.17-15.fc38.x86_64",
        sha256 = "182d753057cb601056b01425986cceb964e24ce24b1765f4f4c02f90560c2c60",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/182d753057cb601056b01425986cceb964e24ce24b1765f4f4c02f90560c2c60",
        ],
    )
    rpm(
        name = "highway-0__1.0.4-1.fc38.x86_64",
        sha256 = "c99f8f5fde4c31ca0704373cc604614024887af4263ca1a34e9eacc94e1b7b3d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c99f8f5fde4c31ca0704373cc604614024887af4263ca1a34e9eacc94e1b7b3d",
        ],
    )
    rpm(
        name = "hwdata-0__0.371-1.fc38.x86_64",
        sha256 = "0bf71936e0c30b500dca276e1bf38bb46406034fd3c6e3891aebe7bc1e3cc4b6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0bf71936e0c30b500dca276e1bf38bb46406034fd3c6e3891aebe7bc1e3cc4b6",
        ],
    )
    rpm(
        name = "iproute-0__6.1.0-1.fc38.x86_64",
        sha256 = "987b5e89ee47761a452c32b1753b695762590299c4e940f293d98c77dc866965",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/987b5e89ee47761a452c32b1753b695762590299c4e940f293d98c77dc866965",
        ],
    )
    rpm(
        name = "iproute-tc-0__6.1.0-1.fc38.x86_64",
        sha256 = "e38b8681df9af2732e5cb72d5d54eb26cf95c8908efed9f4b6fb92ab53acc74a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e38b8681df9af2732e5cb72d5d54eb26cf95c8908efed9f4b6fb92ab53acc74a",
        ],
    )
    rpm(
        name = "iptables-legacy-0__1.8.9-4.fc38.x86_64",
        sha256 = "ee2e072b955ef8e80997b53be022bb55abc02006c44f41c38694c62c734dfee2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee2e072b955ef8e80997b53be022bb55abc02006c44f41c38694c62c734dfee2",
        ],
    )
    rpm(
        name = "iptables-legacy-libs-0__1.8.9-4.fc38.x86_64",
        sha256 = "491cbd3615967da69697e4fd666bbbb00cc201ba814c4be4f3f46bb35a47ba0e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/491cbd3615967da69697e4fd666bbbb00cc201ba814c4be4f3f46bb35a47ba0e",
        ],
    )
    rpm(
        name = "iptables-libs-0__1.8.9-4.fc38.x86_64",
        sha256 = "79f0bbdb2dd45dffe1841f82ee86113172b845bb766aa16b1249d04784e6fd87",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79f0bbdb2dd45dffe1841f82ee86113172b845bb766aa16b1249d04784e6fd87",
        ],
    )
    rpm(
        name = "ipxe-roms-qemu-0__20220210-3.git64113751.fc38.x86_64",
        sha256 = "76924069fe56bb569031f238151258efed76065b6e0721c58b7324c3ba8a681c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76924069fe56bb569031f238151258efed76065b6e0721c58b7324c3ba8a681c",
        ],
    )
    rpm(
        name = "iscsi-initiator-utils-0__6.2.1.4-10.git2a8f9d8.fc38.x86_64",
        sha256 = "a8e7245d1a13f74755f771c2b357efacee8a93cba8c8ff5721daeffab1090ccc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8e7245d1a13f74755f771c2b357efacee8a93cba8c8ff5721daeffab1090ccc",
        ],
    )
    rpm(
        name = "iscsi-initiator-utils-iscsiuio-0__6.2.1.4-10.git2a8f9d8.fc38.x86_64",
        sha256 = "79efb3ae89cb6a329dca19f310a8929dbb8b9aac0e0734e3ad5fa7036d4436d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79efb3ae89cb6a329dca19f310a8929dbb8b9aac0e0734e3ad5fa7036d4436d6",
        ],
    )
    rpm(
        name = "isns-utils-libs-0__0.101-6.fc38.x86_64",
        sha256 = "2b96fc893eda39e55df3ebb9c3b004c0bc898876ae8a54cb25c9991a63a0c6ab",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2b96fc893eda39e55df3ebb9c3b004c0bc898876ae8a54cb25c9991a63a0c6ab",
        ],
    )
    rpm(
        name = "iso-codes-0__4.13.0-1.fc38.x86_64",
        sha256 = "6d3b93923bc6c2b8361df57ec6493e36f5c0ce56bbbf45070ddb0f2067bf1c38",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6d3b93923bc6c2b8361df57ec6493e36f5c0ce56bbbf45070ddb0f2067bf1c38",
        ],
    )
    rpm(
        name = "jack-audio-connection-kit-0__1.9.22-1.fc38.x86_64",
        sha256 = "e1f0ed4d703accd77ab5b3763c1a8decf2c18384495ce781e5a5f1fe924988d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e1f0ed4d703accd77ab5b3763c1a8decf2c18384495ce781e5a5f1fe924988d2",
        ],
    )
    rpm(
        name = "jbigkit-libs-0__2.1-25.fc38.x86_64",
        sha256 = "37015135372d4f9b3fb6a5f92a23c6624c6d9b8c0f1ef1e9a72def1e18e37fe7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/37015135372d4f9b3fb6a5f92a23c6624c6d9b8c0f1ef1e9a72def1e18e37fe7",
        ],
    )

    rpm(
        name = "json-c-0__0.16-4.fc38.x86_64",
        sha256 = "736182ae69e03a19be60ed57486990f9b88cd06eeecb5e06cebc7f4b64ab0f5d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/736182ae69e03a19be60ed57486990f9b88cd06eeecb5e06cebc7f4b64ab0f5d",
        ],
    )
    rpm(
        name = "json-c-devel-0__0.16-4.fc38.x86_64",
        sha256 = "ab6a1edb5737d4c142560a90974b4737765593f9dd469ef48675590c3a5bd931",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ab6a1edb5737d4c142560a90974b4737765593f9dd469ef48675590c3a5bd931",
        ],
    )
    rpm(
        name = "json-glib-0__1.6.6-4.fc38.x86_64",
        sha256 = "c2ae86002363a331661c72eec5cad90a100216e46176d4bdbe2c20e465c1d73c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c2ae86002363a331661c72eec5cad90a100216e46176d4bdbe2c20e465c1d73c",
        ],
    )
    rpm(
        name = "kbd-0__2.5.1-5.fc38.x86_64",
        sha256 = "8a3a4007d921c6a9b372e4573520866b445c18476f907b023ea19b436436ec33",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8a3a4007d921c6a9b372e4573520866b445c18476f907b023ea19b436436ec33",
        ],
    )
    rpm(
        name = "kbd-legacy-0__2.5.1-5.fc38.x86_64",
        sha256 = "97fcab8f93c5213714e0a9d15e48440c567d2ac39e4ed12743198268082f0583",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97fcab8f93c5213714e0a9d15e48440c567d2ac39e4ed12743198268082f0583",
        ],
    )
    rpm(
        name = "kbd-misc-0__2.5.1-5.fc38.x86_64",
        sha256 = "2ea5dfd9d8c8fe7f61022b41fd6f8a2653692acf30253f45e1dc67def1648f32",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2ea5dfd9d8c8fe7f61022b41fd6f8a2653692acf30253f45e1dc67def1648f32",
        ],
    )
    rpm(
        name = "keyutils-0__1.6.1-6.fc38.x86_64",
        sha256 = "85d84343cccc7cc2db8362502fec65e3431751fe60066693afd807116b03039f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/85d84343cccc7cc2db8362502fec65e3431751fe60066693afd807116b03039f",
        ],
    )

    rpm(
        name = "keyutils-libs-0__1.6.1-6.fc38.x86_64",
        sha256 = "b66376e78fe54024531d94036d10b22b5050d52a9e65682b36dc7fdf24405997",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b66376e78fe54024531d94036d10b22b5050d52a9e65682b36dc7fdf24405997",
        ],
    )
    rpm(
        name = "kmod-0__30-4.fc38.x86_64",
        sha256 = "7419671c64795b96be18231e2f5d3f95eca8e6a71771863ac035f961041c1d7c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7419671c64795b96be18231e2f5d3f95eca8e6a71771863ac035f961041c1d7c",
        ],
    )

    rpm(
        name = "kmod-libs-0__30-4.fc38.x86_64",
        sha256 = "19f873b67f23362074c03d5825e709ad521278c02e44bdeb30eba6d3bb3a8e0f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/19f873b67f23362074c03d5825e709ad521278c02e44bdeb30eba6d3bb3a8e0f",
        ],
    )

    rpm(
        name = "krb5-libs-0__1.20.1-8.fc38.x86_64",
        sha256 = "e779fca5e3be85efabe2132bdaf87ac852d78187fbbb9e3d9b37eabccafb4e85",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e779fca5e3be85efabe2132bdaf87ac852d78187fbbb9e3d9b37eabccafb4e85",
        ],
    )
    rpm(
        name = "lame-libs-0__3.100-14.fc38.x86_64",
        sha256 = "4d9ddeeee4d78490e0b556f7adca1bdee4a97d009d49d2e2ad44d70f2b3200db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d9ddeeee4d78490e0b556f7adca1bdee4a97d009d49d2e2ad44d70f2b3200db",
        ],
    )
    rpm(
        name = "langpacks-core-font-en-0__3.0-32.fc38.x86_64",
        sha256 = "56f72f1a9e2b939382860e7a8addc5b991475f49caca13de7da1e5ea75bd2118",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/56f72f1a9e2b939382860e7a8addc5b991475f49caca13de7da1e5ea75bd2118",
        ],
    )
    rpm(
        name = "lcms2-0__2.15-1.fc38.x86_64",
        sha256 = "ccddb031d65cf2ff479a8e38df3511a7fa806b7be5e115bdbbf736254fd2b40b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ccddb031d65cf2ff479a8e38df3511a7fa806b7be5e115bdbbf736254fd2b40b",
        ],
    )
    rpm(
        name = "libX11-0__1.8.4-1.fc38.x86_64",
        sha256 = "15ba50ff91da7613a3b3424cdd72e3dabbd62a3d74b6726ff9795054691fd895",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15ba50ff91da7613a3b3424cdd72e3dabbd62a3d74b6726ff9795054691fd895",
        ],
    )
    rpm(
        name = "libX11-common-0__1.8.4-1.fc38.x86_64",
        sha256 = "c46daaa291c2193e1a9758a50078fe884843de14bc8029ce0c0090104d544451",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c46daaa291c2193e1a9758a50078fe884843de14bc8029ce0c0090104d544451",
        ],
    )
    rpm(
        name = "libX11-xcb-0__1.8.4-1.fc38.x86_64",
        sha256 = "81c5f503ad73acabed5bcbca53a39782ed9c3f8905dcf572aeb710c5604bcd45",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/81c5f503ad73acabed5bcbca53a39782ed9c3f8905dcf572aeb710c5604bcd45",
        ],
    )
    rpm(
        name = "libXau-0__1.0.11-2.fc38.x86_64",
        sha256 = "6adfb391ac7b753e103762a43837276236abbf5771d99ae214290e9423e88bfb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6adfb391ac7b753e103762a43837276236abbf5771d99ae214290e9423e88bfb",
        ],
    )
    rpm(
        name = "libXcomposite-0__0.4.5-9.fc38.x86_64",
        sha256 = "5f4bfef0759c2bff4eff1f25e1f3fbb5bff65af725e34ce35ab9f90bfd5e9d91",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5f4bfef0759c2bff4eff1f25e1f3fbb5bff65af725e34ce35ab9f90bfd5e9d91",
        ],
    )
    rpm(
        name = "libXcursor-0__1.2.1-3.fc38.x86_64",
        sha256 = "f534bd49d1fb78321e26c048b54d48faf959e82a591202916b1e42de6e5e1252",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f534bd49d1fb78321e26c048b54d48faf959e82a591202916b1e42de6e5e1252",
        ],
    )
    rpm(
        name = "libXdamage-0__1.1.5-9.fc38.x86_64",
        sha256 = "d35bfdb0291211c4efd03636866ea4563480214fd101aec4e41a9f842d8b4a9c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d35bfdb0291211c4efd03636866ea4563480214fd101aec4e41a9f842d8b4a9c",
        ],
    )
    rpm(
        name = "libXext-0__1.3.5-2.fc38.x86_64",
        sha256 = "487f8391033a854782006c4379e6c4e4914656b5d743199a0661c56880446aa0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/487f8391033a854782006c4379e6c4e4914656b5d743199a0661c56880446aa0",
        ],
    )
    rpm(
        name = "libXfixes-0__6.0.0-5.fc38.x86_64",
        sha256 = "0f2fe556d3a16d51d662866edc96457ec9ad67cada554afc8d789319e9ba05d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f2fe556d3a16d51d662866edc96457ec9ad67cada554afc8d789319e9ba05d6",
        ],
    )
    rpm(
        name = "libXft-0__2.3.8-2.fc38.x86_64",
        sha256 = "51deb78d07056fa601af5e29653137244e08e849c4e2106202adbc1018c8d1ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/51deb78d07056fa601af5e29653137244e08e849c4e2106202adbc1018c8d1ef",
        ],
    )
    rpm(
        name = "libXi-0__1.8.1-1.fc38.x86_64",
        sha256 = "a8d04e730b4a0aaf9316da8e9ab2e416bda3279a61c04b14e7c7257613d16855",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8d04e730b4a0aaf9316da8e9ab2e416bda3279a61c04b14e7c7257613d16855",
        ],
    )
    rpm(
        name = "libXinerama-0__1.1.5-2.fc38.x86_64",
        sha256 = "c988ffbe8eb0b64aaabc4f5bfded769ba04220c9c0a10d6c6d1611711d1acfe6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c988ffbe8eb0b64aaabc4f5bfded769ba04220c9c0a10d6c6d1611711d1acfe6",
        ],
    )
    rpm(
        name = "libXrandr-0__1.5.2-10.fc38.x86_64",
        sha256 = "b67fd9f12f43f89e7f98edfc8e1ac695ea673c0f9dd7a4bf7afef0ebff4249d6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b67fd9f12f43f89e7f98edfc8e1ac695ea673c0f9dd7a4bf7afef0ebff4249d6",
        ],
    )
    rpm(
        name = "libXrender-0__0.9.11-2.fc38.x86_64",
        sha256 = "f606d5be4cefddabe67be0cefac437f15295005d24f6695d7d6e7a23b65d0c67",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f606d5be4cefddabe67be0cefac437f15295005d24f6695d7d6e7a23b65d0c67",
        ],
    )
    rpm(
        name = "libXtst-0__1.2.4-2.fc38.x86_64",
        sha256 = "c12c962bd3b37657a597c27a529bcc19c25c66a6964694664f4084e21ab3334b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c12c962bd3b37657a597c27a529bcc19c25c66a6964694664f4084e21ab3334b",
        ],
    )
    rpm(
        name = "libXv-0__1.0.11-18.fc38.x86_64",
        sha256 = "a2fc989d789b6182346f2c4e5df9ae7abdb862c4a9ca6445a0ce24758b20d6bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a2fc989d789b6182346f2c4e5df9ae7abdb862c4a9ca6445a0ce24758b20d6bf",
        ],
    )
    rpm(
        name = "libXxf86vm-0__1.1.5-2.fc38.x86_64",
        sha256 = "0ee442b2bba35669ab595d5dc85a32d5aae819ea2a195bc5b1a903fcdc38e8f5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0ee442b2bba35669ab595d5dc85a32d5aae819ea2a195bc5b1a903fcdc38e8f5",
        ],
    )

    rpm(
        name = "libacl-0__2.3.1-6.fc38.x86_64",
        sha256 = "9b093be8a99bfbae03c2f3dd5435fc9508003f7ef21e4280ff72fe814c1d794e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9b093be8a99bfbae03c2f3dd5435fc9508003f7ef21e4280ff72fe814c1d794e",
        ],
    )
    rpm(
        name = "libaio-0__0.3.111-15.fc38.x86_64",
        sha256 = "8068b025fe7051320d35257c9ef8aef2dbdcd2adf5a5e3af356a49127173c2ec",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8068b025fe7051320d35257c9ef8aef2dbdcd2adf5a5e3af356a49127173c2ec",
        ],
    )
    rpm(
        name = "libarchive-0__3.6.1-4.fc38.x86_64",
        sha256 = "0d0890dba8274458d068b981164111f554b47632bf9c82d1ec41c14697f2b4af",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0d0890dba8274458d068b981164111f554b47632bf9c82d1ec41c14697f2b4af",
        ],
    )

    rpm(
        name = "libargon2-0__20190702-2.fc38.x86_64",
        sha256 = "dd044973c572e64f505f3d00482249b2b4d71369babb5395a22861fd55b21d79",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dd044973c572e64f505f3d00482249b2b4d71369babb5395a22861fd55b21d79",
        ],
    )
    rpm(
        name = "libargon2-devel-0__20190702-2.fc38.x86_64",
        sha256 = "3ff6a13db55b4850c0a16ad2912354bb899fa3afc6512c22ccced2a7c2ece03a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3ff6a13db55b4850c0a16ad2912354bb899fa3afc6512c22ccced2a7c2ece03a",
        ],
    )
    rpm(
        name = "libasyncns-0__0.8-24.fc38.x86_64",
        sha256 = "1e0717d82bd42f18dac15803059da2d0b4434de9b1733905f1b8c32b64ac37ce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1e0717d82bd42f18dac15803059da2d0b4434de9b1733905f1b8c32b64ac37ce",
        ],
    )

    rpm(
        name = "libattr-0__2.5.1-6.fc38.x86_64",
        sha256 = "d78d7bc485f099bb08c9de55dd12ea6a984b948face1f947de6ec805663a96c5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d78d7bc485f099bb08c9de55dd12ea6a984b948face1f947de6ec805663a96c5",
        ],
    )
    rpm(
        name = "libb2-0__0.98.1-8.fc38.x86_64",
        sha256 = "9e73a2b591ebf2915bfbe7f9479498a73c982a4c74e96cc555930022e3ef0aba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9e73a2b591ebf2915bfbe7f9479498a73c982a4c74e96cc555930022e3ef0aba",
        ],
    )
    rpm(
        name = "libbasicobjects-0__0.1.1-53.fc38.x86_64",
        sha256 = "52678fdc28452ecec93ac13d4b137b65f71e8132cb4b73474e28a9cd50d444eb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52678fdc28452ecec93ac13d4b137b65f71e8132cb4b73474e28a9cd50d444eb",
        ],
    )

    rpm(
        name = "libblkid-0__2.38.1-4.fc38.x86_64",
        sha256 = "21b5a1a024c2d1877d2b7271fd3f82424eb0bd6b95395ad3a3dae5776eec8714",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/21b5a1a024c2d1877d2b7271fd3f82424eb0bd6b95395ad3a3dae5776eec8714",
        ],
    )
    rpm(
        name = "libblkid-devel-0__2.38.1-4.fc38.x86_64",
        sha256 = "807a7f7d04fba75790ba6c8068deb36f2f9238a5ad3b787b985500102e667e9e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/807a7f7d04fba75790ba6c8068deb36f2f9238a5ad3b787b985500102e667e9e",
        ],
    )
    rpm(
        name = "libblkio-0__1.2.2-3.fc38.x86_64",
        sha256 = "a529ec22202caf5ee6f7b1df674940df1e196d3ec04e45bc1555d1340b88e1aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a529ec22202caf5ee6f7b1df674940df1e196d3ec04e45bc1555d1340b88e1aa",
        ],
    )
    rpm(
        name = "libbpf-2__1.1.0-2.fc38.x86_64",
        sha256 = "8079443881e764cece2f8f6789b39ebbe43226cde61675bdfae5a5a18a439b5f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8079443881e764cece2f8f6789b39ebbe43226cde61675bdfae5a5a18a439b5f",
        ],
    )
    rpm(
        name = "libbrotli-0__1.0.9-11.fc38.x86_64",
        sha256 = "bdd3f5e8edc77ce3e183134535ec838f033ed3bf0e6802e864c0a6c5fc94b22d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bdd3f5e8edc77ce3e183134535ec838f033ed3bf0e6802e864c0a6c5fc94b22d",
        ],
    )
    rpm(
        name = "libcacard-3__2.8.1-4.fc38.x86_64",
        sha256 = "ef35d559fc653efd0a9dcf65188736aa8b5fa3497fc08d9ac4a14850c18de471",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ef35d559fc653efd0a9dcf65188736aa8b5fa3497fc08d9ac4a14850c18de471",
        ],
    )

    rpm(
        name = "libcap-0__2.48-6.fc38.x86_64",
        sha256 = "b6a2b3872182fe877fcd1dd85ef66282fdeec79fab87157327c9fc6cbd80ab15",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b6a2b3872182fe877fcd1dd85ef66282fdeec79fab87157327c9fc6cbd80ab15",
        ],
    )

    rpm(
        name = "libcap-ng-0__0.8.3-5.fc38.x86_64",
        sha256 = "1e0bee6fd4e234796795cd45185b250d8cf894aef0bb95f2793d9453246e1a4a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1e0bee6fd4e234796795cd45185b250d8cf894aef0bb95f2793d9453246e1a4a",
        ],
    )
    rpm(
        name = "libcloudproviders-0__0.3.1-7.fc38.x86_64",
        sha256 = "b94bd98dbc779000ebe658be0b57a073d16184d8599bba54bd3591396963fdd4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b94bd98dbc779000ebe658be0b57a073d16184d8599bba54bd3591396963fdd4",
        ],
    )
    rpm(
        name = "libcollection-0__0.7.0-53.fc38.x86_64",
        sha256 = "27d5bdf6f2655bd2a286636796d7498a905a54e57e78868903ad0f83aba47599",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/27d5bdf6f2655bd2a286636796d7498a905a54e57e78868903ad0f83aba47599",
        ],
    )

    rpm(
        name = "libcom_err-0__1.46.5-4.fc38.x86_64",
        sha256 = "4ed3e7b6b0727b86ae9af17bd4839c06762a417a263a5d22eb7fcb39714bb480",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ed3e7b6b0727b86ae9af17bd4839c06762a417a263a5d22eb7fcb39714bb480",
        ],
    )
    rpm(
        name = "libconfig-0__1.7.3-5.fc38.x86_64",
        sha256 = "d420ce8d4d3a269a611a82748ae4f80720dc563565e3473f06f123ac280d1b5c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d420ce8d4d3a269a611a82748ae4f80720dc563565e3473f06f123ac280d1b5c",
        ],
    )

    rpm(
        name = "libcurl-minimal-0__8.0.1-2.fc38.x86_64",
        sha256 = "fdee60e5fe219d0f53de987a4c773b36a35a377f301a07f10d08bdf88390d849",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fdee60e5fe219d0f53de987a4c773b36a35a377f301a07f10d08bdf88390d849",
        ],
    )
    rpm(
        name = "libdatrie-0__0.2.13-5.fc38.x86_64",
        sha256 = "3fa854ea46f3707087f4fa29d6be103ff1cd8e0a9146ba9431e86d8b90243064",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3fa854ea46f3707087f4fa29d6be103ff1cd8e0a9146ba9431e86d8b90243064",
        ],
    )

    rpm(
        name = "libdb-0__5.3.28-55.fc38.x86_64",
        sha256 = "d7030af9e9e0dd9afc5b9ee02839c94757d6a4f8ebd3b21e6d81ba6141a76d46",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d7030af9e9e0dd9afc5b9ee02839c94757d6a4f8ebd3b21e6d81ba6141a76d46",
        ],
    )
    rpm(
        name = "libdrm-0__2.4.114-2.fc38.x86_64",
        sha256 = "b7db656f5515d7b608224e197b71a7d6330711cb6de60c1287ea241bfa67c90c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7db656f5515d7b608224e197b71a7d6330711cb6de60c1287ea241bfa67c90c",
        ],
    )

    rpm(
        name = "libeconf-0__0.4.0-5.fc38.x86_64",
        sha256 = "6e57ebf25ad25e7a6809aa0d99b5692598fb009a5438e90a2005f9fae5fd3b13",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6e57ebf25ad25e7a6809aa0d99b5692598fb009a5438e90a2005f9fae5fd3b13",
        ],
    )
    rpm(
        name = "libedit-0__3.1-45.20221030cvs.fc38.x86_64",
        sha256 = "974a64a10a3021de8a440ff4810a720f738951abd5bb944110cb9355d4ae8fa8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/974a64a10a3021de8a440ff4810a720f738951abd5bb944110cb9355d4ae8fa8",
        ],
    )
    rpm(
        name = "libepoxy-0__1.5.10-3.fc38.x86_64",
        sha256 = "ac76a31b89680524c699999ca5ddaf0d97abc88e54e92885626aeb70681177e4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ac76a31b89680524c699999ca5ddaf0d97abc88e54e92885626aeb70681177e4",
        ],
    )

    rpm(
        name = "libevent-0__2.1.12-8.fc38.x86_64",
        sha256 = "e9741c40e94cf45bdc699b950c238646c2d56b3ee7984e748b94d8e6f87ba3cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e9741c40e94cf45bdc699b950c238646c2d56b3ee7984e748b94d8e6f87ba3cd",
        ],
    )

    rpm(
        name = "libfdisk-0__2.38.1-4.fc38.x86_64",
        sha256 = "2fb7ee2d94f7ee34cff49ab28659c07b075ed67ac147f817e19d8ee8e0adbc9c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2fb7ee2d94f7ee34cff49ab28659c07b075ed67ac147f817e19d8ee8e0adbc9c",
        ],
    )
    rpm(
        name = "libfdt-0__1.6.1-7.fc38.x86_64",
        sha256 = "934838feba3ed7bd505885a4f89a2de09812a99d53c2372494495265a34846ea",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/934838feba3ed7bd505885a4f89a2de09812a99d53c2372494495265a34846ea",
        ],
    )
    rpm(
        name = "libffado-0__2.4.7-1.fc38.x86_64",
        sha256 = "63a884c9c96f59ea78ed93b2a3028a844ce4c6595b53835d85c0c2c94e48242b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/63a884c9c96f59ea78ed93b2a3028a844ce4c6595b53835d85c0c2c94e48242b",
        ],
    )

    rpm(
        name = "libffi-0__3.4.4-2.fc38.x86_64",
        sha256 = "098e8ba05482205c70c3510907da71819faf5d40b588f865ad2a77d6eaf4af09",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/098e8ba05482205c70c3510907da71819faf5d40b588f865ad2a77d6eaf4af09",
        ],
    )

    rpm(
        name = "libgcc-0__13.1.1-2.fc38.x86_64",
        sha256 = "e3e7c19626666ff545cd74c0cd854696005f3d7d0b4eb1d1c1ac60404188d981",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e3e7c19626666ff545cd74c0cd854696005f3d7d0b4eb1d1c1ac60404188d981",
        ],
    )
    rpm(
        name = "libgcrypt-0__1.10.2-1.fc38.x86_64",
        sha256 = "ef4b2686134e6be036755ee093aad43eb0ce4a4c84c93a2defb755cfeb398754",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ef4b2686134e6be036755ee093aad43eb0ce4a4c84c93a2defb755cfeb398754",
        ],
    )
    rpm(
        name = "libgfapi0-0__11.0-2.fc38.x86_64",
        sha256 = "45b7f6904aa0399cdc0046b40e8de6620edaded3fef6367af4ce69e402dd7eaa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/45b7f6904aa0399cdc0046b40e8de6620edaded3fef6367af4ce69e402dd7eaa",
        ],
    )
    rpm(
        name = "libgfrpc0-0__11.0-2.fc38.x86_64",
        sha256 = "de00c261b04e48cd017da85ed7c8409574ed098a4f30c66444598784460bd0a8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de00c261b04e48cd017da85ed7c8409574ed098a4f30c66444598784460bd0a8",
        ],
    )
    rpm(
        name = "libgfxdr0-0__11.0-2.fc38.x86_64",
        sha256 = "829394cd77201e962cc5bd55771e9d802993b22f93f02b332e85b13a32ab9786",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/829394cd77201e962cc5bd55771e9d802993b22f93f02b332e85b13a32ab9786",
        ],
    )
    rpm(
        name = "libglusterfs0-0__11.0-2.fc38.x86_64",
        sha256 = "79cc5a0e0fd8a2a678c2da2dad6f4e72fb386ca74b9b7f76ea575d69b19c81e6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/79cc5a0e0fd8a2a678c2da2dad6f4e72fb386ca74b9b7f76ea575d69b19c81e6",
        ],
    )
    rpm(
        name = "libglvnd-1__1.6.0-2.fc38.x86_64",
        sha256 = "2ef7a7cbab0e48206f4299edfd4d3226bf5d8c35153e74132ef8abdae7fb4a16",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2ef7a7cbab0e48206f4299edfd4d3226bf5d8c35153e74132ef8abdae7fb4a16",
        ],
    )
    rpm(
        name = "libglvnd-egl-1__1.6.0-2.fc38.x86_64",
        sha256 = "aecb7afdef1466b15b84d1c2e56afbecf1b9c45698d81d19f518f3b88f68fb7c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aecb7afdef1466b15b84d1c2e56afbecf1b9c45698d81d19f518f3b88f68fb7c",
        ],
    )
    rpm(
        name = "libglvnd-glx-1__1.6.0-2.fc38.x86_64",
        sha256 = "276750555cd24a57a46c56371a659d04bcaa41fdc2acb594fff64664a5a62ac6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/276750555cd24a57a46c56371a659d04bcaa41fdc2acb594fff64664a5a62ac6",
        ],
    )
    rpm(
        name = "libgomp-0__13.1.1-2.fc38.x86_64",
        sha256 = "3b72fe2b2c48226ef90c88920d64a734df584c68d04d10b48bc8a67471a8a627",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b72fe2b2c48226ef90c88920d64a734df584c68d04d10b48bc8a67471a8a627",
        ],
    )
    rpm(
        name = "libgpg-error-0__1.47-1.fc38.x86_64",
        sha256 = "40b98bdd00b44fbf808df1d653ecf0f0ed8549643d52b68da2455c5a878f0021",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/40b98bdd00b44fbf808df1d653ecf0f0ed8549643d52b68da2455c5a878f0021",
        ],
    )
    rpm(
        name = "libgudev-0__237-4.fc38.x86_64",
        sha256 = "9a66e83aeac66f4eee8c1a155ba7d8e968c7dda692dae91b1e20339cf77790cc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9a66e83aeac66f4eee8c1a155ba7d8e968c7dda692dae91b1e20339cf77790cc",
        ],
    )
    rpm(
        name = "libgusb-0__0.4.5-1.fc38.x86_64",
        sha256 = "827891e774fa64f898d0201a497c7b75d6d612843d703b8776e933490f96cb80",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/827891e774fa64f898d0201a497c7b75d6d612843d703b8776e933490f96cb80",
        ],
    )
    rpm(
        name = "libibverbs-0__44.0-3.fc38.x86_64",
        sha256 = "9f2059c5d699f3dd2337f0872968123a06cf56b9f349d58bd64a5ef22a9815b4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f2059c5d699f3dd2337f0872968123a06cf56b9f349d58bd64a5ef22a9815b4",
        ],
    )
    rpm(
        name = "libicu-0__72.1-2.fc38.x86_64",
        sha256 = "13d8977b7c978500596a1c5c64802f6ffc8663819794d463956ff6781681120b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/13d8977b7c978500596a1c5c64802f6ffc8663819794d463956ff6781681120b",
        ],
    )

    rpm(
        name = "libidn2-0__2.3.4-2.fc38.x86_64",
        sha256 = "d3416e2b6c7565d7a607225d86b556398827316ae7ce43280b82630f0a022bc0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d3416e2b6c7565d7a607225d86b556398827316ae7ce43280b82630f0a022bc0",
        ],
    )
    rpm(
        name = "libiec61883-0__1.2.0-31.fc38.x86_64",
        sha256 = "c18bf8c15cf0db0e43ab1bae3f436227145d42fdceba510d6896753e79bd4ce0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c18bf8c15cf0db0e43ab1bae3f436227145d42fdceba510d6896753e79bd4ce0",
        ],
    )
    rpm(
        name = "libini_config-0__1.3.1-53.fc38.x86_64",
        sha256 = "dab788633eb3b4abcda75b3da6042ed7f06d68643b314e3363c36a29085c2fc1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dab788633eb3b4abcda75b3da6042ed7f06d68643b314e3363c36a29085c2fc1",
        ],
    )
    rpm(
        name = "libiscsi-0__1.19.0-7.fc38.x86_64",
        sha256 = "d57852324a31b11da67b68de88316a76142631976ff9786fca4624d3e5d34a6c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d57852324a31b11da67b68de88316a76142631976ff9786fca4624d3e5d34a6c",
        ],
    )
    rpm(
        name = "libjpeg-turbo-0__2.1.4-2.fc38.x86_64",
        sha256 = "8842d6398bcf9e488369ff4eb63c879df8472ccf2219e742ec32f827f004a683",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8842d6398bcf9e488369ff4eb63c879df8472ccf2219e742ec32f827f004a683",
        ],
    )
    rpm(
        name = "libjxl-1__0.7.0-6.fc38.x86_64",
        sha256 = "3dffeb49f059f02f6dee914742b595e5335889d2e6a29a235cbf08771746a2aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3dffeb49f059f02f6dee914742b595e5335889d2e6a29a235cbf08771746a2aa",
        ],
    )
    rpm(
        name = "libmnl-0__1.0.5-2.fc38.x86_64",
        sha256 = "729b80bbf6ca427c21dd28d1f7029b736dc42f3f4f0c43d322ddfa168bc2ce9b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/729b80bbf6ca427c21dd28d1f7029b736dc42f3f4f0c43d322ddfa168bc2ce9b",
        ],
    )

    rpm(
        name = "libmount-0__2.38.1-4.fc38.x86_64",
        sha256 = "14541cda2a4516ad6a17f46be6b7ad85ef5f6508d36f209f2ba7bd45bc1504e2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/14541cda2a4516ad6a17f46be6b7ad85ef5f6508d36f209f2ba7bd45bc1504e2",
        ],
    )
    rpm(
        name = "libnetfilter_conntrack-0__1.0.8-7.fc38.x86_64",
        sha256 = "bcd0820949cafece875dc7fcaa6960c95859adc03ddcc7e9ae4ca5b13849c34e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bcd0820949cafece875dc7fcaa6960c95859adc03ddcc7e9ae4ca5b13849c34e",
        ],
    )
    rpm(
        name = "libnfnetlink-0__1.0.1-23.fc38.x86_64",
        sha256 = "3c981697fe61f23ad41b615b4c3197d023ec70f713395fc3104b837c61b74294",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3c981697fe61f23ad41b615b4c3197d023ec70f713395fc3104b837c61b74294",
        ],
    )
    rpm(
        name = "libnfs-0__4.0.0-8.fc38.x86_64",
        sha256 = "b0bb7e8968dc19bef17c09dac0c0cf7fcf1e4c987b8c6ef2ed6698fc6e673ce6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b0bb7e8968dc19bef17c09dac0c0cf7fcf1e4c987b8c6ef2ed6698fc6e673ce6",
        ],
    )
    rpm(
        name = "libnfsidmap-1__2.6.3-0.fc38.x86_64",
        sha256 = "8863ac538e5caf32df743e744dea21374e94f2770f4254391fb23f547c4b6746",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8863ac538e5caf32df743e744dea21374e94f2770f4254391fb23f547c4b6746",
        ],
    )

    rpm(
        name = "libnghttp2-0__1.52.0-1.fc38.x86_64",
        sha256 = "33631973ecf9e6b23e0b7d07d61100fdc2db33261f4a6c43b5a09791b9455291",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/33631973ecf9e6b23e0b7d07d61100fdc2db33261f4a6c43b5a09791b9455291",
        ],
    )

    rpm(
        name = "libnl3-0__3.7.0-3.fc38.x86_64",
        sha256 = "a9d80e55bd59e26338a7778de28caf9eb3874f8d90574c879bae1302beaa862b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9d80e55bd59e26338a7778de28caf9eb3874f8d90574c879bae1302beaa862b",
        ],
    )

    rpm(
        name = "libnsl2-0__2.0.0-5.fc38.x86_64",
        sha256 = "28697cf1b5cb4d62c3bd154fc24a23d91a84a5bda2f974fb64bdd04e91b6cec5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/28697cf1b5cb4d62c3bd154fc24a23d91a84a5bda2f974fb64bdd04e91b6cec5",
        ],
    )
    rpm(
        name = "libogg-2__1.3.5-5.fc38.x86_64",
        sha256 = "5ba52b63836d2647c0d96eb02953f5d778cfab76af5f403704e012428f9720fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5ba52b63836d2647c0d96eb02953f5d778cfab76af5f403704e012428f9720fd",
        ],
    )
    rpm(
        name = "libpath_utils-0__0.2.1-53.fc38.x86_64",
        sha256 = "734dfb65f21a498b133004cbf297617e7fdf14a3de4d69c73c39790764ea3663",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/734dfb65f21a498b133004cbf297617e7fdf14a3de4d69c73c39790764ea3663",
        ],
    )
    rpm(
        name = "libpcap-14__1.10.4-1.fc38.x86_64",
        sha256 = "f4d87eb23450cd3888af3d984ad623d229e5bea482188d25f996e61a939486bf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f4d87eb23450cd3888af3d984ad623d229e5bea482188d25f996e61a939486bf",
        ],
    )
    rpm(
        name = "libpciaccess-0__0.16-8.fc38.x86_64",
        sha256 = "cc0caaa28e566f7675e705b1c7dd5e7df5d88bc1edcafecee02cd95319331499",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cc0caaa28e566f7675e705b1c7dd5e7df5d88bc1edcafecee02cd95319331499",
        ],
    )

    rpm(
        name = "libpkgconf-0__1.8.0-6.fc38.x86_64",
        sha256 = "6b6c98d21642a18c20c24d2b136b02d9842179eb9e63a10a89b55ac24449f58f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6b6c98d21642a18c20c24d2b136b02d9842179eb9e63a10a89b55ac24449f58f",
        ],
    )
    rpm(
        name = "libpmem-0__1.12.1-3.fc38.x86_64",
        sha256 = "276376c85b96d5641d12b66d08dcb51018776369a721667dfcec9a2ef79d0cd6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/276376c85b96d5641d12b66d08dcb51018776369a721667dfcec9a2ef79d0cd6",
        ],
    )
    rpm(
        name = "libpmemobj-0__1.12.1-3.fc38.x86_64",
        sha256 = "b07a6d474f5ab3e52d068843a37fe36473e63e7a44d607ef8a03c7a6e38b13a1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b07a6d474f5ab3e52d068843a37fe36473e63e7a44d607ef8a03c7a6e38b13a1",
        ],
    )
    rpm(
        name = "libpng-2__1.6.37-14.fc38.x86_64",
        sha256 = "b5860547ef64d6c35aea6db2c4cb0b2a6781e0dd19c492e0600a68d845069bbc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b5860547ef64d6c35aea6db2c4cb0b2a6781e0dd19c492e0600a68d845069bbc",
        ],
    )
    rpm(
        name = "libpsl-0__0.21.2-2.fc38.x86_64",
        sha256 = "e0bccc94a740acf317caa4fa1fae6b0d57442ef4be36341472b7db93d588ec13",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0bccc94a740acf317caa4fa1fae6b0d57442ef4be36341472b7db93d588ec13",
        ],
    )

    rpm(
        name = "libpwquality-0__1.4.5-3.fc38.x86_64",
        sha256 = "aefb7d2d96af03f4d7ac5a132138d383faf858011b1740c48fcd152000f3c617",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/aefb7d2d96af03f4d7ac5a132138d383faf858011b1740c48fcd152000f3c617",
        ],
    )
    rpm(
        name = "librados2-2__17.2.6-3.fc38.x86_64",
        sha256 = "d1c99fe9ba26d9ea6020dafab6d0b86652fea9cdde255651affcfb076bb5f908",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d1c99fe9ba26d9ea6020dafab6d0b86652fea9cdde255651affcfb076bb5f908",
        ],
    )
    rpm(
        name = "libraw1394-0__2.1.2-17.fc38.x86_64",
        sha256 = "3cefd03d4330b1d95fb0b796d43e39f85b1a2da63520db8fcb55119d18af0045",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3cefd03d4330b1d95fb0b796d43e39f85b1a2da63520db8fcb55119d18af0045",
        ],
    )
    rpm(
        name = "librbd1-2__17.2.6-3.fc38.x86_64",
        sha256 = "4db364bd6d51b468a6001d8fca2642e0130ce8d58d0c0074ae6a17d032841076",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4db364bd6d51b468a6001d8fca2642e0130ce8d58d0c0074ae6a17d032841076",
        ],
    )
    rpm(
        name = "librdmacm-0__44.0-3.fc38.x86_64",
        sha256 = "bad707db4866dc6f60efb43292d64ceb035411d2f2a796084dfc0b1b9b54f886",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bad707db4866dc6f60efb43292d64ceb035411d2f2a796084dfc0b1b9b54f886",
        ],
    )
    rpm(
        name = "libref_array-0__0.1.5-53.fc38.x86_64",
        sha256 = "d46de9648ae76d81bcee6fc34430a9e7d73519eaa7944fe6b01ea56c7bace6c1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d46de9648ae76d81bcee6fc34430a9e7d73519eaa7944fe6b01ea56c7bace6c1",
        ],
    )
    rpm(
        name = "libsamplerate-0__0.2.2-4.fc38.x86_64",
        sha256 = "f7b2509ec07c780228e1636f4ea23b53e9dc29337d8463c6f800d0b10d4283ae",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f7b2509ec07c780228e1636f4ea23b53e9dc29337d8463c6f800d0b10d4283ae",
        ],
    )

    rpm(
        name = "libseccomp-0__2.5.3-4.fc38.x86_64",
        sha256 = "dec378b594b79258dd8b44836c5371f316bcf5e4596d53dd84badcb6d00090df",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dec378b594b79258dd8b44836c5371f316bcf5e4596d53dd84badcb6d00090df",
        ],
    )

    rpm(
        name = "libselinux-0__3.5-1.fc38.x86_64",
        sha256 = "790c6d821ff575ad51242ec6832ed61c8a3c4e0ece245c3dee3292d19acb23b7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/790c6d821ff575ad51242ec6832ed61c8a3c4e0ece245c3dee3292d19acb23b7",
        ],
    )
    rpm(
        name = "libselinux-devel-0__3.5-1.fc38.x86_64",
        sha256 = "1f93070b1c3e3ac67b6a853217a838da022b97cb31e7792d1dd720d4c16b3845",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1f93070b1c3e3ac67b6a853217a838da022b97cb31e7792d1dd720d4c16b3845",
        ],
    )
    rpm(
        name = "libselinux-utils-0__3.5-1.fc38.x86_64",
        sha256 = "78a15621e7e3dfb5a65b8b8aa482cf5b07f08bcef217ad29435e299d6c8aec74",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/78a15621e7e3dfb5a65b8b8aa482cf5b07f08bcef217ad29435e299d6c8aec74",
        ],
    )

    rpm(
        name = "libsemanage-0__3.5-2.fc38.x86_64",
        sha256 = "1b6b7ad33391919a3315e398d737a764121e2fc9581f75318a999e02bfc0c7c4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1b6b7ad33391919a3315e398d737a764121e2fc9581f75318a999e02bfc0c7c4",
        ],
    )

    rpm(
        name = "libsepol-0__3.5-1.fc38.x86_64",
        sha256 = "15ec70665f200a5423589539c3253677eb3c15d7d620fd9bdfe2d1e429735198",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/15ec70665f200a5423589539c3253677eb3c15d7d620fd9bdfe2d1e429735198",
        ],
    )
    rpm(
        name = "libsepol-devel-0__3.5-1.fc38.x86_64",
        sha256 = "db8b729ba00380a25c5620b5e4e04eab05f9c3a4c3221c035347baffb2ac6675",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/db8b729ba00380a25c5620b5e4e04eab05f9c3a4c3221c035347baffb2ac6675",
        ],
    )
    rpm(
        name = "libsigc__plus____plus__20-0__2.10.8-3.fc38.x86_64",
        sha256 = "dfdb2632a4edeb2871422e9dd4a3c950f2197a3f61d7f9bdd5e5b3e23f50b566",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dfdb2632a4edeb2871422e9dd4a3c950f2197a3f61d7f9bdd5e5b3e23f50b566",
        ],
    )

    rpm(
        name = "libsigsegv-0__2.14-4.fc38.x86_64",
        sha256 = "ac0a6bf295151973d2e34392a134e246560b19b7351ced244abc1ed81dfe5b8e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ac0a6bf295151973d2e34392a134e246560b19b7351ced244abc1ed81dfe5b8e",
        ],
    )
    rpm(
        name = "libslirp-0__4.7.0-3.fc38.x86_64",
        sha256 = "b35a0d6b1ecb151982b6a9342d00e8da9663e8a6da6b21b7c559634f7f29fd2d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b35a0d6b1ecb151982b6a9342d00e8da9663e8a6da6b21b7c559634f7f29fd2d",
        ],
    )

    rpm(
        name = "libsmartcols-0__2.38.1-4.fc38.x86_64",
        sha256 = "dbf5c73c71c798533cbecfa54ba28c42878c455df8cb382087d8a758c3ffe290",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dbf5c73c71c798533cbecfa54ba28c42878c455df8cb382087d8a758c3ffe290",
        ],
    )
    rpm(
        name = "libsndfile-0__1.1.0-6.fc38.x86_64",
        sha256 = "cf57c7179a76ea2eaace6b2f4d99ffac0a9ec83e0e3790408afa0a6408750d41",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cf57c7179a76ea2eaace6b2f4d99ffac0a9ec83e0e3790408afa0a6408750d41",
        ],
    )
    rpm(
        name = "libsoup3-0__3.4.2-2.fc38.x86_64",
        sha256 = "76df96aa9f460902dea3aabde38bd4debf81efd1d4f4ab5c2724a86cdd756d6e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76df96aa9f460902dea3aabde38bd4debf81efd1d4f4ab5c2724a86cdd756d6e",
        ],
    )

    rpm(
        name = "libssh-0__0.10.5-1.fc38.x86_64",
        sha256 = "a383d65f22be848a24bc8e328d136260f15dd8b0401da273bb27e18228bada4a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a383d65f22be848a24bc8e328d136260f15dd8b0401da273bb27e18228bada4a",
        ],
    )

    rpm(
        name = "libssh-config-0__0.10.5-1.fc38.x86_64",
        sha256 = "2aefc576f770dc60682929c5de6a80b3c6adbe596398f8d5a1fb89ef7070a1eb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2aefc576f770dc60682929c5de6a80b3c6adbe596398f8d5a1fb89ef7070a1eb",
        ],
    )

    rpm(
        name = "libssh2-0__1.10.0-7.fc38.x86_64",
        sha256 = "b7456feebe68aac1a17bbfed9ab9dcbba7955315cdb9e6700ad0323a431ef48e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b7456feebe68aac1a17bbfed9ab9dcbba7955315cdb9e6700ad0323a431ef48e",
        ],
    )
    rpm(
        name = "libstdc__plus____plus__-0__13.1.1-2.fc38.x86_64",
        sha256 = "17b1caf2274070cab464dbbe8aa307678df13937f2f315af4bc2eaeffd5915c9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17b1caf2274070cab464dbbe8aa307678df13937f2f315af4bc2eaeffd5915c9",
        ],
    )
    rpm(
        name = "libstemmer-0__2.2.0-5.fc38.x86_64",
        sha256 = "9080efa5f6010316a571f5c2ce1cbcb563f4dca3033631e435842944e25ddb5d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9080efa5f6010316a571f5c2ce1cbcb563f4dca3033631e435842944e25ddb5d",
        ],
    )

    rpm(
        name = "libtasn1-0__4.19.0-2.fc38.x86_64",
        sha256 = "8b49dd88579f1c37e05780202e81022c9400422b830d9bdd9087161683628b22",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8b49dd88579f1c37e05780202e81022c9400422b830d9bdd9087161683628b22",
        ],
    )
    rpm(
        name = "libthai-0__0.1.29-4.fc38.x86_64",
        sha256 = "d13dac92e89bc12b32b2a35bb8a1dea85d7d55d50c447fc844ffa1f7ac7a8248",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d13dac92e89bc12b32b2a35bb8a1dea85d7d55d50c447fc844ffa1f7ac7a8248",
        ],
    )
    rpm(
        name = "libtheora-1__1.1.1-33.fc38.x86_64",
        sha256 = "ec8d5bafe587bfb126625da3d0bb7ec7796b8fcb6232dfc194b22b31c75abd3e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ec8d5bafe587bfb126625da3d0bb7ec7796b8fcb6232dfc194b22b31c75abd3e",
        ],
    )
    rpm(
        name = "libtiff-0__4.4.0-5.fc38.x86_64",
        sha256 = "6a76aacf5f06457cf98223a1417f692cebc5da573f0025e955b44cd765bcc6f8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6a76aacf5f06457cf98223a1417f692cebc5da573f0025e955b44cd765bcc6f8",
        ],
    )

    rpm(
        name = "libtirpc-0__1.3.3-1.rc1.fc38.x86_64",
        sha256 = "e13941f4a922a76da599fb0e00884e530d9ed8ce5145fb5a54f7337a6af5085e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e13941f4a922a76da599fb0e00884e530d9ed8ce5145fb5a54f7337a6af5085e",
        ],
    )
    rpm(
        name = "libtpms-0__0.9.6-1.fc38.x86_64",
        sha256 = "8f32e6fb00f48a87b955f3b337151e4b19e9eff4675fc394f7ba66014b1c9a7a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8f32e6fb00f48a87b955f3b337151e4b19e9eff4675fc394f7ba66014b1c9a7a",
        ],
    )
    rpm(
        name = "libtracker-sparql-0__3.5.3-1.fc38.x86_64",
        sha256 = "5c31734f64f9ed025e7212dc306cb6acea2f4a08f488739f62bccc99beef32cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5c31734f64f9ed025e7212dc306cb6acea2f4a08f488739f62bccc99beef32cd",
        ],
    )

    rpm(
        name = "libunistring-0__1.1-3.fc38.x86_64",
        sha256 = "c4012952872a08b9662963b13e29f89388ce6e695e68fa8c37eb6e62bad62441",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c4012952872a08b9662963b13e29f89388ce6e695e68fa8c37eb6e62bad62441",
        ],
    )
    rpm(
        name = "libunistring1.0-0__1.0-1.fc38.x86_64",
        sha256 = "cd0e8eb5d983a985f7df99718fde6245997bdf088fb6086442a883ddb9ed03e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cd0e8eb5d983a985f7df99718fde6245997bdf088fb6086442a883ddb9ed03e3",
        ],
    )
    rpm(
        name = "libunwind-0__1.6.2-7.fc38.x86_64",
        sha256 = "1bd1b743bcac0e60d5b741ab452b05ab2f60da45128e1be8fbd54be158ed18d3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1bd1b743bcac0e60d5b741ab452b05ab2f60da45128e1be8fbd54be158ed18d3",
        ],
    )
    rpm(
        name = "liburing-0__2.3-2.fc38.x86_64",
        sha256 = "7df7f844d1f384d72a75458a56d9613573fa418f865e9461d4b3e87c27e3842a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7df7f844d1f384d72a75458a56d9613573fa418f865e9461d4b3e87c27e3842a",
        ],
    )
    rpm(
        name = "libusb1-0__1.0.26-2.fc38.x86_64",
        sha256 = "805da27b46f0d8cca2cf21a30e52401ae61ca472ae7c2d096de1cfb4b7a0d15c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/805da27b46f0d8cca2cf21a30e52401ae61ca472ae7c2d096de1cfb4b7a0d15c",
        ],
    )

    rpm(
        name = "libutempter-0__1.2.1-8.fc38.x86_64",
        sha256 = "c5c409a2d5f8890eeab48b27b9f4f02925a6bbabeb21ee5e45694c7c9009f037",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c5c409a2d5f8890eeab48b27b9f4f02925a6bbabeb21ee5e45694c7c9009f037",
        ],
    )

    rpm(
        name = "libuuid-0__2.38.1-4.fc38.x86_64",
        sha256 = "876ef0556ddeca2c8f56536c80a2f6e0f64357f40bacb92f483adb8a0ff29af2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/876ef0556ddeca2c8f56536c80a2f6e0f64357f40bacb92f483adb8a0ff29af2",
        ],
    )
    rpm(
        name = "libuuid-devel-0__2.38.1-4.fc38.x86_64",
        sha256 = "876e5f62d9017f2efca5f0cf3316517217834417e9db7bd924a4c8d0e4b13336",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/876e5f62d9017f2efca5f0cf3316517217834417e9db7bd924a4c8d0e4b13336",
        ],
    )
    rpm(
        name = "libva-0__2.18.0-1.fc38.x86_64",
        sha256 = "8147d68b2a22081a51b73f4d35ed32867c2498fc5d6ad36c2f33a00193a4b849",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8147d68b2a22081a51b73f4d35ed32867c2498fc5d6ad36c2f33a00193a4b849",
        ],
    )

    rpm(
        name = "libverto-0__0.3.2-5.fc38.x86_64",
        sha256 = "292791eb37bc312e845e777b2e0e3173e2d951c2bfbbda125bc619dced7f40bc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/292791eb37bc312e845e777b2e0e3173e2d951c2bfbbda125bc619dced7f40bc",
        ],
    )
    rpm(
        name = "libverto-libevent-0__0.3.2-5.fc38.x86_64",
        sha256 = "b86025b0a0cf900ae91ccbf008de6663676957c790b7e1774867e03c946061cc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b86025b0a0cf900ae91ccbf008de6663676957c790b7e1774867e03c946061cc",
        ],
    )
    rpm(
        name = "libvirt-client-0__9.0.0-3.fc38.x86_64",
        sha256 = "7dc91d8f667815b76f2f194fcfa868cc71c7f296b6f167a1c9ab754f0da8b0ba",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7dc91d8f667815b76f2f194fcfa868cc71c7f296b6f167a1c9ab754f0da8b0ba",
        ],
    )
    rpm(
        name = "libvirt-daemon-0__9.0.0-3.fc38.x86_64",
        sha256 = "fc6d946c0b66948d5ae7c4b7bdaf163c674067c1378be87a3082cdbc6894222e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fc6d946c0b66948d5ae7c4b7bdaf163c674067c1378be87a3082cdbc6894222e",
        ],
    )
    rpm(
        name = "libvirt-daemon-config-network-0__9.0.0-3.fc38.x86_64",
        sha256 = "54dbefc659e4fb2533034b953928187a54c9b6827b58ff79174280c91a8d56d7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/54dbefc659e4fb2533034b953928187a54c9b6827b58ff79174280c91a8d56d7",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-interface-0__9.0.0-3.fc38.x86_64",
        sha256 = "d4242d9f991d51a93f44d8c1bbf3677f2fa0992c814a94f749120d7a2b219c8a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d4242d9f991d51a93f44d8c1bbf3677f2fa0992c814a94f749120d7a2b219c8a",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-network-0__9.0.0-3.fc38.x86_64",
        sha256 = "c28fa4a8f8335716a452b0734ef85b770484ae64ff699cf904e94a55f6ccc532",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c28fa4a8f8335716a452b0734ef85b770484ae64ff699cf904e94a55f6ccc532",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-nodedev-0__9.0.0-3.fc38.x86_64",
        sha256 = "ca0d8e38ab32df657b814d7f5365564c7f528b18aafb48d2294cbb52b34ceb23",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ca0d8e38ab32df657b814d7f5365564c7f528b18aafb48d2294cbb52b34ceb23",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-nwfilter-0__9.0.0-3.fc38.x86_64",
        sha256 = "dc1561fc06b0189ee0d6e485694e2835ccd8baea88f83c16d9a36e10ee290df6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc1561fc06b0189ee0d6e485694e2835ccd8baea88f83c16d9a36e10ee290df6",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-qemu-0__9.0.0-3.fc38.x86_64",
        sha256 = "04e60005e152c77a5a898f1a12521431e1711564334a819361f4caf5150b11b5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/04e60005e152c77a5a898f1a12521431e1711564334a819361f4caf5150b11b5",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-secret-0__9.0.0-3.fc38.x86_64",
        sha256 = "841028aae5d31d3531daaed2875753309ade6dc72be94039968a8824fa9a7af4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/841028aae5d31d3531daaed2875753309ade6dc72be94039968a8824fa9a7af4",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-0__9.0.0-3.fc38.x86_64",
        sha256 = "a43dc8ce0d5675e994944f133973f8272c55b812aa29df9a01791a746efdabc3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a43dc8ce0d5675e994944f133973f8272c55b812aa29df9a01791a746efdabc3",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-core-0__9.0.0-3.fc38.x86_64",
        sha256 = "b48246557e213bb87c4be7d00cfaa5e5e3d62c5bc32f24bd8e0ace8112cf904c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b48246557e213bb87c4be7d00cfaa5e5e3d62c5bc32f24bd8e0ace8112cf904c",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-disk-0__9.0.0-3.fc38.x86_64",
        sha256 = "90932eadeb443d4e70dadff49f9b1913160964931c7df7207016550adf593ebf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/90932eadeb443d4e70dadff49f9b1913160964931c7df7207016550adf593ebf",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-gluster-0__9.0.0-3.fc38.x86_64",
        sha256 = "38e6dc26b0d4f6dcd1cf95ab5073211cd5987b9f7371f9f520800be7789ab94e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/38e6dc26b0d4f6dcd1cf95ab5073211cd5987b9f7371f9f520800be7789ab94e",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-iscsi-0__9.0.0-3.fc38.x86_64",
        sha256 = "b8331635c52650519ca2ee37b0ae9954b6fd4b905a3f5132b18df9fb4089389d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8331635c52650519ca2ee37b0ae9954b6fd4b905a3f5132b18df9fb4089389d",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-iscsi-direct-0__9.0.0-3.fc38.x86_64",
        sha256 = "0126e9b70bce883935b68bb6e8ff59eec362469ab4eb990413ca601aba7f00e1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0126e9b70bce883935b68bb6e8ff59eec362469ab4eb990413ca601aba7f00e1",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-logical-0__9.0.0-3.fc38.x86_64",
        sha256 = "dc8848dbe4ab5cf0a978b72c5ea59bacf756696fe645aed316426b088fad0d1e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dc8848dbe4ab5cf0a978b72c5ea59bacf756696fe645aed316426b088fad0d1e",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-mpath-0__9.0.0-3.fc38.x86_64",
        sha256 = "3fbf564c3aa176e4a521f5867ecfee6e546845d98e3fe5aa773744bf31dab381",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3fbf564c3aa176e4a521f5867ecfee6e546845d98e3fe5aa773744bf31dab381",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-rbd-0__9.0.0-3.fc38.x86_64",
        sha256 = "de7567b831e286438280d037b4655386e0587a0641f68db7d2d5690b9e9df478",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/de7567b831e286438280d037b4655386e0587a0641f68db7d2d5690b9e9df478",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-scsi-0__9.0.0-3.fc38.x86_64",
        sha256 = "4fbd5d9825f27552379b7ecdc4d5d32aed9aaaaa6baa9596d8590087982c935f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4fbd5d9825f27552379b7ecdc4d5d32aed9aaaaa6baa9596d8590087982c935f",
        ],
    )
    rpm(
        name = "libvirt-daemon-driver-storage-zfs-0__9.0.0-3.fc38.x86_64",
        sha256 = "24d0f73660c150f4b6b34a408d7352ccf82e147823aac8b37d8d0775f6f58f21",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/24d0f73660c150f4b6b34a408d7352ccf82e147823aac8b37d8d0775f6f58f21",
        ],
    )
    rpm(
        name = "libvirt-daemon-kvm-0__9.0.0-3.fc38.x86_64",
        sha256 = "e0ecb4b1865b0b55aa08f81a7cc886e64ba5cfe917eec596b3e4e5176924eb68",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0ecb4b1865b0b55aa08f81a7cc886e64ba5cfe917eec596b3e4e5176924eb68",
        ],
    )

    rpm(
        name = "libvirt-devel-0__9.0.0-3.fc38.x86_64",
        sha256 = "9f001a1621c0860b93ecb69ec867a083cc9038940b9a4469ecd22cc9c9d0cf81",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f001a1621c0860b93ecb69ec867a083cc9038940b9a4469ecd22cc9c9d0cf81",
        ],
    )

    rpm(
        name = "libvirt-libs-0__9.0.0-3.fc38.x86_64",
        sha256 = "5d219495350229becd82552aad619dd64428bbcd5e5cd347585f7181b3dc16f6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5d219495350229becd82552aad619dd64428bbcd5e5cd347585f7181b3dc16f6",
        ],
    )
    rpm(
        name = "libvisual-1__0.4.1-1.fc38.x86_64",
        sha256 = "02b66edd8a18c79464f7327517a631579fca36aea16f68d80cd424d5d10df2db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/02b66edd8a18c79464f7327517a631579fca36aea16f68d80cd424d5d10df2db",
        ],
    )
    rpm(
        name = "libvorbis-1__1.3.7-7.fc38.x86_64",
        sha256 = "0e0143576cb83d8df9deca4af1350e25cd4baec239d22b03d6eef9d33e43dc7f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e0143576cb83d8df9deca4af1350e25cd4baec239d22b03d6eef9d33e43dc7f",
        ],
    )
    rpm(
        name = "libwayland-client-0__1.22.0-1.fc38.x86_64",
        sha256 = "368c97e4298fa4e260d913f55bd5ce9f2ed6ac5bfd11718e29269ddd209317ce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/368c97e4298fa4e260d913f55bd5ce9f2ed6ac5bfd11718e29269ddd209317ce",
        ],
    )
    rpm(
        name = "libwayland-cursor-0__1.22.0-1.fc38.x86_64",
        sha256 = "97a7bb0afeb30d688bec290689b04de2fee940803da07fa83edee8773d9708a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97a7bb0afeb30d688bec290689b04de2fee940803da07fa83edee8773d9708a6",
        ],
    )
    rpm(
        name = "libwayland-egl-0__1.22.0-1.fc38.x86_64",
        sha256 = "1c3baa170cd1bf3b015b2bb662f5e92d6f13d990a3fbc79cebbe1fb85736c0a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1c3baa170cd1bf3b015b2bb662f5e92d6f13d990a3fbc79cebbe1fb85736c0a4",
        ],
    )
    rpm(
        name = "libwayland-server-0__1.22.0-1.fc38.x86_64",
        sha256 = "c8d8f8b164729d05f61cbc8a68cfb881794435ff61e69d8083d58caf2d9b9859",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c8d8f8b164729d05f61cbc8a68cfb881794435ff61e69d8083d58caf2d9b9859",
        ],
    )
    rpm(
        name = "libwebp-0__1.3.0-2.fc38.x86_64",
        sha256 = "1d0c135a21c4c8a9c067cf068700e48c0e994ad71a96ef5fb67ad03cb0d86b51",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1d0c135a21c4c8a9c067cf068700e48c0e994ad71a96ef5fb67ad03cb0d86b51",
        ],
    )

    rpm(
        name = "libwsman1-0__2.7.1-10.fc38.x86_64",
        sha256 = "c971c1379535cc585d3bbdf663a9825a0fa0cadadee4d8f2958b2dac5a03a1d8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c971c1379535cc585d3bbdf663a9825a0fa0cadadee4d8f2958b2dac5a03a1d8",
        ],
    )
    rpm(
        name = "libxcb-0__1.13.1-11.fc38.x86_64",
        sha256 = "6ee063251e12f5fb0fde6d5aee982d9b9d27103335e93fcc72f6b2e829769f05",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6ee063251e12f5fb0fde6d5aee982d9b9d27103335e93fcc72f6b2e829769f05",
        ],
    )

    rpm(
        name = "libxcrypt-0__4.4.35-1.fc38.x86_64",
        sha256 = "5ae6ba32fa05c1c247a17ae88cf9761b89b30ad9de5ff38cce18d38497d6eb4d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5ae6ba32fa05c1c247a17ae88cf9761b89b30ad9de5ff38cce18d38497d6eb4d",
        ],
    )
    rpm(
        name = "libxkbcommon-0__1.5.0-2.fc38.x86_64",
        sha256 = "507ffdb912296768699a70c30169077b531b1612e47041551bfe523a4b7b6c7d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/507ffdb912296768699a70c30169077b531b1612e47041551bfe523a4b7b6c7d",
        ],
    )

    rpm(
        name = "libxml2-0__2.10.4-1.fc38.x86_64",
        sha256 = "13f2ec62e10333000a13123a4cae5ebbda270c32ece03247e45bd2b244e7bba5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/13f2ec62e10333000a13123a4cae5ebbda270c32ece03247e45bd2b244e7bba5",
        ],
    )
    rpm(
        name = "libxml__plus____plus__-0__2.42.2-2.fc38.x86_64",
        sha256 = "91de25e9d5c085babb73869cda1aa4a9607c6cef36c4f98358d1824e9aba0154",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/91de25e9d5c085babb73869cda1aa4a9607c6cef36c4f98358d1824e9aba0154",
        ],
    )
    rpm(
        name = "libxshmfence-0__1.3-12.fc38.x86_64",
        sha256 = "9024d598310ee443a2d82e85dd26293d4533703230249832729b495cf4fec5a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9024d598310ee443a2d82e85dd26293d4533703230249832729b495cf4fec5a4",
        ],
    )

    rpm(
        name = "libzstd-0__1.5.5-1.fc38.x86_64",
        sha256 = "7d9a98372505c9c1dff7dfea558b20a44820fda416a609467790577a848de110",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7d9a98372505c9c1dff7dfea558b20a44820fda416a609467790577a848de110",
        ],
    )
    rpm(
        name = "linux-atm-libs-0__2.5.1-34.fc38.x86_64",
        sha256 = "27958b2623e06faf37e427fd4c9750a7b9df35ce38365a93caae068d24ebc95b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/27958b2623e06faf37e427fd4c9750a7b9df35ce38365a93caae068d24ebc95b",
        ],
    )
    rpm(
        name = "llvm-libs-0__16.0.5-1.fc38.x86_64",
        sha256 = "fd94edffb34fcfe627d8d066e480e4c4b40b3a00ff946cd008db3102317e0279",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fd94edffb34fcfe627d8d066e480e4c4b40b3a00ff946cd008db3102317e0279",
        ],
    )
    rpm(
        name = "lm_sensors-libs-0__3.6.0-13.fc38.x86_64",
        sha256 = "00bb2d073d6ad9c49ead0e0fb69bcb48aeebb2e12e60e11d45ae73b615dc4c3d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/00bb2d073d6ad9c49ead0e0fb69bcb48aeebb2e12e60e11d45ae73b615dc4c3d",
        ],
    )
    rpm(
        name = "lttng-ust-0__2.13.5-2.fc38.x86_64",
        sha256 = "d82a0842af5a5bdabc93a335fa3a4a18d563d2e69ce511bbdbb7fa526b0f24aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d82a0842af5a5bdabc93a335fa3a4a18d563d2e69ce511bbdbb7fa526b0f24aa",
        ],
    )
    rpm(
        name = "lua-libs-0__5.4.4-9.fc38.x86_64",
        sha256 = "f0a48ec36269d83120425b269e47ba5c86d5a9a44e0de2665c1d55c10732d25b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f0a48ec36269d83120425b269e47ba5c86d5a9a44e0de2665c1d55c10732d25b",
        ],
    )
    rpm(
        name = "lvm2-0__2.03.18-2.fc38.x86_64",
        sha256 = "98ca6f23dfcd29089994b48a22bc8703359ad347b868d453df2b3924cb91ddb5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/98ca6f23dfcd29089994b48a22bc8703359ad347b868d453df2b3924cb91ddb5",
        ],
    )
    rpm(
        name = "lvm2-libs-0__2.03.18-2.fc38.x86_64",
        sha256 = "c70094904d3b13adfe8bbc8fee9f0761a2ea184b8ec9f5c667b324b789694991",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c70094904d3b13adfe8bbc8fee9f0761a2ea184b8ec9f5c667b324b789694991",
        ],
    )

    rpm(
        name = "lz4-libs-0__1.9.4-2.fc38.x86_64",
        sha256 = "96a8f495896c0ff7520c2cc5c9c173d134efc9ef6c6b0364bc7533aefb578d41",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96a8f495896c0ff7520c2cc5c9c173d134efc9ef6c6b0364bc7533aefb578d41",
        ],
    )
    rpm(
        name = "lzo-0__2.10-8.fc38.x86_64",
        sha256 = "c0a169c3f1295ace00207e4005a1cebf832a34f013f2ed15ac65af936c0c1037",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c0a169c3f1295ace00207e4005a1cebf832a34f013f2ed15ac65af936c0c1037",
        ],
    )
    rpm(
        name = "lzop-0__1.04-10.fc38.x86_64",
        sha256 = "edb6dc6b6c48697a575892cfa889e77c5e7a935d4943d9bccb03082c4384af9e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/edb6dc6b6c48697a575892cfa889e77c5e7a935d4943d9bccb03082c4384af9e",
        ],
    )
    rpm(
        name = "mdevctl-0__1.2.0-3.fc38.x86_64",
        sha256 = "3d986c9a028d57046ea8edd97aadce3923e71f2b28c620c584977282d3386da1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3d986c9a028d57046ea8edd97aadce3923e71f2b28c620c584977282d3386da1",
        ],
    )
    rpm(
        name = "mesa-dri-drivers-0__23.1.2-1.fc38.x86_64",
        sha256 = "1f49fbfbfbde05acdda86244f8139fc265eb8603eece44d428036ef69949caf3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1f49fbfbfbde05acdda86244f8139fc265eb8603eece44d428036ef69949caf3",
        ],
    )
    rpm(
        name = "mesa-filesystem-0__23.1.2-1.fc38.x86_64",
        sha256 = "e8bdccbeca06e042b2cb03567ef7b8c6c4baeafeb31aee02988aacf2dd9858ca",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e8bdccbeca06e042b2cb03567ef7b8c6c4baeafeb31aee02988aacf2dd9858ca",
        ],
    )
    rpm(
        name = "mesa-libEGL-0__23.1.2-1.fc38.x86_64",
        sha256 = "96809fa333b71b051a87f42e4cadeb1d3da83f09f56af5af8992290b49c75878",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96809fa333b71b051a87f42e4cadeb1d3da83f09f56af5af8992290b49c75878",
        ],
    )
    rpm(
        name = "mesa-libGL-0__23.1.2-1.fc38.x86_64",
        sha256 = "17af9d932e30054b7a0f95f62f28ef8d44a1d562e2549d3cd4d2e15b2dea8f08",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17af9d932e30054b7a0f95f62f28ef8d44a1d562e2549d3cd4d2e15b2dea8f08",
        ],
    )
    rpm(
        name = "mesa-libgbm-0__23.1.2-1.fc38.x86_64",
        sha256 = "ba573a2754927f646e9bb07f67999c3f1d50d13174b2e545d220451224db36f1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ba573a2754927f646e9bb07f67999c3f1d50d13174b2e545d220451224db36f1",
        ],
    )
    rpm(
        name = "mesa-libglapi-0__23.1.2-1.fc38.x86_64",
        sha256 = "0e3caa1c984a69c729a0ef5f23234559af441e688911fa5c1c2627d38f729d38",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0e3caa1c984a69c729a0ef5f23234559af441e688911fa5c1c2627d38f729d38",
        ],
    )
    rpm(
        name = "mpdecimal-0__2.5.1-6.fc38.x86_64",
        sha256 = "22f217f91fc2d2a666304c0b360520b13adde47761baa6fed1663bb514b6faf5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/22f217f91fc2d2a666304c0b360520b13adde47761baa6fed1663bb514b6faf5",
        ],
    )

    rpm(
        name = "mpfr-0__4.1.1-3.fc38.x86_64",
        sha256 = "e7c9b0c39f77c6fdf68ff04d8714c10532907a8a9c3e76fb377afe546247737f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e7c9b0c39f77c6fdf68ff04d8714c10532907a8a9c3e76fb377afe546247737f",
        ],
    )
    rpm(
        name = "mpg123-libs-0__1.31.3-1.fc38.x86_64",
        sha256 = "feac51e23991dbcfdfcacee899cc393be1aeb6b4e8a409df4149b5dabf8b18d2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/feac51e23991dbcfdfcacee899cc393be1aeb6b4e8a409df4149b5dabf8b18d2",
        ],
    )
    rpm(
        name = "ncurses-0__6.4-3.20230114.fc38.x86_64",
        sha256 = "4cf0fef9c5587482e955c8cd130f296981e552dc76b1a709c81349d7ee475eee",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4cf0fef9c5587482e955c8cd130f296981e552dc76b1a709c81349d7ee475eee",
        ],
    )

    rpm(
        name = "ncurses-base-0__6.4-3.20230114.fc38.x86_64",
        sha256 = "602145f27fd017858256c6ee880863ef5be17c6d3c6c1354f7f16f6f6348db57",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/602145f27fd017858256c6ee880863ef5be17c6d3c6c1354f7f16f6f6348db57",
        ],
    )

    rpm(
        name = "ncurses-libs-0__6.4-3.20230114.fc38.x86_64",
        sha256 = "6ce309d9fd208bfff831981ee4298ccb25fa72363cb7464f1da03b8214d4351f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6ce309d9fd208bfff831981ee4298ccb25fa72363cb7464f1da03b8214d4351f",
        ],
    )
    rpm(
        name = "ndctl-libs-0__77-1.fc38.x86_64",
        sha256 = "11e9ef8253b940711e031069e6103528a5d4f1cc83aa4de2ada284459b9b01c5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/11e9ef8253b940711e031069e6103528a5d4f1cc83aa4de2ada284459b9b01c5",
        ],
    )

    rpm(
        name = "nettle-0__3.8-3.fc38.x86_64",
        sha256 = "605d6710ba42104ce0434bb37b0ca9a922a8392c14175bc782f8acb70b94c3aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/605d6710ba42104ce0434bb37b0ca9a922a8392c14175bc782f8acb70b94c3aa",
        ],
    )
    rpm(
        name = "nfs-utils-1__2.6.3-0.fc38.x86_64",
        sha256 = "2422ff75f1de3099418a510ada2146ab7fc28a5c6720783c9cc4a87f7f836763",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2422ff75f1de3099418a510ada2146ab7fc28a5c6720783c9cc4a87f7f836763",
        ],
    )
    rpm(
        name = "nspr-0__4.35.0-7.fc38.x86_64",
        sha256 = "259f0d0f4e4ac85491d73880cb300a1a59854d295b45990f519a992bf10fd600",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/259f0d0f4e4ac85491d73880cb300a1a59854d295b45990f519a992bf10fd600",
        ],
    )
    rpm(
        name = "nss-0__3.90.0-1.fc38.x86_64",
        sha256 = "4ea2af363a603e8966bb9f2682cfe095d21d233b4c0857c0160d7937e6e7d6b9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4ea2af363a603e8966bb9f2682cfe095d21d233b4c0857c0160d7937e6e7d6b9",
        ],
    )
    rpm(
        name = "nss-softokn-0__3.90.0-1.fc38.x86_64",
        sha256 = "74bf3d6139b82aa99d33675151624f14e9c19563091d4c56e3179fec43b83b7b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/74bf3d6139b82aa99d33675151624f14e9c19563091d4c56e3179fec43b83b7b",
        ],
    )
    rpm(
        name = "nss-softokn-freebl-0__3.90.0-1.fc38.x86_64",
        sha256 = "80b406c941b73c36130e88d5bd63f670f829991836abbde89c2d32103247450e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/80b406c941b73c36130e88d5bd63f670f829991836abbde89c2d32103247450e",
        ],
    )
    rpm(
        name = "nss-sysinit-0__3.90.0-1.fc38.x86_64",
        sha256 = "d677418fabeb2f926b29b0342a3846a931f0c333620e67ed687481508e58e26b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d677418fabeb2f926b29b0342a3846a931f0c333620e67ed687481508e58e26b",
        ],
    )
    rpm(
        name = "nss-util-0__3.90.0-1.fc38.x86_64",
        sha256 = "027a6c010b1fdaf9ce108ade9dc56d65ef78e19b854760558b4f8df76459f6a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/027a6c010b1fdaf9ce108ade9dc56d65ef78e19b854760558b4f8df76459f6a4",
        ],
    )

    rpm(
        name = "numactl-libs-0__2.0.16-2.fc38.x86_64",
        sha256 = "2f7ccfe2164b0063349128dd1fba018fe76a679f6ecfa8306f04dc5a5db341c7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2f7ccfe2164b0063349128dd1fba018fe76a679f6ecfa8306f04dc5a5db341c7",
        ],
    )
    rpm(
        name = "numad-0__0.5-38.20150602git.fc38.x86_64",
        sha256 = "848d74f79bb0ba85aadd906a0f932216ed86dc779828fd60a95d1e28bfca2db3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/848d74f79bb0ba85aadd906a0f932216ed86dc779828fd60a95d1e28bfca2db3",
        ],
    )

    rpm(
        name = "openldap-0__2.6.4-1.fc38.x86_64",
        sha256 = "13509309959035c338299a33985d56f10896466367e1f62d4fea98123e74bfc7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/13509309959035c338299a33985d56f10896466367e1f62d4fea98123e74bfc7",
        ],
    )

    rpm(
        name = "openssl-libs-1__3.0.9-1.fc38.x86_64",
        sha256 = "c9984097ed1c330ccf4369b4e0c2e006b2de3211f0b128804b2992dea401f914",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c9984097ed1c330ccf4369b4e0c2e006b2de3211f0b128804b2992dea401f914",
        ],
    )
    rpm(
        name = "openssl1.1-1__1.1.1q-4.fc38.x86_64",
        sha256 = "a40c0cc6ce179c74e50c788917ce8dc58cc8dd9b2b945d0f941836cb2dcb33b6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a40c0cc6ce179c74e50c788917ce8dc58cc8dd9b2b945d0f941836cb2dcb33b6",
        ],
    )
    rpm(
        name = "openssl1.1-devel-1__1.1.1q-4.fc38.x86_64",
        sha256 = "f8985b571eb2cdc9299031268eaa41b658c7244a8e3f34e2838699c391cf63cb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f8985b571eb2cdc9299031268eaa41b658c7244a8e3f34e2838699c391cf63cb",
        ],
    )
    rpm(
        name = "opus-0__1.3.1-12.fc38.x86_64",
        sha256 = "6b551665f914f90435870fba0c26935e41b831169d4da7e224212d20b041a035",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6b551665f914f90435870fba0c26935e41b831169d4da7e224212d20b041a035",
        ],
    )
    rpm(
        name = "orc-0__0.4.33-2.fc38.x86_64",
        sha256 = "94aa513b040927c074d1c1667dd8bb81842a8859182326658835ad2b604459fb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/94aa513b040927c074d1c1667dd8bb81842a8859182326658835ad2b604459fb",
        ],
    )

    rpm(
        name = "p11-kit-0__0.24.1-6.fc38.x86_64",
        sha256 = "8e4afbcb9488d9c6a9bf7d0739173b8757ce33a6f5e00f0ab7ccfcf605ed9273",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8e4afbcb9488d9c6a9bf7d0739173b8757ce33a6f5e00f0ab7ccfcf605ed9273",
        ],
    )

    rpm(
        name = "p11-kit-trust-0__0.24.1-6.fc38.x86_64",
        sha256 = "9030a26ff737b7bbb71d5208feebba1a0b2774d58dfb6016824a042e059642d8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9030a26ff737b7bbb71d5208feebba1a0b2774d58dfb6016824a042e059642d8",
        ],
    )

    rpm(
        name = "pam-0__1.5.2-16.fc38.x86_64",
        sha256 = "065b99f3541fd5f1281be2082b77e48b835a591776e92f2327bb0462c67baed0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/065b99f3541fd5f1281be2082b77e48b835a591776e92f2327bb0462c67baed0",
        ],
    )

    rpm(
        name = "pam-libs-0__1.5.2-16.fc38.x86_64",
        sha256 = "63e970f7b3f8c54e1dff90661c26519f32a4bf7486c40f2dd38d55e40660230e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/63e970f7b3f8c54e1dff90661c26519f32a4bf7486c40f2dd38d55e40660230e",
        ],
    )
    rpm(
        name = "pango-0__1.50.14-1.fc38.x86_64",
        sha256 = "8e940899fad2ce1e7087c01a76a3e7ffe82358005cb86615814f4193278f3deb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8e940899fad2ce1e7087c01a76a3e7ffe82358005cb86615814f4193278f3deb",
        ],
    )
    rpm(
        name = "parted-0__3.5-11.fc38.x86_64",
        sha256 = "8d846f866158409c775656b39e372d59cf224936d29972d3b6d14e40d3b832ca",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8d846f866158409c775656b39e372d59cf224936d29972d3b6d14e40d3b832ca",
        ],
    )

    rpm(
        name = "pcre2-0__10.42-1.fc38.1.x86_64",
        sha256 = "cb1caf3e9a4ddc8343c0757c7a2730bf5de2b5f0b4c9ee7d928609566f64f010",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cb1caf3e9a4ddc8343c0757c7a2730bf5de2b5f0b4c9ee7d928609566f64f010",
        ],
    )
    rpm(
        name = "pcre2-devel-0__10.42-1.fc38.1.x86_64",
        sha256 = "06be31ad18b64b8bd263a94f6877e1d920cddbdbb51e4f8d52bfcdd29ab92542",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/06be31ad18b64b8bd263a94f6877e1d920cddbdbb51e4f8d52bfcdd29ab92542",
        ],
    )

    rpm(
        name = "pcre2-syntax-0__10.42-1.fc38.1.x86_64",
        sha256 = "756f64de1e4673f0f617a9f3f12f74cceef5fc093e309d1b1d5dffef287b7d67",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/756f64de1e4673f0f617a9f3f12f74cceef5fc093e309d1b1d5dffef287b7d67",
        ],
    )
    rpm(
        name = "pcre2-utf16-0__10.42-1.fc38.1.x86_64",
        sha256 = "f40336ca626449af90d951f5b7d23ffa727a527a6ab609e28c0928ff383119a2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f40336ca626449af90d951f5b7d23ffa727a527a6ab609e28c0928ff383119a2",
        ],
    )
    rpm(
        name = "pcre2-utf32-0__10.42-1.fc38.1.x86_64",
        sha256 = "af8be96f217fb4d6374aac1da54b2108f5f9d1609cee97f31c95d9bd30ea2c10",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af8be96f217fb4d6374aac1da54b2108f5f9d1609cee97f31c95d9bd30ea2c10",
        ],
    )
    rpm(
        name = "pcsc-lite-libs-0__1.9.9-3.fc38.x86_64",
        sha256 = "07dc5536982278f38c89517465384ef9f376cd27f0b200806268723993da01ad",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/07dc5536982278f38c89517465384ef9f376cd27f0b200806268723993da01ad",
        ],
    )
    rpm(
        name = "perl-Carp-0__1.52-490.fc38.x86_64",
        sha256 = "2c130bb72ffd3e5cb9d9ad740336c24256c62c1f6108f601dc8a3cd69acabfd4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2c130bb72ffd3e5cb9d9ad740336c24256c62c1f6108f601dc8a3cd69acabfd4",
        ],
    )
    rpm(
        name = "perl-Class-Struct-0__0.66-497.fc38.x86_64",
        sha256 = "84aa5eac729bffd0072d9a0c1cc531770bb17334a72c122c90255cc59053c49e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/84aa5eac729bffd0072d9a0c1cc531770bb17334a72c122c90255cc59053c49e",
        ],
    )
    rpm(
        name = "perl-DynaLoader-0__1.52-497.fc38.x86_64",
        sha256 = "4dc2b066cefd16876992ace8e6b508a7cf466bbdec3b027ec1496915eef0d342",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4dc2b066cefd16876992ace8e6b508a7cf466bbdec3b027ec1496915eef0d342",
        ],
    )
    rpm(
        name = "perl-Encode-4__3.19-493.fc38.x86_64",
        sha256 = "f990ab13ce7e075d95d22330b88a1fa6b27917a96eead9cab114f07a92835717",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f990ab13ce7e075d95d22330b88a1fa6b27917a96eead9cab114f07a92835717",
        ],
    )
    rpm(
        name = "perl-Errno-0__1.36-497.fc38.x86_64",
        sha256 = "92534ea32c6207fbf307ad69b55de2c5e88d87fc0f6dcc4655de0bb6d1c609b6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/92534ea32c6207fbf307ad69b55de2c5e88d87fc0f6dcc4655de0bb6d1c609b6",
        ],
    )
    rpm(
        name = "perl-Exporter-0__5.77-490.fc38.x86_64",
        sha256 = "3a19ed0ed6e090a1f13d1dff453e627bcc7d8d8b27347399cbdbad15349ad205",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3a19ed0ed6e090a1f13d1dff453e627bcc7d8d8b27347399cbdbad15349ad205",
        ],
    )
    rpm(
        name = "perl-Fcntl-0__1.15-497.fc38.x86_64",
        sha256 = "5ded875b104d81290484ac620b987c1268dee224993960bcc250d8ef6d3d14f2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5ded875b104d81290484ac620b987c1268dee224993960bcc250d8ef6d3d14f2",
        ],
    )
    rpm(
        name = "perl-File-Basename-0__2.85-497.fc38.x86_64",
        sha256 = "f20b061903d0f46ca47b1829399322edf39ea8d186941f8bc892295562fa30bc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f20b061903d0f46ca47b1829399322edf39ea8d186941f8bc892295562fa30bc",
        ],
    )
    rpm(
        name = "perl-File-Path-0__2.18-490.fc38.x86_64",
        sha256 = "49c83e3f49f53f2f26a78871d052d816e0fca30267b0bfd44165003f43f6c3af",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/49c83e3f49f53f2f26a78871d052d816e0fca30267b0bfd44165003f43f6c3af",
        ],
    )
    rpm(
        name = "perl-File-Temp-1__0.231.100-490.fc38.x86_64",
        sha256 = "f6621e3c22248069eecc159560f01dcb9406242928821cf5fd3e4162197c0bfb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f6621e3c22248069eecc159560f01dcb9406242928821cf5fd3e4162197c0bfb",
        ],
    )
    rpm(
        name = "perl-File-stat-0__1.12-497.fc38.x86_64",
        sha256 = "6877d82e0804e6d825a920fda3a6f73ce8887a32690cabdc5b47f6fef7a961b8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6877d82e0804e6d825a920fda3a6f73ce8887a32690cabdc5b47f6fef7a961b8",
        ],
    )
    rpm(
        name = "perl-Getopt-Long-1__2.54-2.fc38.x86_64",
        sha256 = "7df1c477a100ea57f9c07b00ab81ada3f96b0178c817a160a4a917b83f05473b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7df1c477a100ea57f9c07b00ab81ada3f96b0178c817a160a4a917b83f05473b",
        ],
    )
    rpm(
        name = "perl-Getopt-Std-0__1.13-497.fc38.x86_64",
        sha256 = "427808aec37d13fe1c2a6a9c81b4c2f9f008e7f943be779eac57b145f14787e4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/427808aec37d13fe1c2a6a9c81b4c2f9f008e7f943be779eac57b145f14787e4",
        ],
    )
    rpm(
        name = "perl-HTTP-Tiny-0__0.082-2.fc38.x86_64",
        sha256 = "8762414aa11aff68aef75da8215d963b2bbfa578db6e69935615704df6e429e8",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8762414aa11aff68aef75da8215d963b2bbfa578db6e69935615704df6e429e8",
        ],
    )
    rpm(
        name = "perl-IO-0__1.50-497.fc38.x86_64",
        sha256 = "3747b662390db62ae573b3e274133c70242bf6de19cedf289e25e8625e4960ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3747b662390db62ae573b3e274133c70242bf6de19cedf289e25e8625e4960ef",
        ],
    )
    rpm(
        name = "perl-IPC-Open3-0__1.22-497.fc38.x86_64",
        sha256 = "c21cfbe60d4fc86086da5e61b9759a74b36a1e51d73e2d55fa162d61640b19a6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c21cfbe60d4fc86086da5e61b9759a74b36a1e51d73e2d55fa162d61640b19a6",
        ],
    )
    rpm(
        name = "perl-MIME-Base64-0__3.16-490.fc38.x86_64",
        sha256 = "6e2c3f6337ee8155465f26575af59375995ed4d087ba706658a7dc880b7072c6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6e2c3f6337ee8155465f26575af59375995ed4d087ba706658a7dc880b7072c6",
        ],
    )
    rpm(
        name = "perl-POSIX-0__2.03-497.fc38.x86_64",
        sha256 = "8cb37d0fcb51e8be4ced17e96acc7d207d4792bbe506e8b1e4f6c5428c927e07",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8cb37d0fcb51e8be4ced17e96acc7d207d4792bbe506e8b1e4f6c5428c927e07",
        ],
    )
    rpm(
        name = "perl-PathTools-0__3.84-490.fc38.x86_64",
        sha256 = "4f233d672f25351adc4d7d449161e3f2395756abef07448c756ec80ce05133cd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4f233d672f25351adc4d7d449161e3f2395756abef07448c756ec80ce05133cd",
        ],
    )
    rpm(
        name = "perl-Pod-Escapes-1__1.07-490.fc38.x86_64",
        sha256 = "97efca7b54fb5fc965a81d2d2a3e9c9d9c46426102c0a1a50e62d7908642ff48",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/97efca7b54fb5fc965a81d2d2a3e9c9d9c46426102c0a1a50e62d7908642ff48",
        ],
    )
    rpm(
        name = "perl-Pod-Perldoc-0__3.28.01-491.fc38.x86_64",
        sha256 = "9610fe21209fa2f77606a03ef0e989c11f22b1261523bc64d345937422b91b99",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9610fe21209fa2f77606a03ef0e989c11f22b1261523bc64d345937422b91b99",
        ],
    )
    rpm(
        name = "perl-Pod-Simple-1__3.43-491.fc38.x86_64",
        sha256 = "fe2eca800d354db17a262793f22be6b4dfbeabfa70209a5a1066daa7a9e8615b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe2eca800d354db17a262793f22be6b4dfbeabfa70209a5a1066daa7a9e8615b",
        ],
    )
    rpm(
        name = "perl-Pod-Usage-4__2.03-4.fc38.x86_64",
        sha256 = "b8f18fb72cf4f68b06db501510ce940420f9d3db4bfd09a59e9cdd04d46a5f94",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b8f18fb72cf4f68b06db501510ce940420f9d3db4bfd09a59e9cdd04d46a5f94",
        ],
    )
    rpm(
        name = "perl-Scalar-List-Utils-5__1.63-490.fc38.x86_64",
        sha256 = "e81bfefbb6b440fdd36ba06cfa41075a161457d515fa3d4f96316ee5f1fa6d0d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e81bfefbb6b440fdd36ba06cfa41075a161457d515fa3d4f96316ee5f1fa6d0d",
        ],
    )
    rpm(
        name = "perl-SelectSaver-0__1.02-497.fc38.x86_64",
        sha256 = "af6506efa6bad08559dda26065bba98088340939fbce498e4eec624b2e800e6a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/af6506efa6bad08559dda26065bba98088340939fbce498e4eec624b2e800e6a",
        ],
    )
    rpm(
        name = "perl-Socket-4__2.036-2.fc38.x86_64",
        sha256 = "41dea9026ac860cb22858c2d3b0326b5aa4f19bfc3474ed81b09a74de2284983",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/41dea9026ac860cb22858c2d3b0326b5aa4f19bfc3474ed81b09a74de2284983",
        ],
    )
    rpm(
        name = "perl-Storable-1__3.26-490.fc38.x86_64",
        sha256 = "527d7ef92fd9749493ec8d754f271ddaf6d4858400c381601ec6f08c304429e4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/527d7ef92fd9749493ec8d754f271ddaf6d4858400c381601ec6f08c304429e4",
        ],
    )
    rpm(
        name = "perl-Symbol-0__1.09-497.fc38.x86_64",
        sha256 = "e11f0ea26ca0e9b1d9992cac6b71d8ff55f7fefe1e9fcd4c38ff7b666081f688",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e11f0ea26ca0e9b1d9992cac6b71d8ff55f7fefe1e9fcd4c38ff7b666081f688",
        ],
    )
    rpm(
        name = "perl-Term-ANSIColor-0__5.01-491.fc38.x86_64",
        sha256 = "72ea930c8729e4e38646bf1dda650eb048f91fdbad5925344e61ef98819af0f9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/72ea930c8729e4e38646bf1dda650eb048f91fdbad5925344e61ef98819af0f9",
        ],
    )
    rpm(
        name = "perl-Term-Cap-0__1.18-1.fc38.x86_64",
        sha256 = "8fef4e4e43813548eb26633feb5327c27bf50de882716de493d889141813403c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fef4e4e43813548eb26633feb5327c27bf50de882716de493d889141813403c",
        ],
    )
    rpm(
        name = "perl-Text-ParseWords-0__3.31-490.fc38.x86_64",
        sha256 = "848291c71615532da5367e298859e9cc62d081cf85092df437a8d3dd4cef6274",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/848291c71615532da5367e298859e9cc62d081cf85092df437a8d3dd4cef6274",
        ],
    )
    rpm(
        name = "perl-Text-Tabs__plus__Wrap-0__2023.0511-1.fc38.x86_64",
        sha256 = "184e3e2f1cbb7eedf0bac44b9fa2a56079d6d6b3521df09d8526be8f8e583bcc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/184e3e2f1cbb7eedf0bac44b9fa2a56079d6d6b3521df09d8526be8f8e583bcc",
        ],
    )
    rpm(
        name = "perl-Time-Local-2__1.300-490.fc38.x86_64",
        sha256 = "17212325d5ba561cca3c90da43f630e5002fc467bc8bbaca2e853a1163739a2a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/17212325d5ba561cca3c90da43f630e5002fc467bc8bbaca2e853a1163739a2a",
        ],
    )
    rpm(
        name = "perl-constant-0__1.33-491.fc38.x86_64",
        sha256 = "1c7f29465411d6b1d34c26a0ad8aded73737f564978dde5cdb96856d8582d657",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1c7f29465411d6b1d34c26a0ad8aded73737f564978dde5cdb96856d8582d657",
        ],
    )
    rpm(
        name = "perl-if-0__0.61.000-497.fc38.x86_64",
        sha256 = "9138b2e11479581128126bf86c6ae0d3dc85a62092bdf8075eaaaa9be86e6da5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9138b2e11479581128126bf86c6ae0d3dc85a62092bdf8075eaaaa9be86e6da5",
        ],
    )
    rpm(
        name = "perl-interpreter-4__5.36.1-497.fc38.x86_64",
        sha256 = "d27327ac236cf45055dea7b1adaedf5d68e0cf14026f069ba4fb91e5f3a4b883",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d27327ac236cf45055dea7b1adaedf5d68e0cf14026f069ba4fb91e5f3a4b883",
        ],
    )
    rpm(
        name = "perl-libs-4__5.36.1-497.fc38.x86_64",
        sha256 = "b1ea3c773747249aa316f16be93d2c9cd379dd88a7f73c8a9e6dbcf7e7a2702e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b1ea3c773747249aa316f16be93d2c9cd379dd88a7f73c8a9e6dbcf7e7a2702e",
        ],
    )
    rpm(
        name = "perl-locale-0__1.10-497.fc38.x86_64",
        sha256 = "a6bbf15e98dd98b7e9a2a354cf9c243d193d1fd695e9f7f724070fc5468cfd8a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a6bbf15e98dd98b7e9a2a354cf9c243d193d1fd695e9f7f724070fc5468cfd8a",
        ],
    )
    rpm(
        name = "perl-mro-0__1.26-497.fc38.x86_64",
        sha256 = "1998cdb4556ca15c96f3a9458578db64d8789cf81c3ac8d180b367326f0c6ca6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1998cdb4556ca15c96f3a9458578db64d8789cf81c3ac8d180b367326f0c6ca6",
        ],
    )
    rpm(
        name = "perl-overload-0__1.35-497.fc38.x86_64",
        sha256 = "c7a9095395351ba7f65a1d731aa4e2c539353545b1a1f125aff3f6f1a66abab2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c7a9095395351ba7f65a1d731aa4e2c539353545b1a1f125aff3f6f1a66abab2",
        ],
    )
    rpm(
        name = "perl-overloading-0__0.02-497.fc38.x86_64",
        sha256 = "8e4e2a540d8ea9c28f589b59f987a0d6e67e8d1c629dac8ae4865be3ce6d1a8d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8e4e2a540d8ea9c28f589b59f987a0d6e67e8d1c629dac8ae4865be3ce6d1a8d",
        ],
    )
    rpm(
        name = "perl-parent-1__0.241-1.fc38.x86_64",
        sha256 = "8bcd6d24e5a167c2591e0623761e7ed3d8432c1bb34b1078de4489ad169ce565",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8bcd6d24e5a167c2591e0623761e7ed3d8432c1bb34b1078de4489ad169ce565",
        ],
    )
    rpm(
        name = "perl-podlators-1__5.01-2.fc38.x86_64",
        sha256 = "cfe66b448080f7cd49e877d53fa2a5ec6a7577a30cd16b18eb8a461b96e90418",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/cfe66b448080f7cd49e877d53fa2a5ec6a7577a30cd16b18eb8a461b96e90418",
        ],
    )
    rpm(
        name = "perl-vars-0__1.05-497.fc38.x86_64",
        sha256 = "3270bde5294e384ef1c5de29bdd40896f3b3e99a6320ae4e1356e4e82261e3c0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3270bde5294e384ef1c5de29bdd40896f3b3e99a6320ae4e1356e4e82261e3c0",
        ],
    )
    rpm(
        name = "pixman-0__0.42.2-1.fc38.x86_64",
        sha256 = "81e9fade703e1f69587b65633da04f13a6be7f900e20a699d0f4fdb6c5085984",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/81e9fade703e1f69587b65633da04f13a6be7f900e20a699d0f4fdb6c5085984",
        ],
    )

    rpm(
        name = "pkgconf-0__1.8.0-6.fc38.x86_64",
        sha256 = "90fff1832a2af9b5575eb169e9ebf3428e8a59571dd2a4a9d40a5c046c7d2586",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/90fff1832a2af9b5575eb169e9ebf3428e8a59571dd2a4a9d40a5c046c7d2586",
        ],
    )

    rpm(
        name = "pkgconf-m4-0__1.8.0-6.fc38.x86_64",
        sha256 = "24eee58ec1e2406f58dadad135aa1e39bbd86664c6b60b102ff2ebd070c5a2be",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/24eee58ec1e2406f58dadad135aa1e39bbd86664c6b60b102ff2ebd070c5a2be",
        ],
    )

    rpm(
        name = "pkgconf-pkg-config-0__1.8.0-6.fc38.x86_64",
        sha256 = "52bdcde05929fc2fee65b76892e6bb7366f32c633a0a81e9bc52c84b9736fe92",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/52bdcde05929fc2fee65b76892e6bb7366f32c633a0a81e9bc52c84b9736fe92",
        ],
    )
    rpm(
        name = "policycoreutils-0__3.5-1.fc38.x86_64",
        sha256 = "440fc5c6e6a37c47f13d1fb53a03f5cb0155592a5bcf9312e2d083d4bed0ad40",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/440fc5c6e6a37c47f13d1fb53a03f5cb0155592a5bcf9312e2d083d4bed0ad40",
        ],
    )
    rpm(
        name = "policycoreutils-python-utils-0__3.5-1.fc38.x86_64",
        sha256 = "c0f2da6d9bed7e589877e948eaeb34d6b4f24c52f7d35524f7454bc5c58709db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c0f2da6d9bed7e589877e948eaeb34d6b4f24c52f7d35524f7454bc5c58709db",
        ],
    )
    rpm(
        name = "polkit-0__122-3.fc38.1.x86_64",
        sha256 = "716096df1b34d768c3e6a5985de8e1ee58b2183ad9f987aa754e592bd2793c70",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/716096df1b34d768c3e6a5985de8e1ee58b2183ad9f987aa754e592bd2793c70",
        ],
    )
    rpm(
        name = "polkit-libs-0__122-3.fc38.1.x86_64",
        sha256 = "56705b6a1526960d534b0d3e4247deb4eef2b5fa64ceb03544281b8e9bdc4597",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/56705b6a1526960d534b0d3e4247deb4eef2b5fa64ceb03544281b8e9bdc4597",
        ],
    )
    rpm(
        name = "polkit-pkla-compat-0__0.1-23.fc38.x86_64",
        sha256 = "7ffa0438229228bf5ba18945936d52c3620c95f4a3ffc5c5f0f8774fececac0a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7ffa0438229228bf5ba18945936d52c3620c95f4a3ffc5c5f0f8774fececac0a",
        ],
    )

    rpm(
        name = "popt-0__1.19-2.fc38.x86_64",
        sha256 = "fb3fabd657b8f8603c6e19858beb0d506cf957bbca2f3feb827b64c94563b31f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fb3fabd657b8f8603c6e19858beb0d506cf957bbca2f3feb827b64c94563b31f",
        ],
    )
    rpm(
        name = "protobuf-c-0__1.4.1-4.fc38.x86_64",
        sha256 = "8b3f681cd05e071d4c7b21eff4684a3ca7674599ee984cccd6a69a685eb8a41c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8b3f681cd05e071d4c7b21eff4684a3ca7674599ee984cccd6a69a685eb8a41c",
        ],
    )
    rpm(
        name = "psmisc-0__23.6-2.fc38.x86_64",
        sha256 = "6983318d6b2dfd4eea29448e9853b74b1d009ab37be7add3ff304ff0483714cb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6983318d6b2dfd4eea29448e9853b74b1d009ab37be7add3ff304ff0483714cb",
        ],
    )
    rpm(
        name = "publicsuffix-list-dafsa-0__20230318-1.fc38.x86_64",
        sha256 = "abddab3b7b1f90a3146eeed9e5178d884112932fa8110d1a9a1cfe4584f53cb9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/abddab3b7b1f90a3146eeed9e5178d884112932fa8110d1a9a1cfe4584f53cb9",
        ],
    )
    rpm(
        name = "pulseaudio-libs-0__16.1-4.fc38.x86_64",
        sha256 = "ae1ccbe0ac228ba672bf14582c14c3ab99d90e41e155560169bb1620861255db",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ae1ccbe0ac228ba672bf14582c14c3ab99d90e41e155560169bb1620861255db",
        ],
    )
    rpm(
        name = "python-pip-wheel-0__22.3.1-2.fc38.x86_64",
        sha256 = "ee100ea7fe8bf26d44df719283554a36398d484eee28682694c9e7a249c2d49c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/ee100ea7fe8bf26d44df719283554a36398d484eee28682694c9e7a249c2d49c",
        ],
    )
    rpm(
        name = "python-setuptools-wheel-0__65.5.1-2.fc38.x86_64",
        sha256 = "7417816bd96d7b49e5a98c85eba313afaa8b8802458d7cd9f5ba72ecc31933e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7417816bd96d7b49e5a98c85eba313afaa8b8802458d7cd9f5ba72ecc31933e3",
        ],
    )
    rpm(
        name = "python3-0__3.11.3-2.fc38.x86_64",
        sha256 = "9f6acd253d7c2dd0df445d321ac9df3a2870781cfac6837ba424a7aa0f24b469",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f6acd253d7c2dd0df445d321ac9df3a2870781cfac6837ba424a7aa0f24b469",
        ],
    )
    rpm(
        name = "python3-audit-0__3.1.1-1.fc38.x86_64",
        sha256 = "e1ad1ab7e284106d4aff8a7026d6e848eed005d6adbd0cc336924a41b0776275",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e1ad1ab7e284106d4aff8a7026d6e848eed005d6adbd0cc336924a41b0776275",
        ],
    )
    rpm(
        name = "python3-distro-0__1.8.0-2.fc38.x86_64",
        sha256 = "9d11e90a67f22a3dcb5dd8593b5ca2320b29173aa4d815710d1886ae8868c17f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9d11e90a67f22a3dcb5dd8593b5ca2320b29173aa4d815710d1886ae8868c17f",
        ],
    )
    rpm(
        name = "python3-libs-0__3.11.3-2.fc38.x86_64",
        sha256 = "2656094a2199cb34ffd66c00051e93a15b742bc618bf02b1f5886ea3560e8a6d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2656094a2199cb34ffd66c00051e93a15b742bc618bf02b1f5886ea3560e8a6d",
        ],
    )
    rpm(
        name = "python3-libselinux-0__3.5-1.fc38.x86_64",
        sha256 = "2fb45f352d4f8f51c2124e8857ccbd7d5fe3a477174c7a4597f1fba88073bd39",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2fb45f352d4f8f51c2124e8857ccbd7d5fe3a477174c7a4597f1fba88073bd39",
        ],
    )
    rpm(
        name = "python3-libsemanage-0__3.5-2.fc38.x86_64",
        sha256 = "34a08d3ab88b1ef614a5cea8ce5a15d196c3841ad36130b5a66abd3f8fd5ad18",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/34a08d3ab88b1ef614a5cea8ce5a15d196c3841ad36130b5a66abd3f8fd5ad18",
        ],
    )
    rpm(
        name = "python3-policycoreutils-0__3.5-1.fc38.x86_64",
        sha256 = "45121ee061e87f94bcc0728007f99f85bf115145a2e44f5aef04bf570d35aad0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/45121ee061e87f94bcc0728007f99f85bf115145a2e44f5aef04bf570d35aad0",
        ],
    )
    rpm(
        name = "python3-setools-0__4.4.2-1.fc38.x86_64",
        sha256 = "fd08189ecad59c6784a241b8185c90b0c15a0679b0f9e2fa899b26c28033cfca",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fd08189ecad59c6784a241b8185c90b0c15a0679b0f9e2fa899b26c28033cfca",
        ],
    )
    rpm(
        name = "python3-setuptools-0__65.5.1-2.fc38.x86_64",
        sha256 = "fa61f497c2f94b4bbe9022f2e7a1bb4138aa02db1b5f60706a0056cccf2eb7ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fa61f497c2f94b4bbe9022f2e7a1bb4138aa02db1b5f60706a0056cccf2eb7ac",
        ],
    )
    rpm(
        name = "qemu-audio-alsa-2__7.2.1-2.fc38.x86_64",
        sha256 = "a6c57f5caa7840ecc96db35b0632650ef63105ef5b62fab6dc95615f7a3b7ba7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a6c57f5caa7840ecc96db35b0632650ef63105ef5b62fab6dc95615f7a3b7ba7",
        ],
    )
    rpm(
        name = "qemu-audio-dbus-2__7.2.1-2.fc38.x86_64",
        sha256 = "758bacbe5d56fb594004096cb37ebbc42bf7f8773289c57ae92b8845490dbffa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/758bacbe5d56fb594004096cb37ebbc42bf7f8773289c57ae92b8845490dbffa",
        ],
    )
    rpm(
        name = "qemu-audio-jack-2__7.2.1-2.fc38.x86_64",
        sha256 = "c4f9b1ac87660c4c7b758e490860ba9a53db611d5cc59b9f11d3ce090d1e2ba7",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c4f9b1ac87660c4c7b758e490860ba9a53db611d5cc59b9f11d3ce090d1e2ba7",
        ],
    )
    rpm(
        name = "qemu-audio-oss-2__7.2.1-2.fc38.x86_64",
        sha256 = "910063787967b8d48d8210722748c8002fe1dae379c2cdcd04c838ab6f83a2ac",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/910063787967b8d48d8210722748c8002fe1dae379c2cdcd04c838ab6f83a2ac",
        ],
    )
    rpm(
        name = "qemu-audio-pa-2__7.2.1-2.fc38.x86_64",
        sha256 = "091f80bfdbf4a795cafa67b159e077a47967efcc7379e7d6e62f9facf1823ecf",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/091f80bfdbf4a795cafa67b159e077a47967efcc7379e7d6e62f9facf1823ecf",
        ],
    )
    rpm(
        name = "qemu-audio-sdl-2__7.2.1-2.fc38.x86_64",
        sha256 = "44cb62d1747d673d090a527799bb20818c7a65eac4c030057d465c50bd23f0e5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/44cb62d1747d673d090a527799bb20818c7a65eac4c030057d465c50bd23f0e5",
        ],
    )
    rpm(
        name = "qemu-audio-spice-2__7.2.1-2.fc38.x86_64",
        sha256 = "b68effb6fae480ea2eddc6d18c7eed78620dbbd4595a30e5f05fd589128a866f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b68effb6fae480ea2eddc6d18c7eed78620dbbd4595a30e5f05fd589128a866f",
        ],
    )
    rpm(
        name = "qemu-block-blkio-2__7.2.1-2.fc38.x86_64",
        sha256 = "22739bbf4ec7484146da785b03976b00ee3b2252c895a6b0fc6d9f69063388c5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/22739bbf4ec7484146da785b03976b00ee3b2252c895a6b0fc6d9f69063388c5",
        ],
    )
    rpm(
        name = "qemu-block-curl-2__7.2.1-2.fc38.x86_64",
        sha256 = "5c71c64e3154344573bad1262110648a83d6bcddc79ec68b5b97e6fee8eb0543",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5c71c64e3154344573bad1262110648a83d6bcddc79ec68b5b97e6fee8eb0543",
        ],
    )
    rpm(
        name = "qemu-block-dmg-2__7.2.1-2.fc38.x86_64",
        sha256 = "f6a2afb14da0d20210384b40e37027966d34beb0bfd91fd8e88e6ef05c53cc23",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f6a2afb14da0d20210384b40e37027966d34beb0bfd91fd8e88e6ef05c53cc23",
        ],
    )
    rpm(
        name = "qemu-block-gluster-2__7.2.1-2.fc38.x86_64",
        sha256 = "29292e0035222e63a38ea5fb680e12a1a4696c89e2f8abd8232816e76a6dbab0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/29292e0035222e63a38ea5fb680e12a1a4696c89e2f8abd8232816e76a6dbab0",
        ],
    )
    rpm(
        name = "qemu-block-iscsi-2__7.2.1-2.fc38.x86_64",
        sha256 = "1063a251486f902071a37737d971342625a19fa4bb5fe2df90386879570fb8c2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1063a251486f902071a37737d971342625a19fa4bb5fe2df90386879570fb8c2",
        ],
    )
    rpm(
        name = "qemu-block-nfs-2__7.2.1-2.fc38.x86_64",
        sha256 = "4d482ccbf16ef524afaed3fd7a587acfa709021a6970c04a4c6ba4cc63f972fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d482ccbf16ef524afaed3fd7a587acfa709021a6970c04a4c6ba4cc63f972fd",
        ],
    )
    rpm(
        name = "qemu-block-rbd-2__7.2.1-2.fc38.x86_64",
        sha256 = "c67b4b81d19dbba97f2c634ec3ef76230bd65f385733c1f006676076b31bf714",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c67b4b81d19dbba97f2c634ec3ef76230bd65f385733c1f006676076b31bf714",
        ],
    )
    rpm(
        name = "qemu-block-ssh-2__7.2.1-2.fc38.x86_64",
        sha256 = "77680d074fc923571dd14c668cad3839bf3ebf473e92791450432daa84ee3614",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/77680d074fc923571dd14c668cad3839bf3ebf473e92791450432daa84ee3614",
        ],
    )
    rpm(
        name = "qemu-char-baum-2__7.2.1-2.fc38.x86_64",
        sha256 = "f4daae201b3f3ddb93ae1c8a6191477eb68a97e19005b70ac1ad7c9ab3c2d4f1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f4daae201b3f3ddb93ae1c8a6191477eb68a97e19005b70ac1ad7c9ab3c2d4f1",
        ],
    )
    rpm(
        name = "qemu-char-spice-2__7.2.1-2.fc38.x86_64",
        sha256 = "4d367ef2e559073fd060d2c14ccc708aecc27f5fd50e06977e6de009c78fc5ce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d367ef2e559073fd060d2c14ccc708aecc27f5fd50e06977e6de009c78fc5ce",
        ],
    )
    rpm(
        name = "qemu-common-2__7.2.1-2.fc38.x86_64",
        sha256 = "e02b9a9dee0b8d77082312e165aa441d67643b3c1812540170bab535f8c55f4a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e02b9a9dee0b8d77082312e165aa441d67643b3c1812540170bab535f8c55f4a",
        ],
    )
    rpm(
        name = "qemu-device-display-qxl-2__7.2.1-2.fc38.x86_64",
        sha256 = "3b5b52895a92bfbfd298cf432b28fcc51e3861f69dbe90f38c14c2445a9f3206",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3b5b52895a92bfbfd298cf432b28fcc51e3861f69dbe90f38c14c2445a9f3206",
        ],
    )
    rpm(
        name = "qemu-device-display-vhost-user-gpu-2__7.2.1-2.fc38.x86_64",
        sha256 = "6cd4661c86f2cf4434d2439d2da85616fdf53fd07c96b2ceadc8a980ed84493e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6cd4661c86f2cf4434d2439d2da85616fdf53fd07c96b2ceadc8a980ed84493e",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-gpu-2__7.2.1-2.fc38.x86_64",
        sha256 = "e5f7b0e6ec9b53487bfbda1ece32cd3953c915384e5779f7bae13ba7dd08727a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e5f7b0e6ec9b53487bfbda1ece32cd3953c915384e5779f7bae13ba7dd08727a",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-gpu-ccw-2__7.2.1-2.fc38.x86_64",
        sha256 = "0b3beb09285a88de065c550599c3d049f6669222ccc81cb2caa4e5ea5419b384",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0b3beb09285a88de065c550599c3d049f6669222ccc81cb2caa4e5ea5419b384",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-gpu-gl-2__7.2.1-2.fc38.x86_64",
        sha256 = "d1117bde1f020a1fb69ba88bf6cc08a51900c4c590636b01db65e7f42acc01aa",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d1117bde1f020a1fb69ba88bf6cc08a51900c4c590636b01db65e7f42acc01aa",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-gpu-pci-2__7.2.1-2.fc38.x86_64",
        sha256 = "1f752732d0e2a13652c7195853dbb700aebcb8ca320316d1c8821ff4f685c972",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/1f752732d0e2a13652c7195853dbb700aebcb8ca320316d1c8821ff4f685c972",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-gpu-pci-gl-2__7.2.1-2.fc38.x86_64",
        sha256 = "496e3ad720be68f0c52dee791cd6ebd282379fbaa3e9037d52cfac2a8636060f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/496e3ad720be68f0c52dee791cd6ebd282379fbaa3e9037d52cfac2a8636060f",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-vga-2__7.2.1-2.fc38.x86_64",
        sha256 = "96c50a0180a4758ab39e17a477213cc5bccbcd307cb3249d60bed2d566060f59",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/96c50a0180a4758ab39e17a477213cc5bccbcd307cb3249d60bed2d566060f59",
        ],
    )
    rpm(
        name = "qemu-device-display-virtio-vga-gl-2__7.2.1-2.fc38.x86_64",
        sha256 = "4b654c0a02196a63bc783cbd3862774272e31bc6fd1d27c31e570544787d8cf9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4b654c0a02196a63bc783cbd3862774272e31bc6fd1d27c31e570544787d8cf9",
        ],
    )
    rpm(
        name = "qemu-device-usb-host-2__7.2.1-2.fc38.x86_64",
        sha256 = "807d83cdb74d3b4cd9a07474e763e54ac90aa535fb8f1b1573e74eb0b4ee3bdc",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/807d83cdb74d3b4cd9a07474e763e54ac90aa535fb8f1b1573e74eb0b4ee3bdc",
        ],
    )
    rpm(
        name = "qemu-device-usb-redirect-2__7.2.1-2.fc38.x86_64",
        sha256 = "d42dd802cb3e028386205680d549c727eecabb0c121ef09730ade7a037368555",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d42dd802cb3e028386205680d549c727eecabb0c121ef09730ade7a037368555",
        ],
    )
    rpm(
        name = "qemu-device-usb-smartcard-2__7.2.1-2.fc38.x86_64",
        sha256 = "0f29a663241c428c55c64caef3f03ed159f153d58058c25de468bc53af58d631",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/0f29a663241c428c55c64caef3f03ed159f153d58058c25de468bc53af58d631",
        ],
    )
    rpm(
        name = "qemu-img-2__7.2.1-2.fc38.x86_64",
        sha256 = "840bf28d52bc9b1b0f90e415f507fdb9e875dc52c3153425ecc112a0735fc757",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/840bf28d52bc9b1b0f90e415f507fdb9e875dc52c3153425ecc112a0735fc757",
        ],
    )
    rpm(
        name = "qemu-kvm-2__7.2.1-2.fc38.x86_64",
        sha256 = "beca8df43390186b92594e2e16d49d9c8c09836130d594100b8783f0392601c0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/beca8df43390186b92594e2e16d49d9c8c09836130d594100b8783f0392601c0",
        ],
    )
    rpm(
        name = "qemu-pr-helper-2__7.2.1-2.fc38.x86_64",
        sha256 = "8fc2764e72453d0765f28ee2459bba8d936f7d1af84ca90dcde65922dbaed279",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8fc2764e72453d0765f28ee2459bba8d936f7d1af84ca90dcde65922dbaed279",
        ],
    )
    rpm(
        name = "qemu-system-x86-2__7.2.1-2.fc38.x86_64",
        sha256 = "2b4e4f261d92aaef854b3c96e485c61fc9a4663223a37ec86e2da4e48df6a190",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2b4e4f261d92aaef854b3c96e485c61fc9a4663223a37ec86e2da4e48df6a190",
        ],
    )
    rpm(
        name = "qemu-system-x86-core-2__7.2.1-2.fc38.x86_64",
        sha256 = "8d1f4853ba9c5877e52a6e6bf3093eecbe7c4c272ea4fcc210f0a542cdfc7338",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8d1f4853ba9c5877e52a6e6bf3093eecbe7c4c272ea4fcc210f0a542cdfc7338",
        ],
    )
    rpm(
        name = "qemu-ui-curses-2__7.2.1-2.fc38.x86_64",
        sha256 = "9f254c5529e8e5bf6e1c205491723124480b3f7b800a92469a1049f9210368b5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f254c5529e8e5bf6e1c205491723124480b3f7b800a92469a1049f9210368b5",
        ],
    )
    rpm(
        name = "qemu-ui-egl-headless-2__7.2.1-2.fc38.x86_64",
        sha256 = "2bc76ec5ca1d1b9b4249642e9a7998dbe623d43906a576d0f49a56c7ad046eb3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2bc76ec5ca1d1b9b4249642e9a7998dbe623d43906a576d0f49a56c7ad046eb3",
        ],
    )
    rpm(
        name = "qemu-ui-gtk-2__7.2.1-2.fc38.x86_64",
        sha256 = "81b846b9dc3ee9cb3913cd27e84089d419cea4da0c8d20518335dcd62015fe32",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/81b846b9dc3ee9cb3913cd27e84089d419cea4da0c8d20518335dcd62015fe32",
        ],
    )
    rpm(
        name = "qemu-ui-opengl-2__7.2.1-2.fc38.x86_64",
        sha256 = "b72892784d94d278b43f4fb876079d79a2d02bb8b7808600edc29d0dfca60700",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b72892784d94d278b43f4fb876079d79a2d02bb8b7808600edc29d0dfca60700",
        ],
    )
    rpm(
        name = "qemu-ui-sdl-2__7.2.1-2.fc38.x86_64",
        sha256 = "26f619023e0bc7396c8e1911045a579d30740d30f2bbc51943f237f5f1749076",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/26f619023e0bc7396c8e1911045a579d30740d30f2bbc51943f237f5f1749076",
        ],
    )
    rpm(
        name = "qemu-ui-spice-app-2__7.2.1-2.fc38.x86_64",
        sha256 = "e792febb65c47bfd079c368f784e553543847cc5dfcbb047a50555f63e62c485",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e792febb65c47bfd079c368f784e553543847cc5dfcbb047a50555f63e62c485",
        ],
    )
    rpm(
        name = "qemu-ui-spice-core-2__7.2.1-2.fc38.x86_64",
        sha256 = "64c0d63cd5b9c0b9ea4083fa8472501170b6e14be2330e844736bf951fc6652d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/64c0d63cd5b9c0b9ea4083fa8472501170b6e14be2330e844736bf951fc6652d",
        ],
    )
    rpm(
        name = "quota-1__4.09-2.fc38.x86_64",
        sha256 = "e69a038bc442cbcdeaec89b06af30b7f9eb57017657c1d57b3128b54be8bb153",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e69a038bc442cbcdeaec89b06af30b7f9eb57017657c1d57b3128b54be8bb153",
        ],
    )
    rpm(
        name = "quota-nls-1__4.09-2.fc38.x86_64",
        sha256 = "48cd3d08dc037320afa5f98501088f06198d3d4deff554d38f7f04b440f0ae72",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/48cd3d08dc037320afa5f98501088f06198d3d4deff554d38f7f04b440f0ae72",
        ],
    )

    rpm(
        name = "readline-0__8.2-3.fc38.x86_64",
        sha256 = "a7099c322c45030738bdd90e3de4402c0c80c6ebd993a1749c2e582cf33ee6f2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a7099c322c45030738bdd90e3de4402c0c80c6ebd993a1749c2e582cf33ee6f2",
        ],
    )
    rpm(
        name = "rpcbind-0__1.2.6-4.rc2.fc38.x86_64",
        sha256 = "86b2fce241deb1b66fcd2f15b86824f7202c3a05d53973b68f0170401bccf94c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/86b2fce241deb1b66fcd2f15b86824f7202c3a05d53973b68f0170401bccf94c",
        ],
    )
    rpm(
        name = "rpm-0__4.18.1-3.fc38.x86_64",
        sha256 = "fcd827a4f95db14a707bc604efac1db86973eaa97bd8bf9f78c37a925c09a297",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fcd827a4f95db14a707bc604efac1db86973eaa97bd8bf9f78c37a925c09a297",
        ],
    )
    rpm(
        name = "rpm-libs-0__4.18.1-3.fc38.x86_64",
        sha256 = "10196a125eccfe91f3eb187be4e07d7a7d9304f738c3feed087b6f8c4375f305",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/10196a125eccfe91f3eb187be4e07d7a7d9304f738c3feed087b6f8c4375f305",
        ],
    )
    rpm(
        name = "rpm-plugin-selinux-0__4.18.1-3.fc38.x86_64",
        sha256 = "edc3846d881c4ad341c5e97f7a2b7ec8f5164c26e73839449374b3c6fc2483fd",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/edc3846d881c4ad341c5e97f7a2b7ec8f5164c26e73839449374b3c6fc2483fd",
        ],
    )
    rpm(
        name = "rpm-sequoia-0__1.4.0-3.fc38.x86_64",
        sha256 = "5ae40c78ba1b3d44bda453048fa602bb2cbfef08f5873e3cc268490d5a68d24b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5ae40c78ba1b3d44bda453048fa602bb2cbfef08f5873e3cc268490d5a68d24b",
        ],
    )
    rpm(
        name = "seabios-bin-0__1.16.2-1.fc38.x86_64",
        sha256 = "7e75f0d13dced300b02c7dd7202471e78a017abd4513413a9679021278e2c589",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e75f0d13dced300b02c7dd7202471e78a017abd4513413a9679021278e2c589",
        ],
    )
    rpm(
        name = "seavgabios-bin-0__1.16.2-1.fc38.x86_64",
        sha256 = "8022fcff2140e4ab964b50bb63a6a2afdb501b6014519d79d15f0b64fe1c578b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8022fcff2140e4ab964b50bb63a6a2afdb501b6014519d79d15f0b64fe1c578b",
        ],
    )

    rpm(
        name = "sed-0__4.8-12.fc38.x86_64",
        sha256 = "a6e01b89e814ec42d1c2c6be79240a97a9bd151c857f82a11e129547e069e27f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a6e01b89e814ec42d1c2c6be79240a97a9bd151c857f82a11e129547e069e27f",
        ],
    )
    rpm(
        name = "selinux-policy-0__38.15-1.fc38.x86_64",
        sha256 = "a95a13f416c14a654a3571f229f9b39dfa8c32dd330de620ae08ad359eb317c0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a95a13f416c14a654a3571f229f9b39dfa8c32dd330de620ae08ad359eb317c0",
        ],
    )
    rpm(
        name = "selinux-policy-minimum-0__38.15-1.fc38.x86_64",
        sha256 = "95e811c04645ee5249c2c3050962e04255ed2de37f261e15539c415650a1e78d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/95e811c04645ee5249c2c3050962e04255ed2de37f261e15539c415650a1e78d",
        ],
    )

    rpm(
        name = "setup-0__2.14.3-2.fc38.x86_64",
        sha256 = "c7efb8634b62cdab9e8894174a8c70d20eb431482277231bc54fa8ca8c3681e3",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c7efb8634b62cdab9e8894174a8c70d20eb431482277231bc54fa8ca8c3681e3",
        ],
    )
    rpm(
        name = "sgabios-bin-1__0.20180715git-10.fc38.x86_64",
        sha256 = "29d65ece6fb972f7308ab795e45d039e976722c5687be31e2a9d8a0a7ca832b9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/29d65ece6fb972f7308ab795e45d039e976722c5687be31e2a9d8a0a7ca832b9",
        ],
    )

    rpm(
        name = "shadow-utils-2__4.13-6.fc38.x86_64",
        sha256 = "8be96e09e2e44491b287be44b2b6be0c9f8aeab75fe061d3e8a28b9be19144ef",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/8be96e09e2e44491b287be44b2b6be0c9f8aeab75fe061d3e8a28b9be19144ef",
        ],
    )
    rpm(
        name = "shared-mime-info-0__2.2-3.fc38.x86_64",
        sha256 = "5599d1ba2f455133813c24bb4c4bb9bc0842ccb28371c1c15f0a0dcb6262c004",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5599d1ba2f455133813c24bb4c4bb9bc0842ccb28371c1c15f0a0dcb6262c004",
        ],
    )
    rpm(
        name = "snappy-0__1.1.9-7.fc38.x86_64",
        sha256 = "a1de300e5edca8a97e90e0196a7f38a900e45f410e0ac04cb1366fead20a1659",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a1de300e5edca8a97e90e0196a7f38a900e45f410e0ac04cb1366fead20a1659",
        ],
    )
    rpm(
        name = "spice-server-0__0.15.1-2.fc38.x86_64",
        sha256 = "5c796706d5a73267c3ca82a9331b080a25e7864d705a1c810b6329d77fc7febe",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/5c796706d5a73267c3ca82a9331b080a25e7864d705a1c810b6329d77fc7febe",
        ],
    )
    rpm(
        name = "sqlite-libs-0__3.40.1-2.fc38.x86_64",
        sha256 = "be6d9e9b98733494ee5901d45b849e2dc012a6b41c3ff09d0d212002dbe15dce",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/be6d9e9b98733494ee5901d45b849e2dc012a6b41c3ff09d0d212002dbe15dce",
        ],
    )
    rpm(
        name = "swtpm-0__0.8.0-3.fc38.x86_64",
        sha256 = "6738d2b352e3e08a3482f46640a27135a5d83ba616fcba4902d2569230eb9eb9",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6738d2b352e3e08a3482f46640a27135a5d83ba616fcba4902d2569230eb9eb9",
        ],
    )
    rpm(
        name = "swtpm-libs-0__0.8.0-3.fc38.x86_64",
        sha256 = "e0a615080f43adbc1cfd0247505f2da4725a12ac2bf98fad461d75bc2f21e3c6",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e0a615080f43adbc1cfd0247505f2da4725a12ac2bf98fad461d75bc2f21e3c6",
        ],
    )
    rpm(
        name = "swtpm-tools-0__0.8.0-3.fc38.x86_64",
        sha256 = "7e4b4392d27d104c548c1925db99f6252839c9893ea0896715c3b9f3a498df6f",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/7e4b4392d27d104c548c1925db99f6252839c9893ea0896715c3b9f3a498df6f",
        ],
    )

    rpm(
        name = "systemd-0__253.5-1.fc38.x86_64",
        sha256 = "23c41203de9a6741e263b0d9f5a13bbe88387ba0b7c8325ae3d01d17dd675924",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/23c41203de9a6741e263b0d9f5a13bbe88387ba0b7c8325ae3d01d17dd675924",
        ],
    )
    rpm(
        name = "systemd-container-0__253.5-1.fc38.x86_64",
        sha256 = "9f22f36aa64b2cdc677f1a5997af6aaa5b8f13828d5f1bfc526e011565caf110",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/9f22f36aa64b2cdc677f1a5997af6aaa5b8f13828d5f1bfc526e011565caf110",
        ],
    )

    rpm(
        name = "systemd-devel-0__253.5-1.fc38.x86_64",
        sha256 = "bbd9fd3b74dcad2b1a7b3dd2478dbea49569fc365b8453558a171a15bb987bb2",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bbd9fd3b74dcad2b1a7b3dd2478dbea49569fc365b8453558a171a15bb987bb2",
        ],
    )

    rpm(
        name = "systemd-libs-0__253.5-1.fc38.x86_64",
        sha256 = "3131390399cb6df4e93409013c0af49975b474982f37da7f3e06aae7aa9a2b96",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/3131390399cb6df4e93409013c0af49975b474982f37da7f3e06aae7aa9a2b96",
        ],
    )

    rpm(
        name = "systemd-pam-0__253.5-1.fc38.x86_64",
        sha256 = "a9dc53ec81e7f8ba75ae57aa1d6c7e55781357c700dd73e854fc05936ae65121",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a9dc53ec81e7f8ba75ae57aa1d6c7e55781357c700dd73e854fc05936ae65121",
        ],
    )
    rpm(
        name = "systemd-udev-0__253.5-1.fc38.x86_64",
        sha256 = "f773968807ad670d6c07f1dcb7a35ce131a8365bcedc58593e96bbe53d2df82e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f773968807ad670d6c07f1dcb7a35ce131a8365bcedc58593e96bbe53d2df82e",
        ],
    )
    rpm(
        name = "trousers-0__0.3.15-8.fc38.x86_64",
        sha256 = "6194d3b90915b71d1ffbc9f8e9ae39b4605f7d7382a3a32a016e500f383ce06d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/6194d3b90915b71d1ffbc9f8e9ae39b4605f7d7382a3a32a016e500f383ce06d",
        ],
    )
    rpm(
        name = "trousers-lib-0__0.3.15-8.fc38.x86_64",
        sha256 = "dba986bc506f6bcd9844a2f2c151c7ec47fbd210b580a90312dce39f5c7839d5",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/dba986bc506f6bcd9844a2f2c151c7ec47fbd210b580a90312dce39f5c7839d5",
        ],
    )

    rpm(
        name = "tzdata-0__2023c-1.fc38.x86_64",
        sha256 = "041e8b9be8a87757da8d5b8a2ced4b5aec8bcafd1c0747234cdfe10206eae27b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/041e8b9be8a87757da8d5b8a2ced4b5aec8bcafd1c0747234cdfe10206eae27b",
        ],
    )
    rpm(
        name = "unbound-libs-0__1.17.1-2.fc38.x86_64",
        sha256 = "95ea79593c6ef3cc7e561f222b3c13db8b1c17facd1786bb366a0663624a7580",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/95ea79593c6ef3cc7e561f222b3c13db8b1c17facd1786bb366a0663624a7580",
        ],
    )
    rpm(
        name = "usbredir-0__0.13.0-2.fc38.x86_64",
        sha256 = "a8b03b803e32d5bd74c734a5b1afd989709f2a915c512d9e560d08e3148f5c0d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a8b03b803e32d5bd74c734a5b1afd989709f2a915c512d9e560d08e3148f5c0d",
        ],
    )
    rpm(
        name = "userspace-rcu-0__0.13.2-2.fc38.x86_64",
        sha256 = "76b58342bb839cf953fb2685bb91a929f397da361535766de6a597e9279abfa1",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/76b58342bb839cf953fb2685bb91a929f397da361535766de6a597e9279abfa1",
        ],
    )

    rpm(
        name = "util-linux-0__2.38.1-4.fc38.x86_64",
        sha256 = "f0f8e33332df97afd911093f28c487bc84cbe4dcc7bb468eac5551d235acee62",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f0f8e33332df97afd911093f28c487bc84cbe4dcc7bb468eac5551d235acee62",
        ],
    )

    rpm(
        name = "util-linux-core-0__2.38.1-4.fc38.x86_64",
        sha256 = "b57dbbbee14301e89df618b398ef39b7fc841eaba6be1b6346cf37ed7695c26a",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/b57dbbbee14301e89df618b398ef39b7fc841eaba6be1b6346cf37ed7695c26a",
        ],
    )
    rpm(
        name = "virglrenderer-0__0.10.4-2.20230104git88b9fe3b.fc38.x86_64",
        sha256 = "339bbe509c15eeab65f680cdfb97d64571e8ac8e57ece6e7cbab96e075a42460",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/339bbe509c15eeab65f680cdfb97d64571e8ac8e57ece6e7cbab96e075a42460",
        ],
    )
    rpm(
        name = "virtiofsd-0__1.5.1-1.fc38.x86_64",
        sha256 = "a11f5d602157820ad8f234f7c39c5ce90ca97bb57faae2ac2d0aa179604e8879",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a11f5d602157820ad8f234f7c39c5ce90ca97bb57faae2ac2d0aa179604e8879",
        ],
    )
    rpm(
        name = "vte-profile-0__0.72.2-1.fc38.x86_64",
        sha256 = "4d762f9df170baa7baa20bcbcd3ee027e9380595c673dcb61d0d438472adc22e",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4d762f9df170baa7baa20bcbcd3ee027e9380595c673dcb61d0d438472adc22e",
        ],
    )
    rpm(
        name = "vte291-0__0.72.2-1.fc38.x86_64",
        sha256 = "022273c55492676fd99084ebd78d24c59b40393c4bd76be5a91500875c8aa36b",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/022273c55492676fd99084ebd78d24c59b40393c4bd76be5a91500875c8aa36b",
        ],
    )
    rpm(
        name = "which-0__2.21-39.fc38.x86_64",
        sha256 = "2c8b143f3cb83efa5a31c85bea1da3164ca2dde5e2d75d25115f3e21ef98b4e0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/2c8b143f3cb83efa5a31c85bea1da3164ca2dde5e2d75d25115f3e21ef98b4e0",
        ],
    )
    rpm(
        name = "xen-libs-0__4.17.1-2.fc38.x86_64",
        sha256 = "f551000b5e27e0695ac24e3ce8e289da7500a7c0098c499f441d50143a5dc787",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/f551000b5e27e0695ac24e3ce8e289da7500a7c0098c499f441d50143a5dc787",
        ],
    )
    rpm(
        name = "xen-licenses-0__4.17.1-2.fc38.x86_64",
        sha256 = "4cac4a65e24f4477ce8d6799c9ee732ff498dae572eef89707800acadc15f3ab",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/4cac4a65e24f4477ce8d6799c9ee732ff498dae572eef89707800acadc15f3ab",
        ],
    )
    rpm(
        name = "xkeyboard-config-0__2.38-1.fc38.x86_64",
        sha256 = "59a7a5a775c196961cdc51fb89440a055295c767a632bfa684760e73650aa9a0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/59a7a5a775c196961cdc51fb89440a055295c767a632bfa684760e73650aa9a0",
        ],
    )
    rpm(
        name = "xml-common-0__0.6.3-60.fc38.x86_64",
        sha256 = "a92e1f4689cf3774070bf217c0a0c8628fd877a9c0f98fdadbb597eeb44fa2cb",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/a92e1f4689cf3774070bf217c0a0c8628fd877a9c0f98fdadbb597eeb44fa2cb",
        ],
    )
    rpm(
        name = "xprop-0__1.2.5-3.fc38.x86_64",
        sha256 = "d7e47678c82e0f244e612f3eae69afa3279f20ae4f6762dcbe263e19fb72496d",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/d7e47678c82e0f244e612f3eae69afa3279f20ae4f6762dcbe263e19fb72496d",
        ],
    )
    rpm(
        name = "xz-0__5.4.1-1.fc38.x86_64",
        sha256 = "e911703ffceee37ec1066344820ab0cf9ba8e43d7957395981ba68c4d411a0a4",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/e911703ffceee37ec1066344820ab0cf9ba8e43d7957395981ba68c4d411a0a4",
        ],
    )

    rpm(
        name = "xz-libs-0__5.4.1-1.fc38.x86_64",
        sha256 = "bfce8ac2a2a78a23fb931531fb3d8f530a78f4d5b17f6199bf99b93ca21858c0",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/bfce8ac2a2a78a23fb931531fb3d8f530a78f4d5b17f6199bf99b93ca21858c0",
        ],
    )

    rpm(
        name = "yajl-0__2.1.0-20.fc38.x86_64",
        sha256 = "c4d4acdd6d44c99e4fea352b764fa20de60c0db7d51cdbdce852cd7ebffd5a22",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c4d4acdd6d44c99e4fea352b764fa20de60c0db7d51cdbdce852cd7ebffd5a22",
        ],
    )
    rpm(
        name = "zfs-fuse-0__0.7.2.2-24.fc38.x86_64",
        sha256 = "fe8d5c4e2e7ebb4ba9e82a9c88767b3e3c6f376ec8b1b8131ef9c84507d2e629",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/fe8d5c4e2e7ebb4ba9e82a9c88767b3e3c6f376ec8b1b8131ef9c84507d2e629",
        ],
    )

    rpm(
        name = "zlib-0__1.2.13-3.fc38.x86_64",
        sha256 = "c26d4d161f8eddd7cb794075e383d0f4d3a77aa88e453a2db51e53346981f04c",
        urls = [
            "https://cdn.confidential.cloud/constellation/cas/sha256/c26d4d161f8eddd7cb794075e383d0f4d3a77aa88e453a2db51e53346981f04c",
        ],
    )
