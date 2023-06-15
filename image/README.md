## Setup

- Install mkosi (from git):

    ```sh
    cd /tmp/
    git clone https://github.com/systemd/mkosi
    cd mkosi
    git checkout d8b32fbf3077b612db0024276e73cec3c2c87577
    tools/generate-zipapp.sh
    cp builddir/mkosi /usr/local/bin/
    ```

- Build systemd tooling (from git):

    Ubuntu and Fedora ship outdated versions of systemd tools, so you need to build them from source:

    ```sh
    # Ubuntu
    echo "deb-src http://archive.ubuntu.com/ubuntu/ $(lsb_release -cs) main restricted universe multiverse" | sudo tee -a /etc/apt/sources.list
    sudo apt-get update
    sudo apt-get build-dep systemd
    sudo apt-get install libfdisk-dev
    # Fedora
    sudo dnf builddep systemd

    git clone https://github.com/systemd/systemd --depth=1
    meson systemd/build systemd -Drepart=true -Defi=true -Dbootloader=true
    BINARIES=(
        bootctl
        systemctl
        systemd-analyze
        systemd-dissect
        systemd-nspawn
        systemd-repart
        ukify
    )
    ninja -C systemd/build ${BINARIES[@]}
    SYSTEMD_BIN=$(realpath systemd/build)
    echo installed systemd tools to "${SYSTEMD_BIN}"
    ```

- Install tools:

    <details>
    <summary>Ubuntu / Debian</summary>

    ```sh
    sudo apt-get update
    sudo apt-get install --assume-yes --no-install-recommends \
        bubblewrap \
        coreutils \
        curl \
        dnf \
        e2fsprogs \
        efitools \
        jq \
        mtools \
        ovmf \
        python3-pefile \
        python3-pyelftools \
        python3-setuptools \
        qemu-system-x86 \
        qemu-utils \
        rpm \
        sbsigntool \
        squashfs-tools \
        systemd-container \
        util-linux \
        virt-manager
    ```

    </details>

    <details>
    <summary>Fedora</summary>

    ```sh
    sudo dnf install -y \
        bubblewrap \
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
# export SYSTEMD_BIN=<path to systemd tools>
# OPTIONAL: to create a debug image, export the following line
# export DEBUG=true
# OPTIONAL: to enable the serial console, export the following line
# export AUTOLOGIN=true
# OPTIONAL: symlink custom path to secure boot PKI to ./pki
# ln -s /path/to/pki/folder ./pki
sudo make EXTRA_SEARCH_PATHS="${SYSTEMD_BIN}" -j $(nproc)
```

Raw images will be placed in `mkosi.output.<CSP>/fedora~38/image.raw`.

## Prepare Secure Boot

The generated images are partially signed by Microsoft ([shim loader](https://github.com/rhboot/shim)), and partially signed by Edgeless Systems (systemd-boot and unified kernel images consisting of the linux kernel, initramfs and kernel commandline).

For QEMU and Azure, you can pre-generate the NVRAM variables for secure boot. This is not necessary for GCP, as you can specify secure boot parameters via the GCP API on image creation.

<details>
<summary><a id="qemu-secure-boot">libvirt / QEMU / KVM</a></summary>

```sh
secure-boot/generate_nvram_vars.sh mkosi.output.qemu/fedora~38/image.raw
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
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~38/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~38/image.vhd
export AZURE_VMGS_FILENAME=${AZURE_SECURITY_TYPE}.vmgs
export AZURE_JSON_OUTPUT=${PWD}/mkosi.output.azure/fedora~38/image-upload.json
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
bazel run //image/upload -- aws --verbose --raw-image mkosi.output.aws/fedora~38/image.raw --variant ""  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
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
export GCP_RAW_IMAGE_PATH=${PWD}/mkosi.output.gcp/fedora~38/image.raw
export GCP_IMAGE_PATH=${PWD}/mkosi.output.gcp/fedora~38/image.tar.gz
upload/pack.sh gcp ${GCP_RAW_IMAGE_PATH} ${GCP_IMAGE_PATH}
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- gcp --verbose --raw-image "${GCP_IMAGE_PATH}" --variant "sev-es"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
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
export AZURE_RAW_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~38/image.raw
export AZURE_IMAGE_PATH=${PWD}/mkosi.output.azure/fedora~38/image.vhd
upload/pack.sh azure "${AZURE_RAW_IMAGE_PATH}" "${AZURE_IMAGE_PATH}"
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- azure --verbose --raw-image "${AZURE_IMAGE_PATH}" --variant "cvm"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
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
bazel run //image/upload -- openstack --verbose --raw-image mkosi.output.openstack/fedora~38/image.raw --variant "sev"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>

<details>
<summary>QEMU</summary>

- Install `aws` cli (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- Login to AWS (see [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html))

```sh
# Warning! Never set `--version` to a value that is already used for a release image.
# Instead, use a `ref` that corresponds to your branch name.
bazel run //image/upload -- qemu --verbose --raw-image mkosi.output.qemu/fedora~38/image.raw --variant "default"  --version ref/foo/stream/nightly/v2.7.0-pre-asdf
```

</details>
