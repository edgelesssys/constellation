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

2. Locate the latest debugd images for [GCP](#debugd-gcp-image) and [Azure](#debugd-azure-image)

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

### debugd GCP image

For GCP, run the following command to get a list of all constellation debug images, sorted by their creation date:

```shell
gcloud compute images list --filter="family~'constellation-debug-v.+'" --sort-by=creationTimestamp --project constellation-images --uri | sed 's#https://www.googleapis.com/compute/v1/##'
```

The images are grouped by the Constellation release they were built for.
Choose the newest debugd image for your release and copy the full URI.

### debugd Azure Image

Azure debug images are grouped by the Constellation release they were built for.
To get a list of available releases, run the following:

```shell
az sig image-definition list --resource-group constellation-images --gallery-name Constellation_Debug --query "[].name"  -o table
```

Run the following command to get a list of all constellation debugd images for release v1.5.0, sorted by their creation date:

```shell
RELEASE=v1.5.0
az sig image-version list --resource-group constellation-images --gallery-name Constellation_Debug --gallery-image-definition ${RELEASE} --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table
```

Choose the newest debugd image and copy the full URI.
