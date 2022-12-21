## Setup

- Install mkosi (from git):

    ```sh
    cd /tmp/
    git clone https://github.com/systemd/mkosi
    cd mkosi
    tools/generate-zipapp.sh
    cp builddir/mkosi /usr/local/bin/
    ```

- Install tools:

    <details>
    <summary>Ubuntu / Debian</summary>

    ```sh
    sudo apt-get update
    sudo apt-get install --assume-yes --no-install-recommends \
        dnf \
        systemd-container \
        qemu-system-x86 \
        qemu-utils \
        ovmf \
        e2fsprogs \
        squashfs-tools \
        efitools \
        sbsigntool \
        coreutils \
        curl \
        jq \
        util-linux \
        virt-manager \
        python3-crc32c \
        rpm
    ```

    </details>

    <details>
    <summary>Fedora</summary>

    ```sh
    sudo dnf install -y \
        edk2-ovmf \
        systemd-container \
        qemu \
        e2fsprogs \
        squashfs-tools \
        efitools \
        sbsigntools \
        coreutils \
        curl \
        jq \
        util-linux \
        virt-manager
    ```

    </details>

- Prepare secure boot PKI (see `secure-boot/genkeys.sh`)

## Build

When building your first image, prepare the secure boot PKI (see `secure-boot/genkeys.sh`) for self-signed, locally built images.

After that, you can build the image with:

```sh
# OPTIONAL: to create a debug image, export the following line
# export BOOTSTRAPPER_BINARY=$(realpath ${PWD}/../../build/debugd)
# OPTIONAL: symlink custom path to secure boot PKI to ./pki
# ln -s /path/to/pki/folder ./pki
sudo make -j $(nproc)
```

Raw images will be placed in `mkosi.output.<CSP>/fedora~36/image.raw`.

## Prepare Secure Boot

The generated images are partially signed by Microsoft ([shim loader](https://github.com/rhboot/shim)), and partially signed by Edgeless Systems (systemd-boot and unified kernel images consisting of the linux kernel, initramfs and kernel commandline).

For QEMU and Azure, you can pre-generate the NVRAM variables for secure boot. This is not necessary for GCP, as you can specify secure boot parameters via the GCP API on image creation.

<details>
<summary><a id="qemu-secure-boot">libvirt / QEMU / KVM</a></summary>

```sh
secure-boot/generate_nvram_vars.sh mkosi.output.qemu/fedora~36/image.raw
```

</details>

<details>
<summary><a id="azure-secure-boot">Azure</a></summary>

These steps only have to performed once for a fresh set of secure boot certificates.
VMGS blobs for testing and release images already exist.

First, create a disk without embedded MOK EFI variables.

```sh
# set these variables
export AZURE_SECURITY_TYPE=ConfidentialVM # or TrustedLaunch
export AZURE_RESOURCE_GROUP_NAME= # e.g. "constellation-images"

export AZURE_REGION=northeurope
export AZURE_REPLICATION_REGIONS=
export AZURE_DISK_NAME=constellation-$(date +%s)
export AZURE_SNAPSHOT_NAME=${AZURE_DISK_NAME}
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.vhd
export AZURE_VMGS_FILENAME=${AZURE_SECURITY_TYPE}.vmgs
export AZURE_JSON_OUTPUT=${PWD}/mkosi.output.azure/fedora~36/image-upload.json
export BLOBS_DIR=${PWD}/blobs
upload/pack.sh azure "${AZURE_RAW_IMAGE_PATH}" "${AZURE_IMAGE_PATH}"
upload/upload_azure.sh --disk-name "${AZURE_DISK_NAME}-setup-secure-boot" ""
secure-boot/azure/launch.sh -n "${AZURE_DISK_NAME}-setup-secure-boot" -d --secure-boot true --disk-name "${AZURE_DISK_NAME}-setup-secure-boot"
```

Ignore the running launch script and connect to the serial console once available.
The console shows the message "Verification failed: (0x1A) Security Violation". You can import the MOK certificate via the UEFI shell:

Press OK, then ENTER, then "Enroll key from disk".
Select the following key: `/EFI/loader/keys/auto/db.cer`.
Press Continue, then choose "Yes" to the question "Enroll the key(s)?".
Choose reboot.

Extract the VMGS from the running VM (this includes the MOK EFI variables) and delete the VM:

```sh
secure-boot/azure/extract_vmgs.sh --name "${AZURE_DISK_NAME}-setup-secure-boot"
secure-boot/azure/delete.sh --name "${AZURE_DISK_NAME}-setup-secure-boot"
```

</details>

## Upload to CSP

<details>
<summary>AWS</summary>

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))
- Choose secure boot PKI public keys (one of `pki_dev`, `pki_test`, `pki_prod`)
    - `pki_dev` can be used for local image builds
    - `pki_test` is used by the CI for non-release images
    - `pki_prod` is used for release images

