# First steps with Constellation

The following steps guide you through the process of creating a cluster and deploying a sample app. This example assumes that you have successfully [installed and set up Constellation](install.md),
and have access to a cloud subscription.

:::tip
If you don't have a cloud subscription, check out [MiniConstellation](first-steps-local.md), which lets you set up a local Constellation cluster using virtualization.
:::

## Create a cluster

1. Create the configuration file for your selected cloud provider.

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

    This creates the file `constellation-conf.yaml` in your current working directory.

2. Fill in your cloud provider specific information.

    <Tabs groupId="csp">
    <TabItem value="azure" label="Azure (CLI)">

    You need several resources for the cluster. You can use the following `az` script to create them:

    ```bash
    RESOURCE_GROUP=constellation # enter name of new resource group for your cluster here
    LOCATION=westus # enter location of resources here
    SUBSCRIPTION_ID=$(az account show --query id --out tsv)
    SERVICE_PRINCIPAL_NAME=constell
    az group create --name "${RESOURCE_GROUP}" --location "${LOCATION}"
    az group create --name "${RESOURCE_GROUP}-identity" --location "${LOCATION}"
    az ad sp create-for-rbac -n "${SERVICE_PRINCIPAL_NAME}" --role Owner --scopes "/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${RESOURCE_GROUP}" | tee azureServiceAccountKey.json
    az identity create -g "${RESOURCE_GROUP}-identity" -n "${SERVICE_PRINCIPAL_NAME}"
    identityID=$(az identity show -n "${SERVICE_PRINCIPAL_NAME}" -g "${RESOURCE_GROUP}-identity" --query principalId --out tsv)
    az role assignment create --assignee-principal-type ServicePrincipal --assignee-object-id "${identityID}" --role 'Virtual Machine Contributor' --scope "/subscriptions/${SUBSCRIPTION_ID}"
    az role assignment create --assignee-principal-type ServicePrincipal --assignee-object-id "${identityID}" --role 'Application Insights Component Contributor' --scope "/subscriptions/${SUBSCRIPTION_ID}"
    echo "subscription: ${SUBSCRIPTION_ID}
    tenant: $(az account show --query tenantId -o tsv)
    location: ${LOCATION}
    resourceGroup: ${RESOURCE_GROUP}
    userAssignedIdentity: $(az identity show -n "${SERVICE_PRINCIPAL_NAME}" -g "${RESOURCE_GROUP}-identity" --query id --out tsv)
    appClientID: $(jq -r '.appId' azureServiceAccountKey.json)
    clientSecretValue: $(jq -r '.password' azureServiceAccountKey.json)"
    ```

    Fill the values produced by the script into your configuration file.

    By default, Constellation uses `Standard_DC4as_v5` CVMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. For CVMs, any VM type with a minimum of 4 vCPUs from the [DCasv5 & DCadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/dcasv5-dcadsv5-series) or [ECasv5 & ECadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/ecasv5-ecadsv5-series) families is supported.

    Run `constellation config instance-types` to get the list of all supported options.

    </TabItem>
    <TabItem value="azure-portal" label="Azure (Portal)">

    * **subscription**: The UUID of your Azure subscription, e.g., `8b8bd01f-efd9-4113-9bd1-c82137c32da7`.

      You can view your subscription UUID via `az account show` and read the `id` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription).

    * **tenant**: The UUID of your Azure tenant, e.g., `3400e5a2-8fe2-492a-886c-38cb66170f25`.

      You can view your tenant UUID via `az account show` and read the `tenant` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-ad-tenant).

    * **location**: The Azure datacenter location you want to deploy your cluster in, e.g., `westus`. CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). These are:

      * `westus`
      * `eastus`
      * `northeurope`
      * `westeurope`

    * **resourceGroup**: [Create a new resource group in Azure](https://learn.microsoft.com/azure/azure-resource-manager/management/manage-resource-groups-portal) for your Constellation cluster. Set this configuration field to the name of the created resource group.

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

    * **instanceType**: The VM type you want to use for your Constellation nodes.

      For CVMs, any type with a minimum of 4 vCPUs from the [DCasv5 & DCadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/dcasv5-dcadsv5-series) or [ECasv5 & ECadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/ecasv5-ecadsv5-series) families is supported. It defaults to `Standard_DC4as_v5` (4 vCPUs, 16 GB RAM).

      Run `constellation config instance-types` to get the list of all supported options.

    </TabItem>
    <TabItem value="gcp" label="GCP (CLI)">

    You need a service account for the cluster. You can use the following `gcloud` script to create it:

    ```bash
    SERVICE_ACCOUNT_ID=constell # enter name of service account here
    PROJECT_ID= # enter project id here
    SERVICE_ACCOUNT_EMAIL=${SERVICE_ACCOUNT_ID}@${PROJECT_ID}.iam.gserviceaccount.com
    gcloud iam service-accounts create "${SERVICE_ACCOUNT_ID}" --description="Service account used inside Constellation" --display-name="Constellation service account" --project="${PROJECT_ID}"
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" --role='roles/compute.instanceAdmin.v1'
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" --role='roles/compute.networkAdmin'
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" --role='roles/compute.securityAdmin'
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" --role='roles/compute.storageAdmin'
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" --role='roles/iam.serviceAccountUser'
    gcloud iam service-accounts keys create gcpServiceAccountKey.json --iam-account="${SERVICE_ACCOUNT_EMAIL}"
    echo "project: ${PROJECT_ID}
    serviceAccountKeyPath: $(realpath gcpServiceAccountKey.json)"
    ```

    Fill the values produced by the script into your configuration file.

    By default, Constellation uses `n2d-standard-4` VMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. Supported are all machines from the N2D family with a minimum of 4 vCPUs. Refer to [N2D machine series](https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines) or run `constellation config instance-types` to get the list of all supported options.

    </TabItem>
    <TabItem value="gcp-console" label="GCP (Console)">

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

    * **instanceType**: The VM type you want to use for your Constellation nodes.

      Supported are all machines from the N2D family with a minimum of 4 vCPUs. It defaults to `n2d-standard-4` (4 vCPUs, 16 GB RAM), but you can use any other VMs from the same family. Refer to [N2D machine series](https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines) or run `constellation config instance-types` to get the list of all supported options.

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

    * **image**: The ID of the amazon machine image (AMI) the Constellation nodes will use:

      Constellation OS images are available with the following IDs:

      | AMI | Region |
      | - | - |
      | `ami-0e27ebcefc38f648b` | `eu-central-1` |
      | `ami-098cd37f66523b7c3` | `us-east-2` |
      | `ami-04a87d302e2509aad` | `ap-south-1` |

    * **iamProfileControlPlane**: The name of an IAM instance profile attached to all control-plane nodes.

      Use the [provided Terraform script](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam) to generate the necessary profile. The profile name will be provided as Terraform output value: `control_plane_instance_profile`.

      Alternatively, you can create the AWS profile with a tool of your choice. Use the JSON policy in [main.tf](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam/main.tf) in the resource `aws_iam_policy.control_plane_policy`.

    * **iamProfileWorkerNodes**: The name of an IAM instance profile attached to all worker nodes.

      Use the [provided Terraform script](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam) to generate the necessary profile. The profile name will be provided as Terraform output value: `worker_nodes_instance_profile`.

      Alternatively, you can create the AWS profile with a tool of your choice. Use the JSON policy in [main.tf](https://github.com/edgelesssys/constellation/tree/release/v2.2/hack/terraform/aws/iam/main.tf) in the resource `aws_iam_policy.worker_node_policy`.

    </TabItem>
    </Tabs>

    :::info

    In case you don't have access to CVMs on Azure, you may use less secure  [trusted launch VMs](../workflows/trusted-launch.md) instead. For this, set **confidentialVM** to `false` in the configuration file.

    :::

3. Download the trusted measurements for your configured image.

    ```bash
    constellation config fetch-measurements
    ```

    For details, see the [verification section](../workflows/verify-cluster.md).

4. Create the cluster with one control-plane node and two worker nodes. `constellation create` uses options set in `constellation-conf.yaml`.

    :::tip

    On Azure, you may need to wait 15+ minutes at this point for role assignments to propagate.

    :::

    ```bash
    constellation create --control-plane-nodes 1 --worker-nodes 2 -y
    ```

    This should give the following output:

    ```shell-session
    $ constellation create ...
    Your Constellation cluster was created successfully.
    ```

5. Initialize the cluster

    :::caution

    In this release of Constellation, initialization on **Azure** might be slow and might take up to 60 minutes to initialize all Kubernetes nodes. This has been fixed in v2.4.

    :::

    ```bash
    constellation init
    ```

    This should give the following output:

    ```shell-session
    $ constellation init
    Your Constellation master secret was successfully written to ./constellation-mastersecret.json
    Initializing cluster ...
    Your Constellation cluster was successfully initialized.

    Constellation cluster identifier  g6iMP5wRU1b7mpOz2WEISlIYSfdAhB0oNaOg6XEwKFY=
    Kubernetes configuration          constellation-admin.conf

    You can now connect to your cluster by executing:
            export KUBECONFIG="$PWD/constellation-admin.conf"
    ```

    The cluster's identifier will be different in your output.
    Keep `constellation-mastersecret.json` somewhere safe.
    This will allow you to [recover your cluster](../workflows/recovery.md) in case of a disaster.

    :::info

    Depending on your CSP and region, `constellation init` may take 10+ minutes to complete.

    :::

6. Configure kubectl

    ```bash
    export KUBECONFIG="$PWD/constellation-admin.conf"
    ```

## Deploy a sample application

1. Deploy the [emojivoto app](https://github.com/BuoyantIO/emojivoto)

    ```bash
    kubectl apply -k github.com/BuoyantIO/emojivoto/kustomize/deployment
    ```

2. Expose the frontend service locally

    ```bash
    kubectl wait --for=condition=available --timeout=60s -n emojivoto --all deployments
    kubectl -n emojivoto port-forward svc/web-svc 8080:80 &
    curl http://localhost:8080
    kill %1
    ```

## Terminate your cluster

```bash
constellation terminate
```

This should give the following output:

```shell-session
$ constellation terminate
You are about to terminate a Constellation cluster.
All of its associated resources will be DESTROYED.
This includes any other Terraform workspace in the current directory.
This action is irreversible and ALL DATA WILL BE LOST.
Do you want to continue? [y/n]:
```

Confirm with `y` to terminate the cluster:

```shell-session
Terminating ...
Your Constellation cluster was terminated successfully.
```

:::tip

On Azure, if you have used the `az` script, you can keep the prerequisite resources and reuse them for a new cluster.

Or you can delete them:

```bash
RESOURCE_GROUP=constellation # name of your cluster resource group
APPID=$(jq -r '.appId' azureServiceAccountKey.json)
az ad sp delete --id "${APPID}"
az group delete -g "${RESOURCE_GROUP}-identity" --yes --no-wait
az group delete -g "${RESOURCE_GROUP}" --yes --no-wait
```

:::
