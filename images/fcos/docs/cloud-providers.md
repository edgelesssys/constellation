# Cloud Providers

Custom CoreOS images created here can be uploaded to supported cloud providers. This documents contains information on how to manually spawn cloud provider instances using custom CoreOS images.

## GCP

```shell
gcloud compute instances create <INSTANCE_NAME> --zone=<ZONE>  --machine-type=<MACHINE_TYPE> --image <IMAGE_NAME>  --maintenance-policy=TERMINATE --confidential-compute --shielded-secure-boot --shielded-vtpm --shielded-integrity-monitoring --scopes=https://www.googleapis.com/auth/cloud-platform,https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management,https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/trace.append
```

## Azure

Non-CVM:
```
az image list
# copy image id from output of previous command
az vm create --resource-group <RESOURCE_GROUP> --location <LOCATION> --name <INSTANCE_NAME> --os-type linux --public-ip-sku Standard --image <IMAGE_ID>
```

### Create Marketplace offer

- Upload a vhd and image to azure portal using the Makefile
- Create (or reuse) a `shared image gallery`:
  - Create image gallery if it does not exist yet
    - Search for "Azure compute galleries" in azure portal
    - Click "create"
    - Choose "constellation-images" resource group and pick a name, then click create
- Create a VM image definition
  - Search for "Azure compute galleries" in azure portal and choose the created gallery
  - Click "Create a VM image definition"
  - OS type: Linux
  - OS state: Generalized
  - VM generation: Gen 2
  - Publisher: EdgelessSystems
  - Offer: constellation-coreos
  - SKU: constellation-coreos
  - Source image: Choose image uploaded using Makefile
  - Create
- Create Marketplace offer (on https://partner.microsoft.com/)
  - Navigate to marketplace offers overview (https://partner.microsoft.com/en-us/dashboard/marketplace-offers/overview)
  - If you want to create a new version of an existing plan, skip this section
  - Click "New offer" -> "Azure Virtual Machine"
  - Choose an offer id and alias
  - Create a new plan on "Plan overview" -> "Create new plan", choose a plan id and plan name
  - In "Technical configuration", create a generation, choose "Azure shared image gallery" and select the image created earlier
