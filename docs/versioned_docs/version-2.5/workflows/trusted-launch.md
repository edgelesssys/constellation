# Use Azure trusted launch VMs

Constellation also supports [trusted launch VMs](https://docs.microsoft.com/en-us/azure/virtual-machines/trusted-launch) on Microsoft Azure. Trusted launch VMs don't offer the same level of security as Confidential VMs, but are available in more regions and in larger quantities. The main difference between trusted launch VMs and normal VMs is that the former offer vTPM-based remote attestation. When used with trusted launch VMs, Constellation relies on vTPM-based remote attestation to verify nodes.

:::caution

Trusted launch VMs don't provide runtime encryption and don't keep the cloud service provider (CSP) out of your trusted computing base.

:::

Constellation supports trusted launch VMs with instance types `Standard_D*_v4` and `Standard_E*_v4`. Run `constellation config instance-types` for a list of all supported instance types.

## VM images

Azure currently doesn't support [community galleries for trusted launch VMs](https://docs.microsoft.com/en-us/azure/virtual-machines/share-gallery-community). Thus, you need to manually import the Constellation node image into your cloud subscription.

The latest image is available at `https://cdn.confidential.cloud/constellation/images/azure/trusted-launch/v2.2.0/constellation.img`. Simply adjust the version number to download a newer version.

After you've downloaded the image, create a resource group `constellation-images` in your Azure subscription and import the image.
You can use a script to do this:

```bash
wget https://raw.githubusercontent.com/edgelesssys/constellation/main/hack/importAzure.sh
chmod +x importAzure.sh
AZURE_IMAGE_VERSION=2.2.0 AZURE_RESOURCE_GROUP_NAME=constellation-images AZURE_IMAGE_FILE=./constellation.img ./importAzure.sh
```

The script creates the following resources:

1. A new image gallery with the default name `constellation-import`
2. A new image definition with the default name `constellation`
3. The actual image with the provided version. In this case `2.2.0`

Once the import is completed, use the `ID` of the image version in your `constellation-conf.yaml` for the `image` field. Set `confidentialVM` to `false`.

Fetch the image measurements:

```bash
IMAGE_VERSION=2.2.0
URL=https://public-edgeless-constellation.s3.us-east-2.amazonaws.com//communitygalleries/constellationcvm-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/$IMAGE_VERSION/measurements.yaml
constellation config fetch-measurements -u$URL -s$URL.sig
```

:::info

The [constellation create](create.md) command will issue a warning because manually imported images aren't recognized as production grade images:

```shell-session
Configured image doesn't look like a released production image. Double check image before deploying to production.
```

Please ignore this warning.

:::
