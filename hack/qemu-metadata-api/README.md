# QEMU metadata API

This program provides a metadata API for Constellation on QEMU.

## Dependencies

To interact with QEMU `libvirt` is required.
Install the C libraries:

On Ubuntu:

```shell
sudo apt install libvirt-dev
```

On Fedora:

```shell
sudo dnf install libvirt-devel
```

## Firewalld

If your system uses `firewalld` virtmanager will add itself to the firewall rules managed by `firewalld`.
Your VMs might be unable to communicate with the host.

To fix this open port `8080` (the default port for the QEMU metadata API) for the `libvirt` zone:

```shell
# Open the port
sudo firewall-cmd --zone libvirt --add-port 8080/tcp --permanent
```

## Docker image

Build the image:

```shell
bazel build //hack/qemu-metadata-api:qemumetadata
bazel build //bazel/release:qemumetadata_sum
bazel build //bazel/release:qemumetadata_tar
bazel run //bazel/release:qemumetadata_push
```

A container of the image is automatically started by Terraform.
You can also run the image manually using the following command:

```shell
docker run -it --rm \
    --network host \
    -v /var/run/libvirt/libvirt-sock:/var/run/libvirt/libvirt-sock \
    ghcr.io/edgelesssys/constellation/qemu-metadata-api:latest
```
