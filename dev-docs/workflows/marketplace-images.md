# Using Marketplace Images in Constellation

This document explains the steps a user needs to take to run Constellation with dynamic billing via the cloud marketplaces.

## AWS

Marketplace Images on AWS are not available yet.

## Azure

On Azure, to use a marketplace image, ensure that the subscription has accepted the agreement to use marketplace images:

```bash
az vm image terms accept --publisher edgelesssystems --offer constellation --plan constellation
```

Then, set the VMs to use the marketplace image in the `constellation-conf.yaml` file:

```bash
yq eval -i ".provider.azure.useMarketplaceImage = true" constellation-conf.yaml
```

And ensure that the cluster uses a release image (i.e. `.image=vX.Y.Z` in the `constellation-conf.yaml` file). Afterwards, proceed with the cluster creation as usual.

## GCP

Marketplace Images on GCP are not available yet.
