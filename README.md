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
gcloud compute images list --filter="name~'constellation-.+'" --sort-by=~creationTimestamp --project constellation-images
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
    "gcpConfig": {
      "image": "projects/constellation-images/global/images/constellation-coreos-debugd-TIMESTAMP",
      "firewallInput": {
        "Ingress": [
          {
            "Name": "coordinator",
            "Description": "Coordinator default port",
            "Protocol": "tcp",
            "FromPort": 9000
          },
          {
            "Name": "wireguard",
            "Description": "WireGuard default port",
            "Protocol": "udp",
            "FromPort": 51820
          },
          {
            "Name": "ssh",
            "Description": "SSH",
            "Protocol": "tcp",
            "FromPort": 22
          },
          {
            "Name": "nodeport",
            "Description": "NodePort",
            "Protocol": "tcp",
            "FromPort": 30000,
            "ToPort": 32767
          },
          {
            "Name": "debugd",
            "Description": "debugd default port",
            "Protocol": "tcp",
            "FromPort": 4000
          }
        ]
      }
    },
    "azureConfig": {
      "image": "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos-debugd/versions/0.0.TIMESTAMP",
      "networkSecurityGroupInput": {
        "Ingress": [
          {
            "Name": "coordinator",
            "Description": "Coordinator default port",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "FromPort": 9000
          },
          {
            "Name": "wireguard",
            "Description": "WireGuard default port",
            "Protocol": "udp",
            "IPRange": "0.0.0.0/0",
            "FromPort": 51820
          },
          {
            "Name": "ssh",
            "Description": "SSH",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "FromPort": 22
          },
          {
            "Name": "nodeport",
            "Description": "NodePort",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "FromPort": 30000,
            "ToPort": 32767
          },
          {
            "Name": "debugd",
            "Description": "debugd default port",
            "Protocol": "tcp",
            "IPRange": "0.0.0.0/0",
            "FromPort": 4000
          }
        ]
      }
    }
  }
}
```

# Local image testing with QEMU

To build our images we use the [CoreOS-Assembler (COSA)](https://github.com/edgelesssys/constellation-coreos-assembler).
COSA comes with support to test images locally. After building your image with `make coreos` you can run the image with `make run`.

Our fork adds extra utility by providing scripts to run an image in QEMU with a vTPM attached, or boot multiple VMs to simulate your own local Constellation cluster.

Begin by starting a COSA docker container
```shell
docker run -it --rm \
    --entrypoint bash \
    --device /dev/kvm \
    --device /dev/net/tun \
    --privileged \
    -v </path/to/constellation-image.qcow2>:/constellation-image.qcow2 \
    ghcr.io/edgelesssys/constellation-coreos-assembler
```

## Run a single image

Using the `run-image` script we can launch a single VM with an attached vTPM.
The script expects an image and a name to run. Optionally one may also provide the path to an existing state disk, if none provided a new disk will be created.

Additionally one may configure QEMU CPU (qemu -smp flag, default=2) and memory (qemu -m flag, default=2G) settings, as well as the size of the created state disk in GB (default 2) using environment variables.

To customize CPU settings use `CONSTELL_CPU=[[cpus=]n][,maxcpus=maxcpus][,sockets=sockets][,dies=dies][,cores=cores][,threads=threads]` \
To customize memory settings use `CONSTELL_MEM=[size=]megs[,slots=n,maxmem=size]` \
To customize state disk size use `CONSTELL_STATE_SIZE=n`

Use the following command to boot a VM with 2 CPUs, 2G RAM, a 4GB state disk with the image in `/constellation/coreos.qcow2`.
Logs and state files will be written to `/tmp/test-vm-01`.
```shell
sudo CONSTELL_CPU=2 CONSTELL_MEM=2G CONSTELL_STATE_SIZE=4 run-image /constellation/coreos.qcow2 test-vm-01
```

The command will create a network bridge and add the VM to the bridge, so the host may communicate with the guest VM, as well as allowing the VM to access the internet.

Press <kbd>Ctrl</kbd>+<kbd>A</kbd> <kbd>X</kbd> to stop the VM, this will remove the VM from the bridge but will keep the bridge alive.

Run the following to remove the bridge.
```shell
sudo delete_network_bridge br-constell-0
```

## Create a local cluster

Using the `create-constellation` script we can create multiple VMs using the same image and connected in one network.

The same environment variables as for `run-image` can be used to configure cpu, memory, and state disk size.

Use the following command to create a cluster of 4 VMs, where each VM has 3 CPUs, 4GB RAM and a 5GB state disk.
Logs and state files will be written to `/tmp/constellation`.
```shell
sudo CONSTELL_CPU=3 CONSTELL_MEM=4G CONSTELL_STATE_SIZE=5 create-constellation 4 /constellation/coreos.qcow2
```

The command will use the `run-image` script launch each VM in its own `tmux` session.
View the VMs by running the following
```shell
sudo tmux attach -t constellation-vm-<i>
```

# Development Guides

- [Upgrading Kubernetes](/docs/upgrade-kubernetes.md)

# Deployment Guides

- [Onboarding Customers](/docs/onboarding-customers.md)
