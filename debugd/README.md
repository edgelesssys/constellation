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
     coordinatorPath: "./coordinator"
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
   # Set timestamp from cloud provider image name
   export TIMESTAMP=01234

   yq -i \
       "(.provider | select(. | has(\"azure\")).azure.image) = \"/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos-debugd/versions/0.0.${TIMESTAMP}\"" \
       constellation-conf.yaml

   yq -i \
       "(.provider | select(. | has(\"gcp\")).gcp.image) = \"projects/constellation-images/global/images/constellation-coreos-debugd-${TIMESTAMP}\"" \
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
gcloud compute images list --filter="name~'constellation-coreos-debugd.+'" --sort-by=~creationTimestamp --project constellation-images
```

Choose the newest debugd image with the naming scheme `constellation-coreos-debugd-<timestamp>`.

### debugd Azure Image

For Azure, run the following command to get a list of all constellation debugd images, sorted by their creation date:

```shell
az sig image-version list --resource-group constellation-images --gallery-name Constellation --gallery-image-definition constellation-coreos-debugd --query "sort_by([], &publishingProfile.publishedDate)[].id" -o table
```

Choose the newest debugd image and copy the full URI.