```sh
# set these variables
export AWS_IMAGE_NAME= # e.g. "constellation-v1.0.0"
export PKI=${PWD}/pki

export AWS_REGION=eu-central-1
export AWS_REPLICATION_REGIONS="us-east-2"
export AWS_BUCKET=constellation-images
export AWS_EFIVARS_PATH=${PWD}/mkosi.output.aws/fedora~36/efivars.bin
export AWS_IMAGE_PATH=${PWD}/mkosi.output.aws/fedora~36/image.raw
export AWS_IMAGE_FILENAME=image-$(date +%s).raw
export AWS_JSON_OUTPUT=${PWD}/mkosi.output.aws/fedora~36/image-upload.json
secure-boot/aws/create_uefivars.sh "${AWS_EFIVARS_PATH}"
upload/upload_aws.sh
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
# set these variables
export GCP_IMAGE_FAMILY= # e.g. "constellation"
export GCP_IMAGE_NAME= # e.g. "constellation-v1.0.0"
export PKI=${PWD}/pki

export GCP_PROJECT=constellation-images
export GCP_REGION=europe-west3
export GCP_BUCKET=constellation-images
export GCP_RAW_IMAGE_PATH=${PWD}/mkosi.output.gcp/fedora~36/image.raw
export GCP_IMAGE_FILENAME=$(date +%s).tar.gz
export GCP_IMAGE_PATH=${PWD}/mkosi.output.gcp/fedora~36/image.tar.gz
export GCP_JSON_OUTPUT=${PWD}/mkosi.output.gcp/fedora~36/image-upload.json
upload/pack.sh gcp ${GCP_RAW_IMAGE_PATH} ${GCP_IMAGE_PATH}
upload/upload_gcp.sh
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
# set these variables
export AZURE_GALLERY_NAME= # e.g. "Constellation"
export AZURE_IMAGE_DEFINITION= # e.g. "constellation"
export AZURE_IMAGE_VERSION= # e.g. "1.0.0"
# Set this variable to a path if you want to use Secure Boot.
# Otherwise, set it to export AZURE_VMGS_PATH=
export AZURE_VMGS_PATH= # e.g. nothing OR "path/to/ConfidentialVM.vmgs"
# AZURE_SECURITY_TYPE can be one of
# - "ConfidentialVMSupported" (ConfidentialVM with secure boot disabled),
# - "ConfidentialVM" (ConfidentialVM with Secure Boot) or
# - TrustedLaunch" (Trusted Launch with or without Secure Boot)
export AZURE_SECURITY_TYPE=ConfidentialVMSupported

export AZURE_RESOURCE_GROUP_NAME=constellation-images
export AZURE_REGION=northeurope
export AZURE_REPLICATION_REGIONS="northeurope eastus westeurope westus"
export AZURE_IMAGE_OFFER=constellation
export AZURE_SKU=${AZURE_IMAGE_DEFINITION}
export AZURE_PUBLISHER=edgelesssys
export AZURE_DISK_NAME=constellation-$(date +%s)
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.vhd
export AZURE_JSON_OUTPUT=${PWD}/mkosi.output.azure/fedora~36/image-upload.json
upload/pack.sh azure "${AZURE_RAW_IMAGE_PATH}" "${AZURE_IMAGE_PATH}"
upload/upload_azure.sh -g --disk-name "${AZURE_DISK_NAME}" "${AZURE_VMGS_PATH}"
```

</details>

<details>
<summary>QEMU</summary>

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))

```sh
# set these variables
export REF= # e.g. feat-xyz (branch name encoded with dashes)
export STREAM= # e.g. "nightly", "debug", "stable" (depends on the type of image and if it is a release)
export IMAGE_VERSION= # e.g. v2.1.0" or output of pseudo-version tool
export QEMU_BUCKET=cdn-constellation-backend
export QEMU_BASE_URL="https://cdn.confidential.cloud"
export QEMU_IMAGE_PATH=${PWD}/mkosi.output.qemu/fedora~36/image.raw
export QEMU_JSON_OUTPUT=${PWD}/mkosi.output.qemu/fedora~36/image-upload.json
upload/upload_qemu.sh
```

</details>
