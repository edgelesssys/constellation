# Build

The following are instructions for building all components in the constellation repository, except for images. A manual on how to build images locally can be found in the [image README](/image/README.md).

Prerequisites:

* 20GB (minimum), better 40 GB disk space (required if you want to cross compile for all platforms)
* [Latest version of Go](https://go.dev/doc/install).
* [Bazelisk installed as `bazel` in your path](https://github.com/bazelbuild/bazelisk/releases).
* [Docker](https://docs.docker.com/engine/install/). Can be installed with these commands on Ubuntu 22.04: `sudo apt update && sudo apt install docker.io`. As the build spawns docker containers your user account either needs to be in the `docker` group (Add with `sudo usermod -a -G docker $USER`) or you have to run builds with `sudo`. When using `sudo` remember that your root user might (depending on your distro and local config) not have the go binary in it's PATH. The current PATH can be forwarded to the root env with `sudo env PATH=$PATH <cmd>`.

---
### On Linux
* Packages on Ubuntu:

  ```sh
  sudo apt install build-essential cmake libssl-dev pkg-config libcryptsetup12 libcryptsetup-dev
  ```

* Packages on Fedora:

  ```sh
  sudo dnf install @development-tools pkg-config cmake openssl-devel cryptsetup-libs cryptsetup-devel
  ```

### On Mac

```
brew install bash
```
to fix unsupported shell options used in some build script.

To troubleshoot potential problems with bazel on ARM architecture when running it for the first time, it might help to purge and retry:
```
bazel clean --expunge
```

---

Developer workspace:

```sh
mkdir build
cd build
# build required binaries for a dev build
# and symlink them into the current directory
# also push the built container images
# After the first run, set the pushed imaged to public.
bazel run //:devbuild --container_prefix=ghcr.io/USERNAME/constellation
./constellation ...
```

Overwrite the default container_prefix in the `.bazeloverwriterc` in the root of the workspace:
```bazel
# cat .bazeloverwriterc
build --container_prefix=ghcr.io/USERNAME/constellation
```

Bazel build:

```sh
bazel query //...
bazel build //path/to:target
bazel build //... # build everything (for every supported platform)
bazel build //bootstrapper/cmd/bootstrapper:bootstrapper # build bootstrapper
bazel build //cli:cli_oss # build CLI
bazel build //cli:cli_oss_linux_amd64 # cross compile CLI for linux amd64
bazel build //cli:cli_oss_linux_arm64 # cross compile CLI for linux arm64
bazel build //cli:cli_oss_darwin_amd64 # cross compile CLI for mac amd64
bazel build //cli:cli_oss_darwin_arm64 # cross compile CLI for mac arm64
```

## Remote caching and execution

We use BuildBuddy for remote caching (and maybe remote execution in the future). To use it, you need to join the BuildBuddy organization and get an API key. Then, you can write it to `~/.bazelrc`:

```
build --remote_header=x-buildbuddy-api-key=<redacted>
```

To use the remote cache, build the project with `bazel build --config remote_cache //path/to:target`.
You can also copy the `remote_cache` config from `.bazelrc` to your `~/.bazelrc` and remove the `remote_cache` prefix to make it the default.

# Test

You can run all integration and unitttests like this:

```sh
ctest -j `nproc`
```

You can limit the execution of tests to specific targets with e.g. `ctest -R unit-*` to only execute unit tests.

Some of the tests rely on libvirt and won't work if you don't have a virtualization capable CPU. You can find instructions on setting up libvirt in our [QEMU README](qemu.md).

Running unit tests with Bazel:

```sh
bazel test //...
```

# Opening a PR
Before opening a PR, please run the tests and
```
bazel run //:generate && bazel run //:tidy
bazel run //:check
```

The linter check doesn't work on Mac at the moment, but you can run the linter directly:
```
golangci-lint run
```
Furthermore, the PR titles are used for the changelog, so please stick to our [conventions](https://github.com/edgelesssys/constellation/blob/main/dev-docs/conventions.md#pr-conventions).

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
Use the provided scripts in `/image/measured-boot` to generated measurements for a built image. Measurements for release images are also available in our image API.

# Dependency management

## Go

Go dependencies are managed with [Go modules](https://blog.golang.org/using-go-modules) that are all linked from the main [go.work](/go.work) file.
[Follow the go documentation](https://go.dev/doc/modules/managing-dependencies) on how to use Go modules.

# Bazel

Bazel is the primary build system for this project. It is used to build all Go code and will be used to build all artifacts in the future.
Still, we aim to keep the codebase compatible with `go build` and `go test` as well.
Whenever Go code is changed, you will have to run `bazel run //:tidy` to regenerate the Bazel build files for Go code.

## Bazel commands

* `bazel query //...` - list all targets
* `bazel query //subfolder` - list all targets in a subfolder
* `bazel build //...` - build all targets
* `bazel build //subfolder/...` - build all targets in a subfolder (recursive)
* `bazel build //subfolder:all` - build all targets in a subfolder (non-recursive)
* `bazel build //subfolder:target` - build single target
* `bazel run --run_under="cd $PWD &&" //cli:cli_oss -- create -c 1 -w 1` - build + run a target with arguments in current working directory
* `bazel cquery --output=files //subfolder:target` - get location of a build artifact
* `bazel test //...` - run all tests
* `bazel run //:tidy` - tidy, format and generate
* `bazel run //:check` - execute checks and linters
* `bazel run //:generate` - execute code generation

## Editor integration

You can continue to use the default Go language server and editor integration. This will show you different paths for external dependencies and not use the Bazel cache.
Alternatively, you can use [the go language server integration for Bazel](https://github.com/bazelbuild/rules_go/wiki/Editor-setup). This will use Bazel for dependency resolution and execute Bazel commands for building and testing.

# Image export

To download an image you will have to export it first.
Below you find general instructions on how to do this for GCP and Azure.

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
