# Create your cluster

:::info
This recording presents the essence of this page. It's recommended to read it in full for the motivation and all details.
:::

<asciinemaWidget src="/constellation/assets/create-cluster.cast" rows="20" cols="112" idleTimeLimit="3" preload="true" theme="edgeless" />

---

Creating your cluster requires two steps:

1. Creating the necessary resources in your cloud environment
2. Bootstrapping the Constellation cluster and setting up a connection

See the [architecture](../architecture/orchestration.md) section for details on the inner workings of this process.

:::tip
If you don't have a cloud subscription, you can also set up a [local Constellation cluster using virtualization](../getting-started/first-steps-local.md) for testing.
:::

## The *create* step

This step creates the necessary resources for your cluster in your cloud environment.
Before you create the cluster, make sure to have a [valid configuration file](./config.md).

### Create

<tabs groupId="provider">
<tabItem value="cli" label="CLI">

```bash
constellation create
```

*create* stores your cluster's state in a [`constellation-terraform`](../architecture/orchestration.md#cluster-creation-process) directory in your workspace.

</tabItem>
<tabItem value="terraform" label="Terraform">

Terraform allows for an easier GitOps integration as well as meeting regulatory requirements.
Since the Constellation CLI also uses Terraform under the hood, you can reuse the same Terraform files.

:::info
Familiarize with the [Terraform usage policy](../reference/terraform.md) before manually interacting with Terraform to create a cluster.
Please also refrain from changing the Terraform resource definitions, as Constellation is tightly coupled to them.
:::

Download the Terraform files for the selected CSP from the [GitHub repository](https://github.com/edgelesssys/constellation/tree/main/cli/internal/terraform/terraform).

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

The Constellation [init step](#the-init-step) requires the already created `constellation-config.yaml` and the `constellation-state.yaml`.
Create the `constellation-state.yaml` using the output from the Terraform state and the `constellation-conf.yaml`:

```bash
CONSTELL_IP=$(terraform output ip)
CONSTELL_INIT_SECRET=$(terraform output initSecret | yq -r | tr -d '\n' | base64)
yq eval '.infrastructure.initSecret ="$CONSTELL_INIT_SECRET"' --inplace constellation-state.yaml
yq eval '.infrastructure.clusterEndpoint ="$CONSTELL_IP"' --inplace constellation-state.yaml
```

</tabItem>
<tabItem value="self-managed" label="Self-managed">

Self-managed infrastructure allows for managing the cloud resources necessary for a Constellation cluster separate from the Constellation CLI.
This provides flexibility in DevOps and can meet potential regulatory requirements.

To self-manage the infrastructure of your cluster, download the Terraform files for the selected CSP from the [Constellation GitHub repository](https://github.com/edgelesssys/constellation/tree/main/cli/internal/terraform/terraform).
They contain a minimum configuration for the resources necessary to run a Constellation cluster on the corresponding CSP. From this base, you can now add, edit, or substitute resources per your own requirements with the infrastructure
management tooling of your choice. You need to keep the essential functionality of the base configuration in order for your cluster to function correctly.

Make sure all necessary resources are created, e.g., through checking your CSP's portal and retrieve the necessary values, aligned with the outputs (specified in `outputs.tf`) of the base configuration.

Fill these outputs into the corresponding fields of the `constellation-state.yaml` file. For example, fill the IP or DNS name your cluster can be reached at into the `.Infrastructure.ClusterEndpoint` field.

Continue with [initializing your cluster](#the-init-step).

</tabItem>
</tabs>

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


### Troubleshooting
In case `init` fails, the CLI collects logs from the bootstrapping instance and stores them inside `constellation-cluster.log`.
