# Creating a dev-cluster

After building (see [here](./build-develop-deploy.md#build)), you can find all CLIs and binaries in the `build` directory.

The cluster creation mostly follows the [official docs instructions](https://docs.edgeless.systems/constellation/getting-started/first-steps), but varies slightly in the following steps:

## Config

### Debug cluster

`./constellation config generate <CSP>`
by default uses the referenced nightly image.
To replace them with the latest debug image, run

```sh
bazel run //internal/api/versionsapi/cli -- latest --ref main --stream debug
```

to fetch the latest version and insert in the `image`` field of the config file.

Set `debug: true` in the config.

Create the cluster and deploy the debug images:
`./constellation create ...`

`./cdbg deploy`

Finally run:

`./constellation init`

### OSS version?
<!-- not sure -->
```sh
./constellation config fetch-measurements --insecure`
```
