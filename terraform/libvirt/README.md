# Auotmated local image testing with QEMU / libvirt / terraform

## Usage

Prerequisite:

- [qcow2 constellation image](/image/)
- [setup](#setup-libvirt--terraform)
- [qemu-metadata-api](/hack/qemu-metadata-api/README.md)

Optional: Write a `terraform.tfvars` file in the terraform workspace (`terraform/libvirt`), defining required variables and overriding optional variables.
See [variables.tf](./variables.tf) for a description of all available variables.
```tfvars
constellation_coreos_image="/path/to/image.qcow2"
# optional other vars, uncomment and change as needed
# control_plane_count=3
# worker_count=2
# vcpus=2
# memory=2048
# state_disk_size=10
# ip_range_start=100
# machine="q35"
```

Create terraform resources from within terraform workspace (`terraform/libvirt`):
```shell-session
cd terraform/libvirt
terraform init
terraform plan
terraform apply

# set CONST_DIR to your constellation workspace
export TF_DIR=$(pwd)
export CONST_DIR=$(pwd)
go run ../../hack/terraform-to-state/create-state.go  "${TF_DIR}" "${CONST_DIR}"

# use constellation (everything after constellation create)
constellation config generate qemu
# run cdbg if using a debug image
cdbg deploy
constellation init

# cleanup
rm constellation-state.json constellation-mastersecret.base64 constellation-admin.conf wg0.conf
terraform destroy
```

## Setup libvirt & Terraform

<details>
<summary>Ubuntu</summary>

[General reference](https://ubuntu.com/server/docs/virtualization-libvirt)
```shell-session
# Install Terraform
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get update && sudo apt-get install terraform
# install libvirt, KVM and tools
sudo apt install qemu-kvm libvirt-daemon-system xsltproc
sudo systemctl enable libvirtd
sudo usermod -a -G libvirt $USER
# reboot
```
</details>

<details>
<summary>Fedora</summary>

```shell-session
sudo dnf install -y dnf-plugins-core
sudo dnf config-manager --add-repo https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
sudo dnf -y install terraform qemu-kvm libvirt-daemon-config-network libvirt-daemon-kvm xsltproc
sudo usermod -a -G libvirt $USER
# reboot
```
</details>

## Change libvirt settings (on Ubuntu)
Open `/etc/libvirt/qemu.conf` and change the following settings:

```
security_driver = "none"
```
Then restart libvirt

```shell-session
sudo systemctl restart libvirtd
```

## Setup emulated TPM (on Ubuntu)
Only works if swtpm is version 0.7 or newer!
Ubuntu currently ships swtpm 0.6.3 so you need to install swtpm [from launchpad](https://launchpad.net/~stefanberger/+archive/ubuntu/swtpm-jammy/).

1. Uninstall current version of swtpm (if installed)
	```
	sudo apt remove swtpm swtpm-tools
	```
2. Add ppa (this command shows the ppa for Ubuntu 22.04 jammy but others are available)
	```
	sudo add-apt-repository ppa:stefanberger/swtpm-jammy
	sudo apt update
	```
3. Install swtpm
	```
	sudo apt install swtpm swtpm-tools
	```
4. Patch configuration under `/etc/swtpm_setup.conf`
	```
	# Program invoked for creating certificates
	create_certs_tool = /usr/bin/swtpm_localca
	```
5. Patch ownership of `/var/lib/swtpm-localca`
   ```shell-session
   sudo chown -R swtpm:root /var/lib/swtpm-localca
   ```

## Misc

- List all domains: `virsh list --all`
- Destroy domain with nvram: `virsh undefine --nvram <name>`

# Manual local image testing with QEMU

> Note: This document describes a manual method of deploying VMs with QEMU look at [terraform/libvirt](/terraform/libvirt) for an automated alternative.

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
