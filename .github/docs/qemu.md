# Local image testing with QEMU / libvirt

To create local testing clusters using QEMU, some prerequisites have to be met:

- [qcow2 constellation image](/image/README.md)
- [qemu-metadata-api container image](/hack/qemu-metadata-api/README.md)

Deploying the VMs requires `libvirt` to be installed and configured correctly.
You may either use [your local libvirt setup](#local-libvirt-setup) if it meets the requirements, or use a [containerized libvirt in docker](#containerized-libvirt).

## Containerized libvirt

Constellation will automatically deploy a containerized libvirt instance, if no connection URI is defined in the Constellation config file.
Follow the steps in our [libvirt readme](../../cli/internal/libvirt/README.md) if you wish to build your own image.

## Local libvirt setup

<details>
<summary>Ubuntu</summary>

### Install required packages

[General reference](https://ubuntu.com/server/docs/virtualization-libvirt)

```shell-session
sudo apt install qemu-kvm libvirt-daemon-system xsltproc
sudo systemctl enable libvirtd
sudo usermod -a -G libvirt $USER
# reboot
```

### Setup emulated TPM

Using a virtual TPM (vTPM) with QEMU only works if swtpm is version 0.7 or newer!
Ubuntu 22.04 currently ships swtpm 0.6.3, so you need to install swtpm [from launchpad](https://launchpad.net/~stefanberger/+archive/ubuntu/swtpm-jammy/).

1. Uninstall current version of swtpm (if installed)

    ```shell-session
    sudo apt remove swtpm swtpm-tools
    ```

2. Add ppa (this command shows the ppa for Ubuntu 22.04 jammy but others are available)

    ```shell-session
    sudo add-apt-repository ppa:stefanberger/swtpm-jammy
    sudo apt update
    ```

3. Install swtpm

    ```shell-session
    sudo apt install swtpm swtpm-tools
    ```

4. Patch configuration under `/etc/swtpm_setup.conf`

    ```shell-session
    # Program invoked for creating certificates
    create_certs_tool = /usr/bin/swtpm_localca
    ```

5. Patch ownership of `/var/lib/swtpm-localca`

   ```shell-session
   sudo chown -R swtpm:root /var/lib/swtpm-localca
   ```

</details>

<details>
<summary>Fedora</summary>

```shell-session
sudo dnf install -y dnf-plugins-core
sudo dnf -y install qemu-kvm libvirt-daemon-config-network libvirt-daemon-kvm xsltproc swtpm
sudo usermod -a -G libvirt $USER
# reboot
```

</details>

### Update libvirt settings

Open `/etc/libvirt/qemu.conf` and change the following settings:

```shell-session
security_driver = "none"
```

Then restart libvirt

```shell-session
sudo systemctl restart libvirtd
```

## Troubleshooting

### VMs are not properly cleaned up after a failed `constellation create` command

Terraform may fail to remove your VMs, in which case you need to do so manually.

- List all domains: `virsh list --all`
- Destroy domains with nvram: `virsh undefine --nvram <name>`

### VMs have no internet access

`iptables` rules may prevent your VMs form properly accessing the internet.
Make sure your rules are not dropping forwarded packages.

List and check your rules:

```shell
$ sudo iptables -S
-P INPUT ACCEPT
-P FORWARD DROP
-P OUTPUT ACCEPT
-N DOCKER
...
```

If your `FORWARD` chain is set to `DROP`, you will need to update your rules:

```shell
sudo iptables -P FORWARD ACCEPT
```
