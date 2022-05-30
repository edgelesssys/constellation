# Local image testing with QEMU / libvirt / terraform

## Usage

Prerequisite:

- [qcow2 constellation image](/image/)
- [setup](#setup-libvirt--terraform)

Optional: Write a `terraform.tfvars` file in the terraform workspace (`terraform/libvirt`), defining required variables and overriding optional variables.
See [variables.tf](./variables.tf) for a description of all available variables.
```tfvars
constellation_coreos_image_qcow2="/path/to/image.qcow2"
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
