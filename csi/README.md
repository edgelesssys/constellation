# Constellation CSI tools

These packages are intended to be used by [Kubernetes CSI drivers](https://kubernetes.io/blog/2019/01/15/container-storage-interface-ga/) to enable transparent encryption of storage on the node.

## Dependencies

This package uses the C library [`libcryptsetup`](https://gitlab.com/cryptsetup/cryptsetup/) for device mapping and crypto operations.

* Install on Ubuntu:

    ```bash
    sudo apt install libcryptsetup12 libcryptsetup-dev
    ```

* Install on Fedora:

    ```bash
    sudo dnf install cryptsetup-libs cryptsetup-devel
    ```

## Testing

Running the integration test requires root privileges.
Build and run the test:

``` bash
go test -c -tags=integration ./test/
sudo ./test.test
```
