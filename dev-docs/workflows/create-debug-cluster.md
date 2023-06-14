# Creating a Debug cluster

A debug cluster allows quicker iteration cycles during development by being able to upload new bootstrapper binaries through the `cdbg` tool.

After building (see [here](./build-develop-deploy.md#build)), you can find all CLIs and binaries in the `build` directory.

The cluster creation mostly follows the [official docs instructions](https://docs.edgeless.systems/constellation/getting-started/first-steps), but varies slightly in the following steps:

`./constellation config generate <CSP>`
by default uses the referenced nightly image.
To replace them with the latest debug image, run

```sh
bazel run //internal/api/versionsapi/cli -- latest --ref main --stream debug
```

to fetch the latest version and insert in the `image` field of the config file.

Before cluster creation you need to configure the cluster as debug.
Set `debugCluster: true` in the config:

```sh
yq eval -i '.debugCluster=true' constellation-conf.yaml
```

Create the cluster and deploy the debug images:

```sh
./constellation create ...
```

```sh
./cdbg deploy
```

Finally run:

```sh
./constellation init
```
