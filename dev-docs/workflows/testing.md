# Test

## Unit tests

Running unit tests with Bazel:

```sh
bazel test //...
```

## Integration tests

You can run all integration like this:

```sh
ctest -j `nproc`
```

You can limit the execution of tests to specific targets with e.g. `ctest -R integration-node-operator`.

Some of the tests rely on libvirt and won't work if you don't have a virtualization capable CPU. You can find instructions on setting up libvirt in our [QEMU README](qemu.md).
