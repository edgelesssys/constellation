# Create your cluster

Creating your cluster requires two steps:

1. Creating the necessary resources in your cloud environment
2. Bootstrapping the Constellation cluster and setting up a connection

See the [architecture](../architecture/orchestration.md) section for details on the inner workings of this process.

:::tip
If you don't have a cloud subscription, check out [MiniConstellation](../getting-started/first-steps-local.md), which lets you set up a local Constellation cluster using virtualization.
:::

## The *create* step

This step creates the necessary resources for your cluster in your cloud environment.
Before you create the cluster, make sure to have a [valid configuration file](./config.md).

### Create

<Tabs groupId="provider">
<TabItem value="cli" label="CLI">

Choose the initial size of your cluster.
The following command creates a cluster with one control-plane and two worker nodes:

```bash
constellation create --control-plane-nodes 1 --worker-nodes 2
```

For details on the flags, consult the command help via `constellation create -h`.

*create* stores your cluster's state into a [`terraform.tfstate`](../architecture/orchestration.md#cluster-creation-process) file in your workspace.

</TabItem>
<TabItem value="terraform" label="Terraform">

Constellation supports managing the infrastructure via Terraform. This allows for an easier GitOps integration as well as meeting regulatory requirements.
Since the Constellation CLI also uses Terraform under the hood, you can reuse the same Terraform files.
For now, please refrain from changing the Terraform resource definitions, as Constellation is tightly coupled to them.

Download the Terraform files for the selected CSP from the [GitHub repository](https://github.com/edgelesssys/constellation/tree/main/terraform/infrastructure).

Create a `terraform.tfvars` file.
There, define all needed variables found in `variables.tf` using the values from the `constellation-config.yaml`.

To find the image reference for your CSP and region, execute:

```bash
CONSTELL_VER=vX.Y.Z
curl -s https://cdn.confidential.cloud/constellation/v1/ref/-/stream/stable/$CONSTELL_VER/image/info.json | jq
```

Initialize and apply Terraform to create the configured infrastructure:

```bash
terraform init
terraform apply
```

The Constellation [init step](#the-init-step) requires the already created `constellation-config.yaml` and the `constellation-id.json`.
Create the `constellation-id.json` using the output from the Terraform state and the `constellation-conf.yaml`:

```bash
CONSTELL_IP=$(terraform output ip)
CONSTELL_INIT_SECRET=$(terraform output initSecret | jq -r | tr -d '\n' | base64)
CONSTELL_CSP=$(cat constellation-conf.yaml | yq ".provider | keys | .[0]")
jq --null-input --arg cloudprovider "$CONSTELL_CSP" --arg ip "$CONSTELL_IP" --arg initsecret "$CONSTELL_INIT_SECRET" '{"cloudprovider":$cloudprovider,"ip":$ip,"initsecret":$initsecret}' > constellation-id.json
```

</TabItem>
</Tabs>

## The *init* step

The following command initializes and bootstraps your cluster:

```bash
constellation init
```

Next, configure `kubectl` for your cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
```

🏁 That's it. You've successfully created a Constellation cluster.
