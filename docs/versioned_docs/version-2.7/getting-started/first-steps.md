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

1. Create the configuration file and IAM resources for your selected cloud provider

    First, you need to create a [configuration file](../workflows/config.md) and an [IAM configuration](../workflows/config.md#creating-an-iam-configuration). The easiest way to do this is the following CLI command:

    <Tabs groupId="csp">

    <TabItem value="azure" label="Azure">

    ```bash
    constellation iam create azure --region=westus --resourceGroup=constellTest --servicePrincipal=spTest --generate-config
    ```

    This command creates IAM configuration on the Azure region `westus` creating a new resource group `constellTest` and a new service principal `spTest`. It also creates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

    Note that CVMs are currently only supported in a few regions, check [Azure's products available by region](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=virtual-machines&regions=all). These are:
    * `westus`
    * `eastus`
    * `northeurope`
    * `westeurope`

    </TabItem>

    <TabItem value="gcp" label="GCP">

    ```bash
    constellation iam create gcp --projectID=yourproject-12345 --zone=europe-west2-a --serviceAccountID=constell-test --generate-config
    ```

    This command creates IAM configuration in the GCP project `yourproject-12345` on the GCP zone `europe-west2-a` creating a new service account `constell-test`. It also creates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

    Note that only regions offering CVMs of the `C2D` or `N2D` series are supported. You can find a [list of all regions in Google's documentation](https://cloud.google.com/compute/docs/regions-zones#available), which you can filter by machine type `C2D` or `N2D`.

    </TabItem>

    <TabItem value="aws" label="AWS">

    ```bash
    constellation iam create aws --zone=eu-central-1a --prefix=constellTest --generate-config
    ```

    This command creates IAM configuration for the AWS zone `eu-central-1a` using the prefix `constellTest` for all named resources being created. It also creates the configuration file `constellation-conf.yaml` in your current directory with the IAM values filled in.

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

2. Create the cluster with one control-plane node and two worker nodes. `constellation create` uses options set in `constellation-conf.yaml`.
    If you want to manually use [Terraform](../reference/terraform.md) for managing the cloud resources instead, follow the corresponding instructions in the [Create workflow](../workflows/create.md).

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

3. Initialize the cluster

    ```bash
    constellation init
    ```

    This should give the following output:

    ```shell-session
    $ constellation init
    Your Constellation master secret was successfully written to ./constellation-mastersecret.json
    Note: If you just created the cluster, it can take a few minutes to connect.
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

4. Configure kubectl

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
