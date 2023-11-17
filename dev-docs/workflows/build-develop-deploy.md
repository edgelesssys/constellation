# Getting started locally

The following are instructions for building all components in the constellation repository, except for images. A manual on how to build images locally can be found in the [image README](/image/README.md).

Prerequisites:

* 20GB (minimum), better 40 GB disk space (required if you want to cross compile for all platforms)
* 16GB of unutilized RAM for a full Bazel build.
* [Latest version of Go](https://go.dev/doc/install).
* Unless you use Nix / NixOS: [Bazelisk installed as `bazel` in your path](https://github.com/bazelbuild/bazelisk/releases).
* We require Nix to be installed. It is recommended to install nix using the [determinate systems installer](https://github.com/DeterminateSystems/nix-installer) (or to use NixOS as host system).
* [Docker](https://docs.docker.com/engine/install/). Can be installed with these commands on Ubuntu 22.04: `sudo apt update && sudo apt install docker.io`. As the build spawns docker containers your user account either needs to be in the `docker` group (Add with `sudo usermod -a -G docker $USER`) or you have to run builds with `sudo`. When using `sudo` remember that your root user might (depending on your distro and local config) not have the go binary in it's PATH. The current PATH can be forwarded to the root env with `sudo env PATH=$PATH <cmd>`.

## Prequisites

### Linux

* If you don't want to perform any setup, you can get a shell with Bazel and all required dependencies by running:

  ```sh
  # better would be: nix develop -i
  # but this doesn't play nice with bashrc, colored output and non-hermetic tools
  nix develop
  ```

### Mac

* To fix unsupported shell options used in some build script:

  ```sh
  brew install bash
  ```

* To troubleshoot potential problems with bazel on ARM architecture when running it for the first time, it might help to purge and retry:

```sh
bazel clean --expunge
```

## Settings

### Authenticate with the GitHub registry

Get a GitHub token with access to the registry and authenticate through `docker` (see [here](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)).

Optionally, you can customize the default registry target:

```sh
echo "build --container_prefix=ghcr.io/USERNAME/constellation" >> .bazeloverwriterc
```

### Authenticate with your cloud provider

To provision the constellation cluster on the provider infrastructure, please authenticate with respective provider.

E.g. AWS:

```sh
aws configure
```

For more details, see [here](https://docs.edgeless.systems/constellation/getting-started/install#set-up-cloud-credentials).

## Build

>IMPORTANT: The OSS version is not identical with the official release. Notably, when executing `constellation create` you will need to set the image version and measurements yourself.

Build the binaries always in a separate directory. Since you can only have one IAM + cluster configuration in one directory, it is recommended to create a sub-directory for each cloud provider. This way you can easily deploy your dev-build on each of the providers by switching to their sub-directory:

```sh
mkdir build && cd build
# ( mkdir (aws|gcp|azure) && cd (aws|gcp|azure) )
# build required binaries for a dev build
# and symlink them into the current directory
# also push the built container images
bazel run //:devbuild --cli_edition=oss --container_prefix=ghcr.io/USERNAME/constellation
./constellation ...
```

>IMPORTANT: New images are private by default. To set it to public see [here](https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility).

### Specialized Bazel build targets

See [bazel](./bazel.md#build).

## Develop & Contribute

### VS Code setup

We recommend to set up your IDE to conform with our conventions (see [here](./dev-setup.md)).

### Testing

See [here](./testing.md) on how to run tests and how we write tests.

### Contributing a PR

#### Pre-Opening a PR

Before opening a PR, please run [these Bazel targets](./bazel.md#pre-pr-checks)

#### PR-Guidelines

See [here](./pull-request.md).

## Deploy

> :warning: Debug images are not safe to use in production environments. :warning:
The debug images will open an additional **unsecured** port (4000) which accepts any binary to be run on the target machine. **Make sure that those machines are not exposed to the internet.**

### Cloud

To familiarize yourself with debugd and learn how to deploy a cluster using it, read [this](/debugd/README.md) manual.
If you want to deploy a cluster for production, please refer to our user documentation [here](https://docs.edgeless.systems/constellation/getting-started/first-steps#create-a-cluster).

### Locally

In case you want to have quicker iteration cycles during development you might want to setup a local cluster.
You can do this by utilizing our QEMU setup.
Instructions on how to set it up can be found in the [QEMU README](qemu.md).

## Image export

To download an image you will have to export it first.
Below you find general instructions on how to do this for GCP and Azure.

### GCP

In order to download an image you will have to export it to a bucket you have access to:

* "Owner" permissions on the project
* "Storage Admin" permissions on the bucket
* Export with:

  ```bash
  gcloud compute images export --image=<image_path> --destination-uri=<bucket_uri> --export-format=qcow2 --project=<image_project>
  ```

* Click on "Download" on the created object

### Azure

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
