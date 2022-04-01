# constellation-coordinator

## Prerequisites
* Go 1.18

### Ubuntu 20.04
```sh
sudo apt install build-essential cmake libssl-dev pkg-config libcryptsetup12 libcryptsetup-dev
curl https://sh.rustup.rs -sSf | sh
```

### Amazon Linux
```sh
sudo yum install cmake3 gcc make
curl https://sh.rustup.rs -sSf | sh
```

## Build
```sh
mkdir build
cd build
cmake ..
make -j`nproc`
```

## CMake build options:

### Release build

This options leaves out debug symbols and turns on more compiler optimizations.

```sh
cmake -DCMAKE_BUILD_TYPE=Release ..
```

### Static build (coordinator as static binary, no dependencies on libc or other libraries)

Install the musl-toolchain

Ubuntu / Debian:
```sh
sudo apt install -y musl-tools
rustup target add x86_64-unknown-linux-musl
```

From source (Amazon-Linux):
```sh
wget https://musl.libc.org/releases/musl-1.2.2.tar.gz
tar xfz musl-1.2.2.tar.gz
cd musl-1.2.2
./configure
make -j `nproc`
sudo make install
rustup target add x86_64-unknown-linux-musl
```
Add `musl-gcc` to your PATH:
```sh
export PATH=$PATH:/usr/loca/musl/bin/
```

Compile the coordinator
```sh
cmake -DCOORDINATOR_STATIC_MUSL=ON ..
```

## Cloud credentials

Using the CLI or debug-CLI requires the user to make authorized API calls to the AWS or GCP API.

### Google Cloud Platform (GCP)

If you are running from within a Google VM, and the VM is allowed to access the necessary APIs, no further configuration is needed.

Otherwise you have a couple options:

1. Use the `gcloud` CLI tool

    ```shell
    gcloud auth application-default login
    ```
    This will ask you to log into your Google account, and then create your credentials.
    The Constellation CLI will automatically load these credentials when needed.

