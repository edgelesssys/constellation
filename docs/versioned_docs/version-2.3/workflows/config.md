# Configure your cluster

Before you can create your cluster, you need to configure the identity and access management (IAM) for your cloud service provider (CSP) and choose machine types for the nodes.

## Creating the configuration file

You can generate a configuration file for your CSP by using the following CLI command:

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

```bash
constellation config generate azure
```

</TabItem>
<TabItem value="gcp" label="GCP">

```bash
constellation config generate gcp
```

</TabItem>
<TabItem value="aws" label="AWS">

```bash
constellation config generate aws
```

</TabItem>
</Tabs>

This creates the file `constellation-conf.yaml` in the current directory.

## Choosing a VM type

Constellation supports the following VM types:
<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

By default, Constellation uses `Standard_DC4as_v5` CVMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. For CVMs, any VM type with a minimum of 4 vCPUs from the [DCasv5 & DCadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/dcasv5-dcadsv5-series) or [ECasv5 & ECadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/ecasv5-ecadsv5-series) families is supported.

You can also run `constellation config instance-types` to get the list of all supported options.

</TabItem>
<TabItem value="gcp" label="GCP">

By default, Constellation uses `n2d-standard-4` VMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. Supported are all machines with a minimum of 4 vCPUs from the N2D family. Refer to [N2D machine series](https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines) or run `constellation config instance-types` to get the list of all supported options.

</TabItem>
<TabItem value="aws" label="AWS">

By default, Constellation uses `m6a.xlarge` VMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. Supported are all nitroTPM-enabled machines with a minimum of 4 vCPUs (`xlarge` or larger). Refer to the [list of nitroTPM-enabled instance types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enable-nitrotpm-prerequisites.html) or run `constellation config instance-types` to get the list of all supported options.

</TabItem>
</Tabs>

Fill the desired VM type into the **instanceType** field in the `constellation-conf.yml` file.

## Creating an IAM configuration

You can create an IAM configuration for your cluster automatically using the `constellation iam create` command.
<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

You must be authenticated with the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) in the shell session.

```bash
constellation iam create azure --region=westus --resourceGroup=constellTest --servicePrincipal=spTest
```

This command creates IAM configuration on the Azure region `westus` creating a new resource group `constellTest` and a new service principal `spTest`.

Note that CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). These are:

* `westus`
* `eastus`
* `northeurope`
* `westeurope`

Paste the output into the corresponding fields of the `constellation-conf.yaml` file.

:::tip
Since `clientSecretValue` is a sensitive value, you can leave it empty in the configuration file and pass it via an environment variable instead. To this end, create the environment variable `CONSTELL_AZURE_CLIENT_SECRET_VALUE` and set it to the secret value.
:::

</TabItem>
<TabItem value="gcp" label="GCP">

You must be authenticated with the [GCP CLI](https://cloud.google.com/sdk/gcloud) in the shell session.

```bash
constellation iam create gcp --projectID=yourproject-12345 --zone=europe-west2-a --serviceAccountID=constell-test
```

This command creates IAM configuration in the GCP project `yourproject-12345` on the GCP zone `europe-west2-a` creating a new service account `constell-test`.

Note that only regions offering CVMs of the `N2D` series are supported. You can find a [list of all regions in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available), which you can filter by machine type `N2D`.

Paste the output into the corresponding fields of the `constellation-conf.yaml` file.

</TabItem>
<TabItem value="aws" label="AWS">

You must be authenticated with the [AWS CLI](https://aws.amazon.com/en/cli/) in the shell session.

```bash
constellation iam create aws --zone=eu-central-1a --prefix=constellTest
```

This command creates IAM configuration for the AWS zone `eu-central-1a` using the prefix `constellTest` for all named resources being created.

Constellation OS images are currently replicated to the following regions:

* `eu-central-1`
* `us-east-2`
* `ap-south-1`

If you require the OS image to be available in another region, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md&title=Support+new+AWS+image+region:+xx-xxxx-x).

You can find a list of all [regions in AWS's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions).

Paste the output into the corresponding fields of the `constellation-conf.yaml` file.

</TabItem>
</Tabs>

<details>
<summary>Alternatively, you can manually create the IAM configuration on your CSP.</summary>

The following describes the configuration fields and how you obtain the required information or create the required resources.

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

* **subscription**: The UUID of your Azure subscription, e.g., `8b8bd01f-efd9-4113-9bd1-c82137c32da7`.

  You can view your subscription UUID via `az account show` and read the `id` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription).

* **tenant**: The UUID of your Azure tenant, e.g., `3400e5a2-8fe2-492a-886c-38cb66170f25`.

  You can view your tenant UUID via `az account show` and read the `tenant` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-ad-tenant).

* **location**: The Azure datacenter location you want to deploy your cluster in, e.g., `westus`. CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). These are:

  * `westus`
  * `eastus`
  * `northeurope`
  * `westeurope`

