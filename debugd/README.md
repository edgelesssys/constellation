# debug daemon (debugd)

Debugd is a tool we built to allow for shorter iteration cycles during development.
The debugd gets embedded into OS images at the place where the bootstrapper normally sits.
Therefore, when a debug image is started, the debugd starts executing instead of the bootstrapper.
The debugd will then wait for a request from the `cdbg` tool to upload a bootstrapper binary.
Once the upload is finished debugd will start the bootstrapper.
Subsequently you can initialize your cluster with `constellation init` as usual.

## Build cdbg

```shell
mkdir -p build
cmake ..
make cdbg
```

## debugd & cdbg usage
Before continuing, remeber to [set up](https://docs.edgeless.systems/constellation/getting-started/install#set-up-cloud-credentials) your cloud credentials for the CLI to work.

With `cdbg` and `yq` installed in your path:

1. Run `constellation config generate` to create a new default configuration

2. Locate the latest debugd images by running `hack/find-image/find-image.sh`

3. Modify the `constellation-conf.yaml` to use an image with the debugd already included and add required firewall rules:

   ```shell-session
   # Set full reference of cloud provider image name
   export IMAGE_URI=
   ```

   ```shell-session
   yq -i \
       "(.provider | select(. | has(\"azure\")).azure.image) = \"${IMAGE_URI}\"" \
        constellation-conf.yaml
   yq -i \
       "(.provider | select(. | has(\"gcp\")).gcp.image) = \"${IMAGE_URI}\"" \
       constellation-conf.yaml

   yq -i \
       "(.debugCluster) = true" \
       constellation-conf.yaml
   ```

4. Run `constellation create […]`

5. Run `./cdbg deploy`

   By default, `cdbg` searches for the bootstrapper in the current path (`./bootstrapper`). You can define a custom path by appending the argument `--bootstrapper <path to bootstrapper>` to `cdbg deploy`.

6. Run `constellation init […]` as usual

### Logcollection to Opensearch

You can enable the logcollection of debugd to send logs to Opensearch.

On Azure, ensure your user assigned identity has the `Key Vault Secrets User` role assigned on the key vault `opensearch-creds`.

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

### debugd images

For a full list of image naming conventions and how to retreive them check [image version documentation](/.github/docs/README.md#image-versions)
