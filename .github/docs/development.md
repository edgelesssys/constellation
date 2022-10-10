# Build

The following are instructions for building all components in the constellation repository, except for images. A manual on how to build images locally can be found in the [image README](/image/README.md).

Prerequisites:

* 20 GB disk space
* [Go 1.18](https://go.dev/doc/install). Can be installed with these commands:

  ```sh
  wget https://go.dev/dl/go1.18.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz && export PATH=$PATH:/usr/local/go/bin
  ```

* [Docker](https://docs.docker.com/engine/install/). Can be installed with these commands on Ubuntu 22.04: `sudo apt update && sudo apt install docker.io`. As the build spawns docker containers your user account either needs to be in the `docker` group (Add with `sudo usermod -a -G docker $USER`) or you have to run builds with `sudo`. When using `sudo` remember that your root user might (depending on your distro and local config) not have the go binary in it's PATH. The current PATH can be forwarded to the root env with `sudo env PATH=$PATH <cmd>`.

* Packages on Ubuntu:

  ```sh
  sudo apt install build-essential cmake libssl-dev pkg-config libcryptsetup12 libcryptsetup-dev
  ```

* Packages on Fedora:

  ```sh
  sudo dnf install @development-tools pkg-config cmake openssl-devel cryptsetup-libs cryptsetup-devel
  ```

```sh
mkdir build
cd build
cmake ..
make -j`nproc`
```

# Test

You can run all integration and unitttests like this:

```sh
ctest -j `nproc`
```

You can limit the execution of tests to specific targets with e.g. `ctest -R unit-*` to only execute unit tests.

Some of the tests rely on libvirt and won't work if you don't have a virtualization capable CPU. You can find instructions on setting up libvirt in our [QEMU README](qemu.md).

# Deploy

> :warning: Debug images are not safe to use in production environments. :warning:
The debug images will open an additional **unsecured** port (4000) which accepts any binary to be run on the target machine. **Make sure that those machines are not exposed to the internet.**

## Cloud

To familiarize yourself with debugd and learn how to deploy a cluster using it, read [this](/debugd/README.md) manual.
If you want to deploy a cluster for production, please refer to our user documentation [here](https://docs.edgeless.systems/constellation/getting-started/first-steps#create-a-cluster).

## Locally

In case you want to have quicker iteration cycles during development you might want to setup a local cluster.
You can do this by utilizing our QEMU setup.
Instructions on how to set it up can be found in the [QEMU README](qemu.md).

# Verification

In order to verify your cluster we describe a [verification workflow](https://docs.edgeless.systems/constellation/workflows/verify-cluster) in our official docs.
Apart from that you can also reproduce some of the measurements described in the [docs](https://docs.edgeless.systems/constellation/architecture/attestation#runtime-measurements) locally.
To do so we built a tool that creates a VM, collects the PCR values and reports them to you.
To run the tool execute the following command in `/hack/image-measurement`:

```sh
go run . -path <image_path> -type <image_type>
```

`<image_path>` needs to point to a valid image file.
The image can be either in raw or QEMU's `qcow2` format.
This format is specified in the `<image_type>` argument.

You can compare the values of PCR 4, 8 and 9 to the ones you are seeing in your `constellation-conf.yaml`.
The PCR values depend on the image you specify in the `path` argument.
Therefore, if you want to verify a cluster deployed with a release image you will have to download the images first.

After collecting the measurements you can put them into your `constellation-conf.yaml` under the `measurements` key in order to enforce them.

# Image export

To download an image you will have to export it first.
Below you find general instructions on how to do this for GCP and Azure.
You can find values for `<image_path>` in the `version_manifest.json` that is part of each constellation release.

## GCP

In order to download an image you will have to export it to a bucket you have access to:

* "Owner" permissions on the project
* "Storage Admin" permissions on the bucket
* Export with:

  ```bash
  gcloud compute images export --image=<image_path> --destination-uri=<bucket_uri> --export-format=qcow2 --project=<image_project>
  ```

* Click on "Download" on the created object

## Azure

To download an image from Azure you will have to create a disk from the image and generate a download link for that disk:

```bash
#!/usr/bin/env bash

VERSION=0.0.1
TARGET_DISK=export-${VERSION}

az disk create -g <resource_group> -l <target_region> -n $TARGET_DISK --hyper-v-generation V2 --os-type Linux --sku standard_lrs --security-type TrustedLaunch --gallery-image-reference <image_path>

az disk grant-access -n $TARGET_DISK -g constellation-images --access-level Read --duration-in-seconds 3600 | jq -r .accessSas
```

Depending on you internet connection you might have to modify the duration value.
The duration value specifies for how long the link is usable.