* **resourceGroup**: [Create a new resource group in Azure](https://learn.microsoft.com/azure/azure-resource-manager/management/manage-resource-groups-portal) for your Constellation cluster. Set this configuration     field to the name of the created resource group.

* **userAssignedIdentity**: [Create a new managed identity in Azure](https://learn.microsoft.com/azure/active-directory/managed-identities-azure-resources/how-manage-user-assigned-managed-identities). You should create the identity in a different resource group as all resources within the cluster resource group will be deleted on cluster termination.

  Add two role assignments to the identity: `Virtual Machine Contributor` and `Application Insights Component Contributor`. The `scope` of both should refer to the previously created cluster resource group.

  Set the configuration value to the full ID of the created identity, e.g., `/subscriptions/8b8bd01f-efd9-4113-9bd1-c82137c32da7/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity`. You can get it by opening the `JSON View` from the `Overview` section of the identity.

  The user-assigned identity is used by instances of the cluster to access other cloud resources.
  For more information about managed identities refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-manage-user-assigned-managed-identities).

* **appClientID**: [Create a new app registration in Azure](https://learn.microsoft.com/azure/active-directory/develop/quickstart-register-app).

  Set `Supported account types` to `Accounts in this organizational directory only` and leave the `Redirect URI` empty.

  Set the configuration value to the `Application (client) ID`, e.g., `86ec31dd-532b-4a8c-a055-dd23f25fb12f`.

  In the cluster resource group, go to `Access Control (IAM)` and set the created app registration as `Owner`.

* **clientSecretValue**: In the previously created app registration, go to `Certificates & secrets` and create a new `Client secret`.

  Set the configuration value to the secret value.

  :::tip
  Since this is a sensitive value, alternatively you can leave `clientSecretValue` empty in the configuration file and pass it via an environment variable instead. To this end, create the environment variable `CONSTELL_AZURE_CLIENT_SECRET_VALUE` and set it to the secret value.
  :::

</TabItem>

<TabItem value="gcp" label="GCP">

* **project**: The ID of your GCP project, e.g., `constellation-129857`.

  You can find it on the [welcome screen of your GCP project](https://console.cloud.google.com/welcome). For more information refer to [Google's documentation](https://support.google.com/googleapi/answer/7014113).

* **region**: The GCP region you want to deploy your cluster in, e.g., `us-west1`.

  You can find a [list of all regions in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available).

* **zone**: The GCP zone you want to deploy your cluster in, e.g., `us-west1-a`.

  You can find a [list of all zones in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available).

* **serviceAccountKeyPath**: To configure this, you need to create a GCP [service account](https://cloud.google.com/iam/docs/service-accounts) with the following permissions:

  * `Compute Instance Admin (v1) (roles/compute.instanceAdmin.v1)`
  * `Compute Network Admin (roles/compute.networkAdmin)`
  * `Compute Security Admin (roles/compute.securityAdmin)`
  * `Compute Storage Admin (roles/compute.storageAdmin)`
  * `Service Account User (roles/iam.serviceAccountUser)`

  Afterward, create and download a new JSON key for this service account. Place the downloaded file in your Constellation workspace, and set the config parameter to the filename, e.g., `constellation-129857-15343dba46cb.json`.

</TabItem>

<TabItem value="aws" label="AWS">

* **region**: The name of your chosen AWS data center region, e.g., `us-east-2`.

  Constellation OS images are currently replicated to the following regions:
  * `eu-central-1`
  * `us-east-2`
  * `ap-south-1`

  If you require the OS image to be available in another region, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md&title=Support+new+AWS+image+region:+xx-xxxx-x).

  You can find a list of all [regions in AWS's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions).

* **zone**: The name of your chosen AWS data center availability zone, e.g., `us-east-2a`.

  Learn more about [availability zones in AWS's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-availability-zones).

* **iamProfileControlPlane**: The name of an IAM instance profile attached to all control-plane nodes.

  Use the [provided Terraform script](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam) to generate the necessary profile. The profile name will be provided as Terraform output value: `control_plane_instance_profile`.

  Alternatively, you can create the AWS profile with a tool of your choice. Use the JSON policy in [main.tf](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam/main.tf) in the resource `aws_iam_policy.control_plane_policy`.

* **iamProfileWorkerNodes**: The name of an IAM instance profile attached to all worker nodes.

  Use the [provided Terraform script](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam) to generate the necessary profile. The profile name will be provided as Terraform output value: `worker_nodes_instance_profile`.

  Alternatively, you can create the AWS profile with a tool of your choice. Use the JSON policy in [main.tf](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam/main.tf) in the resource `aws_iam_policy.worker_node_policy`.

</TabItem>

</Tabs>
</details>

Now that you've configured your CSP, you can [create your cluster](./create.md).

## Deleting an IAM configuration

You can keep created IAM configurations and reuse them for new clusters. Alternatively, you can also delete them if they aren't being used anymore.

**Prerequisites:**

* [Terraform](https://developer.hashicorp.com/terraform/downloads) is installed on your machine.
* Access to the `terraform.tfstate` file created by the `constellation iam create` command.

You can delete the IAM configuration using the following commands:

```bash
# Navigate to the directory containing the terraform.tfstate file
cd constellation-iam-terraform
# Destroy the IAM configuration via Terraform
terraform destroy
# Confirm deletion by typing "yes"
```
