# Constellation images

We use the [Fedora CoreOS Assembler](https://coreos.github.io/coreos-assembler/) to build the base image for Constellation nodes.

## Setup

1. Install prerequisites:
   - [Docker](https://docs.docker.com/engine/install/) or [Podman](https://podman.io/getting-started/installation)
   - [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-linux)
   - [azcopy](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-v10)
   - [Google Cloud CLI](https://cloud.google.com/sdk/docs/install)
   - [gsutil](https://cloud.google.com/storage/docs/gsutil_install#linux)
   - Ubuntu:

        ```shell-session
        sudo apt install -y bash coreutils cryptsetup-bin grep libguestfs-tools make parted pv qemu-system qemu-utils sed tar util-linux wget
        ```

2. Log in to GCP and Azure

   ```shell-session
   gcloud auth login
   az login
   ```

3. [Log in to the ghcr.io package registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)
4. Ensure read and write access to `/dev/kvm` (and repeat after every reboot)

    ```shell-session
    sudo chmod 666 /dev/kvm
    ```

## Configuration

Create a configuration file in `image/config.mk` to override any of the variables found at the top of the [Makefile](Makefile).
Important settings are:

- `BOOTSTRAPPER_BINARY`: path to a bootstrapper binary. Can be substituted with a path to a `debugd` binary if a debug image should be built. The binary has to be built before!
- `CONTAINER_ENGINE`: container engine used to run COSA. either `podman` or `docker`.
- `COSA_INIT_REPO`: Git repository containing CoreOS config. Cloned in `cosa-init` target.
- `COSA_INIT_BRANCH`: Git branch checked out from `COSA_INIT_REPO`. Can be used to test out changes on another branch before merging.
- `NETRC` path to a netrc file containing a GitHub PAT. Used to authenticate to GitHub from within the COSA container.
- `GCP_IMAGE_NAME`: Image name for the GCP image. Set to include a timestamp when using the build pipeline. Can be set to a custom value if you wat to upload a custom image for testing on GCP.
- `AZURE_IMAGE_NAME`: Image name for the Azure image. Can be set to a custom value if you wat to upload a custom image for testing on Azure.

Example `config.mk` to create a debug image with docker and name it `my-custom-image`:

```Makefile
BOOTSTRAPPER_BINARY = ../build/debugd
CONTAINER_ENGINE = docker
GCP_IMAGE_NAME = my-custom-image
AZURE_IMAGE_NAME = my-custom-image
```

## Build an image

> It is always advisable to create an image from a clean `build` dir.

Clean up the `build` dir and remove old images (⚠ this will undo any local changes to the CoreOS configuration!):

```shell-session
sudo make clean
```

- Build QEMU image (for local testing only)

  ```shell-session
  make coreos
  ```

- Build Azure image (without upload)

  ```shell-session
  make image-azure
  ```

- Build Azure image (with upload)

  ```shell-session
  make image-azure upload-azure
  ```

- Build GCP image (without upload)

  ```shell-session
  make image-gcp
  ```

- Build GCP image (with upload)

  ```shell-session
  make image-gcp upload-gcp
  ```

Resulting images for the CSPs can be found under [images](images/). QEMU images are stored at `build/builds/latest/` with a name ending in `.qcow2`.
