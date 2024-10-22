# Reproduce released artifacts

Constellation has first-class support for [reproducible builds](https://reproducible-builds.org).
Reproducing the released artifacts is an alternative to [signature verification](verify-cli.md) that doesn't require trusting Edgeless Systems' release process.
The following sections describe how to rebuild an artifact and how Constellation ensures that the rebuild reproduces the artifacts bit-by-bit.

## Build environment prerequisites

The build systems used by Constellation - [Bazel](https://bazel.build/) and [Nix](https://nixos.org) - are designed for deterministic, reproducible builds.
These two dependencies should be the only prerequisites for a successful build.
However, it can't be ruled out completely that peculiarities of the host affect the build result.
Thus, we recommend the following host setup for best results:

1. A Linux operating system not older than v5.4.
2. The GNU C library not older than v2.31 (avoid `musl`).
3. GNU `coreutils` not older than v8.30 (avoid `busybox`).
4. An `ext4` filesystem for building.
5. AppArmor turned off.

This is given, for example, on an Ubuntu 22.04 system, which is also used for reproducibility tests.

:::note

To avoid any backwards-compatibility issues, the host software versions should also not be much newer than the Constellation release.

:::

## Run the build

The following instructions outline qualitatively how to reproduce a build.
Constellation implements these instructions in the [Reproducible Builds workflow](https://github.com/edgelesssys/constellation/actions/workflows/reproducible-builds.yml), which continuously tests for reproducibility.
The workflow is a good place to look up specific version numbers and build steps.

1. Check out the Constellation repository at the tag corresponding to the release.

   ```bash
   git clone https://github.com/edgelesssys/constellation.git
   cd constellation
   git checkout v2.20.0
   ```

2. [Install the Bazel release](https://bazel.build/install) specified in `.bazelversion`.
3. [Install Nix](https://nixos.org/download/) (any recent version should do).
4. Run the build with `bazel build $target` for one of the following targets of interest:

   ```data
   //cli:cli_enterprise_darwin_amd64
   //cli:cli_enterprise_darwin_arm64
   //cli:cli_enterprise_linux_amd64
   //cli:cli_enterprise_linux_arm64
   //cli:cli_enterprise_windows_amd64
   ```

5. Compare the build result with the downloaded release artifact.

<!-- TODO(burgerdev): document reproducing images -->

## Feedback

Reproduction failures often indicate a bug in the build system or in the build definitions.
Therefore, we're interested in any reproducibility issues you might encounter.
[Start a bug report](https://github.com/edgelesssys/constellation/issues/new/choose) and describe the details of your build environment.
Make sure to include your result binary or a [`diffoscope`](https://diffoscope.org/) report, if possible.
