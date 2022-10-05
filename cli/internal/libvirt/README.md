# Containerized libvirt

To avoid dependency issues with the libvirt setup of the host, we provide a containerized libvirt instance.
If no libvirt connection string is provided in the Constellation config file during create,
this container is deployed to provide a libvirt daemon for orchestrating Constellation nodes in QEMU.

The container will listen for libvirt connections on `localhost:16599`.
Connecting to the libvirt daemon running in the container and manage the deployment using `virsh` run the following:

```shell
virsh -c "qemu+tcp://localhost:16599/system"
```

## Docker image

Build the image:

```shell
DOCKER_BUILDKIT=1 docker build -t ghcr.io/edgelesssys/constellation/libvirt:latest -f cli/internal/libvirt/Dockerfile .
```

A container of the image is automatically started by the CLI.
You can also run the image manually using the following command:

```shell
docker run -it --rm \
    --network host \
    --privileged true \
    ghcr.io/edgelesssys/constellation/libvirt:latest
```
