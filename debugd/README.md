# debug daemon (debugd)

Debugd is a tool we built to allow for shorter iteration cycles during development.
The debugd gets embedded into OS images at the place where the bootstrapper normally sits.
Therefore, when a debug image is started, the debugd starts executing instead of the bootstrapper.
The debugd will then wait for a request from the `cdbg` tool to upload a bootstrapper binary.
Once the upload is finished debugd will start the bootstrapper.
Subsequently you can initialize your cluster with `constellation apply` as usual.

## Build cdbg

The `cdbg` tool is part of the `//:devbuild` target, if you follow the generic build instructions at [build-develop-deploy](../dev-docs/workflows/build-develop-deploy.md).

If you need to build `cdbg` standalone for your current platform, you can run

```sh
bazel build //debugd/cmd/cdbg:cdbg_host
```

## debugd & cdbg usage

Follow the [debug-cluster workflow](../dev-docs/workflows/debug-cluster.md) to deploy a bootstrapper with `cdbg` and `debugd`.

### Logcollection to Opensearch

You can enable the logcollection of debugd to send logs to Opensearch.

On Azure, ensure your user assigned identity has the `Key Vault Secrets User` role assigned on the key vault `opensearch-creds`.

On AWS, attach the `SecretManagerE2E` policy to your control-plane and worker node role.

When deploying with cdbg, enable by setting the `logcollect=true` and your name `logcollect.admin=yourname`.

```shell-session
./cdbg deploy --info logcollect=true,logcollect.admin=yourname

# OR

./cdbg deploy --info logcollect=true --info logcollect.admin=yourname
```

Other available fields can be found in [the filed list](/debugd/internal/debugd/logcollector/fields.go)

For QEMU, the credentials for Opensearch must be parsed via the info flag as well:

```shell-session
./cdbg deploy \
    --info logcollect=true \
    --info logcollect.admin=yourname \
    --info qemu.opensearch-pw='xxxxxxx'

```

Remember to use single quotes for the password.

You will also need to increase the memory size of QEMU to 4GB.
