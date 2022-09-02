# Azure trusted lanuch VMs

Constellation supports Azure trusted launch VMs. These are VMs with instance type `Standard_D*_v4` and `Standard_E*_v4`.

To find out which instance types are supported by your constellation version use `constellation create --help`.

## Virtual machine images

Azure currently does not support [community galleries for trusted launch VMs](https://docs.microsoft.com/en-us/azure/virtual-machines/share-gallery-community). Therefore each user needs to import our virtual machine image into their cloud subscription.

Each released Constellation VM image is made publicly available via [AWS S3](https://aws.amazon.com/s3/). For example, to import an image for Constellation v2.0.0, download the corresponding image from [https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/azure_image_exports/2.0.0](https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/azure_image_exports/2.0.0). Simply adjust the last three numbers, if you want to download an image for a different version.

Afterwards you can use our [importAzure.sh](https://github.com/edgelesssys/constellation/blob/main/hack/importAzure.sh) script to make this image available in your Azure account:

```bash
AZURE_IMAGE_VERSION=2.0.0 AZURE_RESOURCE_GROUP_NAME=constellation-images AZURE_IMAGE_FILE=./2.0.0 ./importAzure.sh
```

Please make sure that the specified resource group exists, before executing this script.

The script will create the following resources:
1. A new image gallery with the default name `constellation-import`
2. A new image definition with the default name `constellation`
3. The actual image with the provided version. In this case `2.0.0`

Once the import is completed you can use the `ID` of the image version in your `constellation-conf.yaml` for the `image` field, and set `confidentialVM` to `false`.

:::info

Please note that [create.md](create) will issue a warning, because it is not able to recognize manually imported images as production grade images:

```shell-session
Configured image does not look like a released production image. Double check image before deploying to production.
```

This warning can safely be ignored!

:::
