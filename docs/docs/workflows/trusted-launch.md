# Azure trusted launch VMs

Constellation supports Azure trusted launch VMs. These are VMs with instance type `Standard_D*_v4` and `Standard_E*_v4`.

:::caution

Trusted launch VMs don't provide [runtime encryption](../overview/confidential-kubernetes.md).
For highest security, use Confidential VMs.

:::

Run `constellation config instance-types` to show all supported instance types.

## VM images

Azure currently doesn't support [community galleries for trusted launch VMs](https://docs.microsoft.com/en-us/azure/virtual-machines/share-gallery-community). So you need to import the VM image into your cloud subscription.

The latest image is available at [https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/azure_image_exports/2.0.0](https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/azure_image_exports/2.0.0). Simply adjust the last three numbers if you want to download an image for a different version.

After you've downloaded the image, create a resource group `constellation-images` in your Azure subscription and import the image.
You can use a script to do this:
```bash
wget https://github.com/edgelesssys/constellation/blob/main/hack/importAzure.sh
chmod +x importAzure.sh
AZURE_IMAGE_VERSION=2.0.0 AZURE_RESOURCE_GROUP_NAME=constellation-images AZURE_IMAGE_FILE=./2.0.0 ./importAzure.sh
```

The script creates the following resources:
1. A new image gallery with the default name `constellation-import`
2. A new image definition with the default name `constellation`
3. The actual image with the provided version. In this case `2.0.0`

Once the import is completed, use the `ID` of the image version in your `constellation-conf.yaml` for the `image` field. Set `confidentialVM` to `false`.

:::info

The [constellation create](create.md) command will issue a warning because manually imported images aren't recognized as production grade images:

```shell-session
Configured image doesn't look like a released production image. Double check image before deploying to production.
```

Please ignore this warning.

:::
