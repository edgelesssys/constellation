# constellation-coordinator

## Prerequisites
* Go 1.18

### Ubuntu 20.04
```sh
sudo apt install build-essential cmake libssl-dev pkg-config libcryptsetup12 libcryptsetup-dev
```

## Build
```sh
mkdir build
cd build
cmake ..
make -j`nproc`
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
[Follow this introduction on how to install and setup `cdbg`](#debug-daemon-debugd)

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
