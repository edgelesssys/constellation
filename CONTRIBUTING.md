## First steps

Thank you for getting involved! Before you start, please familiarize yourself with the [documentation](https://constellation-docs.edgeless.systems/6c320851-bdd2-41d5-bf10-e27427398692).

Please follow our [Code of Conduct](CODE_OF_CONDUCT.md) when interacting with this project.

If you want to support our development:

* Add a GitHub Star to the project
* Share our projects on social media
* Join the [Confidential Computing Discord](https://discord.gg/rH8QTH56JN)

Constellation is licensed under the [TODO](LICENSE). When contributing, you also need to agree to our [Contributor License Agreement](https://cla-assistant.io/edgelesssys/constellation).

## Development guidelines

Adhere to the style and best practices described in [Effective Go](https://golang.org/doc/effective_go.html). Read [Common Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for further information.

## Pull request process

Submissions should remain focused in scope and avoid containing unrelated commits.
For pull requests, we employ the following workflow:

1. Fork the repository to your own GitHub account
2. Create a branch locally with a descriptive name
3. Commit changes to the branch
4. Write your code according to our development guidelines
5. Push changes to your fork
6. Clean up your commit history
7. Open a PR in our repository and summarize the changes in the description

## Reporting issues and bugs, asking questions

This project uses the GitHub issue tracker. Please check the existing issues before submitting to avoid duplicates.

To report a security issue, contact security@edgeless.systems.

Your bug report should cover the following points:

* A quick summary and/or background of the issue
* Steps to reproduce (be specific, e.g., provide sample code)
* What you expected would happen
* What actually happens
* Further notes:
  * Thoughts on possible causes
  * Tested workarounds or fixes

## Major changes and feature requests

You should discuss larger changes and feature requests with the maintainers. Please open an issue describing your plans.

[Run CI e2e tests](/.github/docs/README.md)

## Repository Layout

Core components:

* [access_manager](access_manager): Contains the access-manager pod used to persist SSH users based on a K8s ConfigMap
* [cli](cli): The CLI is used to manage a Constellation cluster
* [bootstrapper](bootstrapper): The bootstrapper is a node agent whose most important task is to bootstrap a node
* [image](image): Build files for the Constellation disk image
* [kms](kms): Constellation's key management client and server
* [mount](mount): Package used by CSI plugins to create and mount encrypted block devices
* [state](state): Contains the disk-mapper that maps the encrypted node data disk during boot

Development components:

* [3rdparty](3rdparty): Contains the third party dependencies used by Constellation
* [conformance](conformance): Kubernetes conformance tests
* [debugd](debugd): Debug daemon and client
* [hack](hack): Development tools
* [proto](proto): Proto files generator
* [terraform](terraform): Infrastructure management using terraform (instead of `constellation create/destroy`)
  * [libvirt](terraform/libvirt): Deploy local cluster using terraform, libvirt and QEMU
* [test](test): Integration test

Additional repositories:

* [constellation-docs](https://github.com/edgelesssys/constellation-docs): End-user documentation
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

### Debug Images

> :warning: These images are not safe to use in production environments. :warning:

As described in [debugd](/debugd/README.md), it is possible to use a CoreOS image targeted at dev environments. This image allows to upload any [bootstrapper](/bootstrapper/README.md) using [cdbg](/debugd/cdbg).

To enable the upload, an additional **unsecured** port (4000) is opened which accepts any binary to be run on target machine. **Make sure that this machine is not exposed to the internet.**

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

## Linting

This projects uses [golangci-lint](https://golangci-lint.run/) for linting.
You can [install golangci-lint](https://golangci-lint.run/usage/install/#linux-and-windows) locally,
but there is also a CI action to ensure compliance.

To locally run all configured linters, execute

```sh
golangci-lint run ./...
```

It is also recommended to use golangci-lint (and [gofumpt](https://github.com/mvdan/gofumpt) as formatter) in your IDE, by adding the recommended VS Code Settings or by [configuring it yourself](https://golangci-lint.run/usage/integrations/#editor-integration)

## Nested Go modules

As this project contains nested Go modules, it is recommended to create a local Go workspace, so your IDE can lint multiple modules at once.

```go
go 1.18

use (
 .
 ./hack
 ./operators/constellation-node-operator
)
```

You can find an introduction in the [Go workspace tutorial](https://go.dev/doc/tutorial/workspaces).

If you have changed dependencies within a module and have run `go mod tidy`, you can use `go work sync` to sync versions of the same dependency of the different modules.

## Recommended VS Code Settings

The following can be added to your personal `settings.json`, but it is recommended to add it to
the `<REPOSITORY>/.vscode/settings.json` repo, so the settings will only affect this repository.

```jsonc
    // Use gofumpt as formatter.
    "gopls": {
      "formatting.gofumpt": true,
    },
    // Use golangci-lint as linter. Make sure you've installed it.
    "go.lintTool":"golangci-lint",
    "go.lintFlags": ["--fast"],
    // You can easily show Go test coverage by running a package test.
    "go.coverageOptions": "showUncoveredCodeOnly",
    // Executing unit tests with race detection.
    // You can add preferences like "-v" or "-count=1"
    "go.testFlags": ["-race"],
    // Enable language features for files with build tags.
    // Attention! This leads to integration test being executed when
    // running a package test within a package containing integration
    // tests.
    "go.buildTags": "integration",
```

## Naming convention

### Network

IP addresses:

* ip: numeric IP address
* host: either IP address or hostname
* endpoint: host+port

### Keys

* key: symmetric key
* pubKey: public key
* privKey: private key
