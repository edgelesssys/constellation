# debug daemon (debugd)

## Build cdbg

```shell
mkdir -p build
cmake ..
make cdbg
```

## debugd & cdbg usage

With `cdbg` and `yq` installed in your path:

1. Run `constellation config generate` to create a new default configuration

2. Locate the latest debugd images for [GCP](/.github/docs/README.md#gcp) and [Azure](/.github/docs/README.md#azure)

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


### debugd images

For a full list of image naming conventions and how to retreive them check [image version documentation](/.github/docs/README.md#image-versions)
