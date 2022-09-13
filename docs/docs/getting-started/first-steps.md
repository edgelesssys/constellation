# First steps

The following steps will guide you through the process of creating a cluster and deploying a sample app. This example assumes that you have successfully [installed and set up Constellation](install.md).

## Create a cluster

1. Create the configuration file for your selected cloud provider.

    <tabs groupId="csp">
    <tabItem value="azure" label="Azure" default>

    ```bash
    constellation config generate azure
    ```

    </tabItem>
    <tabItem value="gcp" label="GCP">

    ```bash
    constellation config generate gcp
    ```

    </tabItem>
    </tabs>

    This creates the file `constellation-conf.yaml` in your current working directory.

2.  Fill in your cloud provider specific information.

    <tabs groupId="csp">
    <tabItem value="azure" label="Azure (CLI)" default>

    For a quick start it's recommended to use the following `az` script to automatically create all required resources:

    ```bash
    RESOURCE_GROUP=constellation # enter name of resource group here
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
    echo "subscription: ${SUBSCRIPTION_ID}"
    echo "tenant: $(az account show --query tenantId -o tsv)"
    echo "location: ${LOCATION}"
    echo "resourceGroup: ${RESOURCE_GROUP}"
    echo "userAssignedIdentity: $(az identity show -n "${SERVICE_PRINCIPAL_NAME}" -g "${RESOURCE_GROUP}-identity" --query id --out tsv)"
    echo "appClientID: $(jq -r '.appId' azureServiceAccountKey.json)"
    echo "clientSecretValue: $(jq -r '.password' azureServiceAccountKey.json)"
    ```

    Fill the values produced by the script into your configuration file.

    By default, Constellation uses `Standard_DC4as_v5` CVMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. For CVMs, any VM type with a minimum of 4 vCPUs from the [DCasv5 & DCadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/dcasv5-dcadsv5-series) or [ECasv5 & ECadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/ecasv5-ecadsv5-series) families is supported.

    Run `constellation config instance-types` to get the list of all supported options.

    :::info

    In case you don't have access to CVMs, you may use less secure [trusted launch VMs](../workflows/trusted-launch.md) instead. For this, set **confidentialVM** to `false` in the configuration file.

    :::

    </tabItem>
    <tabItem value="azure-portal" label="Azure (Portal)">

    * **subscription**: Is the UUID of your Azure subscription, e.g., `8b8bd01f-efd9-4113-9bd1-c82137c32da7`.

        You can view your subscription UUID via `az account show` and read the `id` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription).

    * **tenant**: Is the UUID of your Azure tenant, e.g., `3400e5a2-8fe2-492a-886c-38cb66170f25`.

        You can view your tenant UUID via `az account show` and read the `tenant` field. For more information refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-ad-tenant).

    * **location**: Is the Azure datacenter location you want to deploy your cluster in, e.g., `westus`. Notice that CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). Currently these are supported:

        * `westus`
        * `eastus`
        * `northeurope`
        * `westeurope`

    * **instanceType**: Is the VM type you want to use for your Constellation nodes.

      For CVMs, any type with a minimum of 4 vCPUs from the [DCasv5 & DCadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/dcasv5-dcadsv5-series) or [ECasv5 & ECadsv5](https://docs.microsoft.com/en-us/azure/virtual-machines/ecasv5-ecadsv5-series) families is supported. It defaults to `Standard_DC4as_v5` (4 vCPUs, 16 GB RAM).

      Run `constellation config instance-types` to get the list of all supported options.

    * **resourceGroup**: [Create a new resource group in Azure](https://portal.azure.com/#create/Microsoft.ResourceGroup), to deploy your Constellation cluster into. Afterwards set the configuration field to the name of the created resource group, e.g., `constellation`.

    * **userAssignedIdentity**: [Create a new managed identity in Azure](https://portal.azure.com/#create/Microsoft.ManagedIdentity). Notice that the identity should be created in a different resource group as all resources within the cluster resource group will be deleted on cluster termination.

        After creation, add two role assignments to the identity, for the roles `Virtual Machine Contributor` and `Application Insights Component Contributor`. The `scope` of both should refer to the previously created resource group.

        Set the configuration value to the full ID of the created identity, e.g., `/subscriptions/8b8bd01f-efd9-4113-9bd1-c82137c32da7/resourcegroups/constellation-identity/providers/Microsoft.ManagedIdentity/userAssignedIdentities/constellation-identity`.

        The user-assigned identity is used by instances of the cluster to access other cloud resources.

        For more information about managed identities refer to [Azure's documentation](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-manage-user-assigned-managed-identities).

    * **appClientID**: [Create a new app registration in Azure](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/CreateApplicationBlade/quickStartType~/null/isMSAApp~/false).

        As `Supported account types` choose `Accounts in this organizational directory only`, and leave the `Redirect URI` empty.

        In the cluster resource group, go to `Access Control (IAM)`, and set the created app registration as `Owner`.

        Set the configuration value to the `Application (client) ID`, e.g., `86ec31dd-532b-4a8c-a055-dd23f25fb12f`.

    * **clientSecretValue**: In our previously created app registration, go to `Certificates & secrets` and create a new `Client secret`.

        Set the configuration value to the secret value.

    </tabItem>
    <tabItem value="gcp" label="GCP (CLI)">

    For a quick start it's recommended to use the following `gcloud` script to automatically create all required resources:

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
    echo "project: ${PROJECT_ID}"
    echo "serviceAccountKeyPath: $(realpath gcpServiceAccountKey.json)"
    ```

    Fill the values produced by the script into your configuration file.

    By default, Constellation uses `n2d-standard-4` VMs (4 vCPUs, 16 GB RAM) to create your cluster. Optionally, you can switch to a different VM type by modifying **instanceType** in the configuration file. Supported are all machines from the N2D family. Refer to [N2D machine series](https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines) or run `constellation config instance-types` to get the list of all supported options.

    </tabItem>
    <tabItem value="gcp-console" label="GCP (Console)">

    * **project**: Is the ID of your GCP project, e.g., `constellation-129857`.

        You will find it on the [welcome screen of your GCP project](https://console.cloud.google.com/welcome). For more information refer to [Google's documentation](https://support.google.com/googleapi/answer/7014113).

    * **region**: Is the GCP region you want to deploy your cluster in, e.g., `us-west-1`.

        You can find a [list of all regions in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available).

    * **zone**: Is the GCP zone you want to deploy your cluster in, e.g., `us-west-1a`.

        You can find a [list of all zones in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available).

    * **instanceType**: Is the VM type you want to use for your Constellation nodes.

        Supported are all machines from the N2D family. It defaults to `n2d-standard-4` (4 vCPUs, 16 GB RAM), but you can use any other VMs from the same family. Refer to [N2D machine series](https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machines) or run `constellation config instance-types` to get the list of all supported options.

    * **serviceAccountKeyPath**: To configure this, you need to create a GCP [service account](https://cloud.google.com/iam/docs/service-accounts) with the following permissions:

        - `Compute Instance Admin (v1) (roles/compute.instanceAdmin.v1)`
        - `Compute Network Admin (roles/compute.networkAdmin)`
        - `Compute Security Admin (roles/compute.securityAdmin)`
        - `Compute Storage Admin (roles/compute.storageAdmin)`
        - `Service Account User (roles/iam.serviceAccountUser)`

        Afterwards, create and download a new `JSON` key for this service account. Place the downloaded file in your Constellation workspace, and set the config parameter to the filename, e.g., `constellation-129857-15343dba46cb.json`.

    </tabItem>
    </tabs>

3. Download the measurements for your configured image.

    ```bash
    constellation config fetch-measurements
    ```

    This command is necessary to download the latest trusted measurements for your configured image.

    For more details, see the [verification section](../workflows/verify-cluster.md).

4. Create the cluster with one control-plane node and two worker nodes. `constellation create` uses options set in `constellation-conf.yaml` automatically.

    :::tip

    On Azure, after creating a user-assigned managed identity and app registration, Azure AD needs time to propagate permissions to different regions due to replication lag. To avoid permission errors, it's recommended to wait at least 15 minutes after role assignments before creating the cluster.

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

    ```bash
    constellation init
    ```

    This should give the following output:

    ```shell-session
    $ constellation init
    Creating service account ...
    Your Constellation cluster was successfully initialized.
    Constellation cluster's identifier  g6iMP5wRU1b7mpOz2WEISlIYSfdAhB0oNaOg6XEwKFY=
    Kubernetes configuration            constellation-admin.conf
    You can now connect to your cluster by executing:
            export KUBECONFIG="$PWD/constellation-admin.conf"
    ```

    The cluster's identifier will be different in your output.
    Keep `constellation-mastersecret.json` somewhere safe.
    This will allow you to [recover your cluster](../workflows/recovery.md) in case of a disaster.

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
Terminating ...
Your Constellation cluster was terminated successfully.
```

In case you have used `az` CLI to create your environment, make sure to clean up afterwards:

```bash
APPID=$(jq -r '.appId' azureServiceAccountKey.json)
az ad sp delete --id "${APPID}"
az group delete -g "${RESOURCE_GROUP}-identity" --yes --no-wait
az group delete -g "${RESOURCE_GROUP}" --yes --no-wait
```
