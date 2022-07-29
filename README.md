# Constellation

This is the main repository of Constellation.

![E2ETestAzure](https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-azure.yml/badge.svg?branch=main)
![E2ETestGCP](https://github.com/edgelesssys/constellation/actions/workflows/e2e-test-gcp.yml/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/edgelesssys/constellation)](https://goreportcard.com/report/github.com/edgelesssys/constellation)
[![Discord Chat](https://img.shields.io/badge/chat-on%20Discord-blue)](https://discord.gg/rH8QTH56JN)

Core components:

* [access_manager](access_manager): Contains the access-manager pod used to persist SSH users based on a K8s ConfigMap
* [cli](cli): The CLI is used to manage a Constellation cluster
* [bootstrapper](bootstrapper): The bootstrapper is a node agent whose most important task is to bootstrap a node
* [image](image): Build files for the Constellation disk image
* [kms](kms): Constellation's key management client and server
* [mount](mount): Package used by CSI plugins to create and mount encrypted block devices
* [state](state): Contains the disk-mapper that maps the encrypted node data disk during boot

Development components:

* [conformance](conformance): Kubernetes conformance tests
* [debugd](debugd): Debug daemon and client
* [hack](hack): Development tools
* [proto](proto): Proto files generator
* [terraform](terraform): Infrastructure management using terraform (instead of `constellation create/destroy`)
  * [libvirt](terraform/libvirt): Deploy local cluster using terraform, libvirt and QEMU
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

## Testing

You can run all integration and unitttests like this:

```sh
ctest -j `nproc`
```

## Cloud credentials

Using the CLI requires the user to make authorized API calls to the CSP API. See the [docs](https://constellation-docs.edgeless.systems/6c320851-bdd2-41d5-bf10-e27427398692/#/getting-started/install?id=cloud-credentials) for configuration.

## Deploying a locally compiled bootstrapper binary

By default, `constellation create ...` will spawn cloud provider instances with a pre-baked bootstrapper binary.
For testing, you can use the constellation debug daemon (debugd) to upload your local bootstrapper binary to running instances and to obtain SSH access.
[Follow this introduction on how to install and setup `cdbg`](debugd/README.md)

## Development Guides

* [Upgrading Kubernetes](/docs/upgrade-kubernetes.md)
* [Manual local image testing](/docs/local-image-testing.md)

## Deployment Guides

* [Onboarding Customers](/docs/onboarding-customers.md)
