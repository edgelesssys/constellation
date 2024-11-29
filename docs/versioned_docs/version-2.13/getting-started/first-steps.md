# First steps with Constellation

The following steps guide you through the process of creating a cluster and deploying a sample app. This example assumes that you have successfully [installed and set up Constellation](install.md),
and have access to a cloud subscription.

:::tip
If you don't have a cloud subscription, you can also set up a [local Constellation cluster using virtualization](../getting-started/first-steps-local.md) for testing.
:::

:::note
If you encounter any problem with the following steps, make sure to use the [latest release](https://github.com/edgelesssys/constellation/releases/latest) and check out the [known issues](https://github.com/edgelesssys/constellation/issues?q=is%3Aopen+is%3Aissue+label%3A%22known+issue%22).
:::

## Create a cluster

1. Create the [configuration file](../workflows/config.md) and state file for your cloud provider.

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

2. Create your [IAM configuration](../workflows/config.md#creating-an-iam-configuration).

    <Tabs groupId="csp">

    <TabItem value="azure" label="Azure">

    ```bash
    constellation iam create azure --region=westus --resourceGroup=constellTest --servicePrincipal=spTest --update-config
    ```

    This command creates IAM configuration on the Azure region `westus` creating a new resource group `constellTest` and a new service principal `spTest`. It also updates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

    Note that CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). These are:
    * `westus`
    * `eastus`
    * `northeurope`
    * `westeurope`
    * `southeastasia`

    </TabItem>

    <TabItem value="gcp" label="GCP">

    ```bash
    constellation iam create gcp --projectID=yourproject-12345 --zone=europe-west2-a --serviceAccountID=constell-test --update-config
    ```

    This command creates IAM configuration in the GCP project `yourproject-12345` on the GCP zone `europe-west2-a` creating a new service account `constell-test`. It also updates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

    Note that only regions offering CVMs of the `C2D` or `N2D` series are supported. You can find a [list of all regions in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available), which you can filter by machine type `C2D` or `N2D`.

    </TabItem>

    <TabItem value="aws" label="AWS">

    ```bash
    constellation iam create aws --zone=us-east-2a --prefix=constellTest --update-config
    ```

    This command creates IAM configuration for the AWS zone `us-east-2a` using the prefix `constellTest` for all named resources being created. It also updates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

    Depending on the attestation variant selected on config generation, different regions are available.
    AMD SEV-SNP machines (requires the default attestation variant `awsSEVSNP`) are currently available in the following regions:
     * `eu-west-1`
     * `us-east-2`

    You can find a list of regions that support AMD SEV-SNP in [AWS's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html).

    NitroTPM machines (requires the attestation variant `awsNitroTPM`) are available in all regions.
    Constellation OS images are currently replicated to the following regions:
     * `eu-central-1`
     * `eu-west-1`
     * `eu-west-3`
     * `us-east-2`
     * `ap-south-1`

    If you require the OS image to be available in another region, [let us know](https://github.com/edgelesssys/constellation/issues/new?assignees=&labels=&template=feature_request.md&title=Support+new+AWS+image+region:+xx-xxxx-x).

    You can find a list of all [regions in AWS's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions).

    </TabItem>
    </Tabs>

    :::tip
    To learn about all options you have for managing IAM resources and Constellation configuration, see the [Configuration workflow](../workflows/config.md).
    :::

<!--
    :::info

    In case you don't have access to CVMs on Azure, you may use less secure  [trusted launch VMs](../workflows/trusted-launch.md) instead. For this, set **confidentialVM** to `false` in the configuration file.

    :::
-->

3. Create the cluster. `constellation create` uses options set in `constellation-conf.yaml`.
    If you want to manually manage your cloud resources, for example by using [Terraform](../reference/terraform.md), follow the corresponding instructions in the [Create workflow](../workflows/create.md).

    :::tip

    On Azure, you may need to wait 15+ minutes at this point for role assignments to propagate.

    :::

    ```bash
    constellation create -y
    ```

    This should give the following output:

    ```shell-session
    $ constellation create -y
    Your Constellation cluster was created successfully.
    ```

4. Initialize the cluster.

    ```bash
    constellation apply
    ```

    This should give the following output:

    ```shell-session
    $ constellation apply
    Your Constellation master secret was successfully written to ./constellation-mastersecret.json
    Connecting
    Initializing cluster
    Installing Kubernetes components
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

    Depending on your CSP and region, `constellation apply` may take 10+ minutes to complete.

    :::

5. Configure kubectl.

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

Use the CLI to terminate your cluster. If you manually used [Terraform](../reference/terraform.md) to manage your cloud resources, follow the corresponding instructions in the [Terminate workflow](../workflows/terminate.md).

```bash
constellation terminate
```

This should give the following output:

```shell-session
$ constellation terminate
You are about to terminate a Constellation cluster.
All of its associated resources will be DESTROYED.
This action is irreversible and ALL DATA WILL BE LOST.
Do you want to continue? [y/n]:
```

Confirm with `y` to terminate the cluster:

```shell-session
Terminating ...
Your Constellation cluster was terminated successfully.
```

Optionally, you can also [delete your IAM resources](../workflows/config.md#deleting-an-iam-configuration).
