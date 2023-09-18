## Setup

Ensure you have Nix installed. This is recommended in general but a requirement for the following steps.
Consult the [developer docs](/dev-docs/workflows/build-develop-deploy.md) for more info.
At the very least, `nix` should be in your PATH and either `common --config=nix`
has to be set in the `.bazelrc` or you need to append `--config=nix` to each Bazel command.

## Build

You can build any image using Bazel.
Start by querying the available images:

```sh
bazel query //image/system/...
```

You can either build a group of images (all images for a cloud provider, a stream, ...) or a single image by selecting a target.

```sh
bazel build //image/system:openstack_qemu-vtpm_debug
```

The location of the destination folder can be queried like this:

```sh
bazel cquery --output=files //image/system:openstack_qemu-vtpm_debug
```

## Upload to CSP

Warning! Never set `--version` to a value that is already used for a release image.

<details>
<summary>AWS</summary>

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))
- Choose secure boot PKI public keys (one of `pki_dev`, `pki_test`, `pki_prod`)
  - `pki_dev` can be used for local image builds
  - `pki_test` is used by the CI for non-release images
  - `pki_prod` is used for release images

```sh
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- image aws --verbose --raw-image path/to/constellation.raw --attestation-variant ""  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

<details>
<summary>GCP</summary>

- Install `gcloud` and `gsutil` (see [here](https://cloud.google.com/sdk/docs/install))
- Login to GCP (see [here](https://cloud.google.com/sdk/docs/authorizing))
- Choose secure boot PKI public keys (one of `pki_dev`, `pki_test`, `pki_prod`)
  - `pki_dev` can be used for local image builds
  - `pki_test` is used by the CI for non-release images
  - `pki_prod` is used for release images

```sh
export GCP_RAW_IMAGE_PATH=$(realpath path/to/constellation.raw)
export GCP_IMAGE_PATH=path/to/image.tar.gz
upload/pack.sh gcp ${GCP_RAW_IMAGE_PATH} ${GCP_IMAGE_PATH}
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- image gcp --verbose --raw-image "${GCP_IMAGE_PATH}" --attestation-variant "sev-es"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

<details>
<summary>Azure</summary>

Note:

> For testing purposes, it is a lot simpler to disable Secure Boot for the uploaded image!
> Disabling Secure Boot allows you to skip the VMGS creation steps above.

- Install `az` and `azcopy` (see [here](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli))
- Login to Azure (see [here](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli))
- Optional (if Secure Boot should be enabled) [Prepare virtual machine guest state (VMGS) with customized NVRAM or use existing VMGS blob](#azure-secure-boot)

```sh
export AZURE_RAW_IMAGE_PATH=path/to/constellation.raw
export AZURE_IMAGE_PATH=path/to/image.vhd
upload/pack.sh azure "${AZURE_RAW_IMAGE_PATH}" "${AZURE_IMAGE_PATH}"
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- image azure --verbose --raw-image "${AZURE_IMAGE_PATH}" --attestation-variant "cvm"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

<details>
<summary>OpenStack</summary>

Note:

> OpenStack is not one a global cloud provider, but rather a software that can be installed on-premises.
> This means we do not upload the image to a cloud provider, but to our CDN.

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))

```sh
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- image openstack --verbose --raw-image path/to/constellation.raw --attestation-variant "sev"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

<details>
<summary>QEMU</summary>

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))

```sh
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- image qemu --verbose --raw-image path/to/constellation.raw --attestation-variant "default"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

## Kernel

The Kernel is built from the srpm published under [edgelesssys/constellation-kernel](https://github.com/edgelesssys/constellation-kernel).
We track the latest longterm release, use sources directly from [kernel.org](https://www.kernel.org/) and build the Kernel using the steps specified in the
srpm spec file.

After building a Kernel rpm, we upload it to our CDN and use it in our image builds.
