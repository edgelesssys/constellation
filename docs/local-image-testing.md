# Local image testing with QEMU

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
