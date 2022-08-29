# debug daemon (debugd)

## Build cdbg

```shell
mkdir -p build
cmake ..
make cdbg
```

## debugd & cdbg usage

With `cdbg` and `yq` installed in your path:

0. Write the configuration file for cdbg `cdbg-conf.yaml`:

   ```yaml
   cdbg:
     authorizedKeys:
       - username: my-username
         publicKey: ssh-rsa AAAAB…LJuM=
     bootstrapperPath: "./bootstrapper"
     systemdUnits:
       - name: some-custom.service
         contents: |-
           [Unit]
           Description=…
   ```

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
       ".ingressFirewall += {
           \"name\": \"debugd\",
           \"description\": \"debugd default port\",
           \"protocol\": \"tcp\",
           \"iprange\": \"0.0.0.0/0\",
           \"fromport\": 4000,
           \"toport\": 0
       }" \
       constellation-conf.yaml
   ```

4. Run `constellation create […]`

5. Run `./cdbg deploy`

6. Run `constellation init […]` as usual

### debugd images

For a full list of image naming conventions and how to retreive them check [image version documentation](/.github/docs/README.md#image-versions)
