# Using Constellation via Cloud Marketplaces

Constellation is available through the Marketplaces of AWS, Azure, GCP, and STACKIT. This allows you to create self-managed Constellation clusters that are billed on a pay-per-use basis (hourly, per vCPU) with your CSP account. You can still get direct support by Edgeless Systems. For more information, please [contact us](https://www.edgeless.systems/enterprise-support/).

This document explains how to run Constellation with the dynamically billed cloud marketplace images.

<Tabs groupId="csp">
<TabItem value="aws" label="AWS">

To use Constellation's marketplace images, ensure that you are subscribed to the [marketplace offering](https://aws.amazon.com/marketplace/pp/prodview-2mbn65nv57oys) through the web portal.

Then, enable the use of marketplace images in your Constellation `constellation-conf.yaml` [config file](../workflows/config.md):

```bash
yq eval -i ".provider.aws.useMarketplaceImage = true" constellation-conf.yaml
```

</TabItem>
<TabItem value="azure" label="Azure">

Constellation has a private marketplace plan. Please [contact us](https://www.edgeless.systems/enterprise-support/) to gain access.

To use a marketplace image, you need to accept the marketplace image's terms once for your subscription with the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/vm/image/terms?view=azure-cli-latest):

```bash
az vm image terms accept --publisher edgelesssystems --offer constellation --plan constellation
```

Then, enable the use of marketplace images in your Constellation `constellation-conf.yaml` [config file](../workflows/config.md):

```bash
yq eval -i ".provider.azure.useMarketplaceImage = true" constellation-conf.yaml
```

</TabItem>
<TabItem value="gcp" label="GCP">

To use a marketplace image, ensure that the account is entitled to use marketplace images by Edgeless Systems by accepting the terms through the [web portal](https://console.cloud.google.com/marketplace/vm/config/edgeless-systems-public/constellation).

Then, enable the use of marketplace images in your Constellation `constellation-conf.yaml` [config file](../workflows/config.md):

```bash
yq eval -i ".provider.gcp.useMarketplaceImage = true" constellation-conf.yaml
```

</TabItem>
<TabItem value="stackit" label="STACKIT">

On STACKIT, the selected Constellation image is always a marketplace image. You can find more information on the STACKIT portal.

</TabItem>
</Tabs>

Ensure that the cluster uses an official release image version (i.e., `.image=vX.Y.Z` in the `constellation-conf.yaml` file).

From there, you can proceed with the [cluster creation](../workflows/create.md) as usual.
