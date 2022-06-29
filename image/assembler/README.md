# Constellation-OS Assembler

This container image uses [coreos assembler](https://coreos.github.io/coreos-assembler/) as a base (`quay.io/coreos-assembler/coreos-assembler`) to facilitate a build environment for the Constellation-OS.
The root filesystem takes its contents (OSTree) from [constellation-fedora-coreos-config](https://github.com/edgelesssys/constellation-fedora-coreos-config).
The constellation specific changes are tracked in this repository: https://github.com/edgelesssys/constellation-coreos-assembler
And the Constellation-OS Assembler can be pulled from ghcr.io/edgelesssys/constellation-coreos-assembler

## Setup

Prerequisites: `podman` and `qemu-kvm` are installed, nested virtualization is enabled.
Make sure your user is allowed read and write access on `/dev/kvm`.
If the device is not mounted in the container try the following command, and restart the container:
``` shell
sudo chmod 666 /dev/kvm
```

## Using the Assembler to create a bootable operating system

1. Create the assembler image as described [here](#creating-the-assembler-image)
2. Source the `fcos/.env` file to enable the `cosa` bash alias:
   ```
   source fcos/.env
   ```
3. Set the `BOOTSTRAPPER_BINARY` environment variable to a path of the compiled bootstrapper binary. It will be mounted in the cosa container and copied into the resulting coreos image.
   ```
   BOOTSTRAPPER_BINARY="/path/to/bootstrapper"
   ```
4. Go into the build folder and initialize cosa:
   ```
   cd fcos/build
   cosa init https://github.com/edgelesssys/constellation-fedora-coreos-config
   cosa fetch
   ```
5. Build the OS image:
   ```
   cosa build
   ```
6. Create an image for a cloud provider
    ```
    cosa buildextend-gcp
    cosa buildextend-aws
    cosa buildextend-azure
    [...]
    ```

## Using a locally checked out git repo of the coreos-config during development

Simply set the environment variable `COREOS_ASSEMBLER_CONFIG_GIT` to the local folder and perform the rest of the steps as usual:
```
COREOS_ASSEMBLER_CONFIG_GIT=/path/to/constellation-fedora-coreos-config
```
