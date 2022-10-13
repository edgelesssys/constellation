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
        virt-manager
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

```sh
# OPTIONAL: to create a debug image, export the following line
# export BOOTSTRAPPER_BINARY=$(realpath ${PWD}/../../build/debugd)
# OPTIONAL: specify path to secure boot PKI
# export PKI=/path/to/pki/folder
sudo make -j $(nproc)
```

Raw images will be placed in `mkosi.output.<CSP>/fedora~36/image.raw`.

## Prepare Secure Boot

The generated images are partially signed by Microsoft ([shim loader](https://github.com/rhboot/shim)), and partially signed by Edgeless Systems (systemd-boot and unified kernel images consisting of the linux kernel, initramfs and kernel commandline).

For QEMU and Azure, you can pre-generate the NVRAM variables for secure boot. This is not necessary for GCP, as you can specify secure boot parameters via the GCP API on image creation.

<details>
<summary>libvirt / QEMU / KVM</summary>

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
export AZURE_DISK_NAME=constellation-$(date +%s)
export AZURE_SNAPSHOT_NAME=${AZURE_DISK_NAME}
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.vhd
export AZURE_VMGS_FILENAME=${AZURE_SECURITY_TYPE}.vmgs
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
<summary>GCP</summary>

- Install `gcloud` and `gsutil` (see [here](https://cloud.google.com/sdk/docs/install))
- Login to GCP (see [here](https://cloud.google.com/sdk/docs/authorizing))
- Prepare secure boot PKI (see `secure-boot/genkeys.sh`)

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
upload/pack.sh gcp ${GCP_RAW_IMAGE_PATH} ${GCP_IMAGE_PATH}
upload/upload_gcp.sh
```

</details>

<details>
<summary>Azure</summary>

- Install `az` and `azcopy` (see [here](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli))
- Login to Azure (see [here](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli))
- Prepare secure boot PKI (see `secure-boot/genkeys.sh`)
- [Prepare virtual machine guest state (VMGS) with customized NVRAM or use existing VMGS blob](#azure-secure-boot)

```sh
# set these variables
export AZURE_GALLERY_NAME= # e.g. "Constellation"
export AZURE_IMAGE_DEFINITION= # e.g. "constellation"
export AZURE_IMAGE_VERSION= # e.g. "1.0.0"
export AZURE_VMGS_PATH= # e.g. "path/to/ConfidentialVM.vmgs"
export AZURE_SECURITY_TYPE=ConfidentialVM # or TrustedLaunch

export AZURE_RESOURCE_GROUP_NAME=constellation-images
export AZURE_REGION=northeurope
export AZURE_REPLICATION_REGIONS="northeurope eastus westeurope westus"
export AZURE_IMAGE_OFFER=constellation
export AZURE_SKU=constellation
export AZURE_PUBLISHER=edgelesssys
export AZURE_DISK_NAME=constellation-$(date +%s)
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~36/image.vhd
upload/pack.sh azure "${AZURE_RAW_IMAGE_PATH}" "${AZURE_IMAGE_PATH}"
upload/upload_azure.sh -g --disk-name "${AZURE_DISK_NAME}" "${AZURE_VMGS_PATH}"
```

</details>
