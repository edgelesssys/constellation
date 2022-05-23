# Constellation

This is the main repository of Constellation.

Core components:

* [cli](cli): The CLI is used to manage a Constellation cluster
* [coordinator](coordinator): The Coordinator is a node agent whose most important task is to bootstrap a node
* [image](image): Build files for the Constellation disk image
* [kms](kms): Constellation's key management client and server
* [mount](mount): Package used by CSI plugins to create and mount encrypted block devices
* [state](state): Contains the disk-mapper that maps the encrypted node data disk during boot

Development components:

* [conformance](conformance): Kubernetes conformance tests
* [debugd](debugd): Debug daemon and client
* [hack](hack): Development tools
* [proto](proto): Proto files generator
* [test](test): Integration test

Additional repositories:

* [constellation-docs](https://github.com/edgelesssys/constellation-docs): End-user documentation
* [constellation-coreos-assembler](https://github.com/edgelesssys/constellation-coreos-assembler): Build environment for CoreOS images with changes for Constellation
* [constellation-fedora-coreos-config](https://github.com/edgelesssys/constellation-fedora-coreos-config): CoreOS build configuration with changes for Constellation
* [edg-azuredisk-csi-driver](https://github.com/edgelesssys/edg-azuredisk-csi-driver): Azure CSI driver with encryption on node
* [edg-gcp-compute-persistent-disk-csi-driver](https://github.com/edgelesssys/edg-gcp-compute-persistent-disk-csi-driver): GCP CSI driver with encryption on node

## Build

Prerequisites:

* [Go 1.18](https://go.dev/doc/install)
* [Docker](https://docs.docker.com/engine/install/)
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

## Cloud credentials

Using the CLI requires the user to make authorized API calls to the CSP API. See the [docs](https://constellation-docs.edgeless.systems/6c320851-bdd2-41d5-bf10-e27427398692/#/getting-started/install?id=cloud-credentials) for configuration.

## Deploying a locally compiled coordinator binary

By default, `constellation create ...` will spawn cloud provider instances with a pre-baked coordinator binary.
For testing, you can use the constellation debug daemon (debugd) to upload your local coordinator binary to running instances and to obtain SSH access.
[Follow this introduction on how to install and setup `cdbg`](debugd/README.md)

## Development Guides

* [Upgrading Kubernetes](/docs/upgrade-kubernetes.md)
* [Local image testing](/docs/local-image-testing.md)

## Deployment Guides

* [Onboarding Customers](/docs/onboarding-customers.md)
