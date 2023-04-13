# Bazel build container

This container enables running Bazel inside a container, with the host cache mounted.

To use the container, run

```shell
source container.sh
startBazelServer
```

You can then execute Bazel commands like you normally would do, as the sourced `bazel`
function shadows binaries you might have in your path:

```shell
bazel query //...
```

To terminate the container, which is running as daemon in the background, execute

```shell
stopBazelServer
```