2. Set up a service account and pass the credentials manually

    Follow [Google's guide](https://cloud.google.com/docs/authentication/production#manually) for setting up your credentials.

### Amazon Web Services (AWS)

To use the CLI with an Constellation cluster on AWS configure the following files:


```bash
$ cat ~/.aws/credentials
[default]
aws_access_key_id = XXXXX
aws_secret_access_key = XXXXX
```

```bash
$ cat ~/.aws/config
[default]
region = us-east-2
```

### Azure

To use the CLI with an Constellation cluster on Azure execute:
```bash
az login
```

### Deploying a locally compiled coordinator binary

By default, `constellation create ...` will spawn cloud provider instances with a pre-baked coordinator binary.
For testing, you can use the constellation debug daemon (debugd) to upload your local coordinator binary to running instances and to obtain SSH access.
[Follow this introduction on how to install and setup `cdbg`](#debugd-debug-daemon)

# debug daemon (debugd)

## debugd Prerequisites

* Go 1.18

## Build debugd

```
mkdir -p build
go build -o build/debugd debugd/debugd/cmd/debugd/debugd.go
```

## Build & install cdbg

The go install command for cdbg only works inside the checked out repository due to replace directives in the `go.mod` file.

```
git clone https://github.com/edgelesssys/constellation && cd constellation
go install github.com/edgelesssys/constellation/debugd/cdbg
```

## debugd & cdbg usage

With `cdbg` installed in your path:

1. Run `constellation --dev-config /path/to/dev-config create […]` while specifying a cloud-provider image with the debugd already included. See [Configuration](#debugd-configuration) for a dev-config with a custom image and firewall rules to allow incoming connection on the debugd default port 4000.
2. Run `cdbg deploy --dev-config /path/to/dev-config`
3.  Run `constellation init […]` as usual



### debugd GCP image

For GCP, run the following command to get a list of all constellation images, sorted by their creation date:
```
gcloud compute images list --filter="name~'constellation-.+'" --sort-by=~creationTimestamp
```
Choose the newest debugd image with the naming scheme `constellation-coreos-debugd-<timestamp>`.

### debugd Azure Image

For Azure, run the following command to get a list of all constellation debugd images, sorted by their creation date:
```
az sig image-version list --resource-group constellation-images --gallery-name Constellation --gallery-image-definition constellation-coreos-debugd --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table
```
Choose the newest debugd image and copy the full URI.

## debugd Configuration

You should first locate the newest debugd image for your cloud provider ([GCP](#debugd-gcp-image), [Azure](#debugd-azure-image)).

This tool uses the dev-config file from `constellation-coordinator` and extends it with more fields.
See this example on what the possible settings are and how to setup the constellation cli to use a cloud-provider image and firewall rules with support for debugd:
```json
{
   "cdbg":{
      "authorized_keys":[
         {
            "user":"my-username",
            "pubkey":"ssh-rsa AAAAB…LJuM="
         }
      ],
      "coordinator_path":"/path/to/coordinator",
      "systemd_units":[
         {
            "name":"some-custom.service",
            "contents":"[Unit]\nDescription=…"
         }
      ]
   },
   "provider": {
    "gcpconfig": {
      "image": "constellation-coreos-debugd-TIMESTAMP",
      "firewallinput": {
        "Ingress": [
          {
            "Name": "coordinator",
            "Description": "Coordinator default port",
            "Protocol": "tcp",
            "Port": 9000
          },
          {
            "Name": "wireguard",
            "Description": "WireGuard default port",
            "Protocol": "udp",
            "Port": 51820
          },
          {
            "Name": "ssh",
            "Description": "SSH",
            "Protocol": "tcp",
            "Port": 22
          },
          {
            "Name": "debugd",
            "Description": "debugd default port",
            "Protocol": "tcp",
            "Port": 4000
          }
        ]
      }
    },
    "azureconfig": {
      "image": "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos-debugd/versions/0.0.TIMESTAMP",
      "networksecuritygroupinput": {
        "Ingress": [
          {
            "Name": "coordinator",
            "Description": "Coordinator default port",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "Port": 9000
          },
          {
            "Name": "wireguard",
            "Description": "WireGuard default port",
            "Protocol": "udp",
            "IPRange": "0.0.0.0/0",
            "Port": 51820
          },
          {
            "Name": "ssh",
            "Description": "SSH",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "Port": 22
          },
          {
            "Name": "debugd",
            "Description": "debugd default port",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "Port": 4000
          }
        ]
      }
    }
  }
}
```
# constellation-kms-client

This library provides an interface for the key management services used with constellation.
It's intendet for the Constellation CSI Plugins and the CLI.

## KMS

The Cloud KMS is where we store our key encryption key (KEK).
It should be initiated by the CLI and provided with a key release policy.
The CSP Plugin can request to encrypt data encryption keys (DEK) with the DEK to safely store them on persistent memory.
The [kms](pkg/kms) package interacts with the Cloud KMS APIs.
Currently planned are KMS are:

* AWS KMS
* GCP CKM
* Azure Key Vault


## Storage

Storage is where the CSI Plugin stores the encrypted DEKs.
Currently planned are:

* AWS S3, SSP
* GCP GCS
* Azure Blob
# constellation-images
# constellation-mount-utils
Wrapper for https://github.com/kubernetes/mount-utils


## Dependencies

This package uses the C library [`libcryptsetup`](https://gitlab.com/cryptsetup/cryptsetup/) for device mapping.

To install the required dependencies on Ubuntu run:
```shell
sudo apt install libcryptsetup-dev
```

To install or upgrade `go.mod` dependencies from private repositories run:
```
GOPRIVATE=github.com/edgelesssys/constellation-coordinator go get github.com/edgelesssys/constellation-coordinator
GOPRIVATE=github.com/edgelesssys/constellation-kms-client go get github.com/edgelesssys/constellation-kms-client
```

## Testing

A small test programm is available in `test/main.go`.
To build the programm run:
```shell
go build -o test/crypt ./test/
```

Create a new crypt device for `/dev/sdX` and map it to `/dev/mapper/volume01`:
```shell
sudo test/crypt -source /dev/sdX -target volume01 -v 4
```

You can now interact with the mapped volume as if it was an unformatted device:
```shell
sudo mkfs.ext4 /dev/mapper/volume01
sudo mount /dev/mapper/volume01 /mnt/volume01
```

Close the mapped volume:
```shell
sudo umount /mnt/volume01
sudo test/crypt -c -target volume01 -v 4
```
