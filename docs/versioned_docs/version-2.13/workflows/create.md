# Create your cluster

:::info
This recording presents the essence of this page. It's recommended to read it in full for the motivation and all details.
:::

<AsciinemaWidget src="/constellation/assets/create-cluster.cast" rows="20" cols="112" idleTimeLimit="3" preload="true" theme="edgeless" />

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

<Tabs groupId="usage">
<TabItem value="cli" label="CLI">

```bash
constellation create
```

*create* stores your cluster's state in a [`constellation-terraform`](../architecture/orchestration.md#cluster-creation-process) directory in your workspace.

</TabItem>
<TabItem value="self-managed" label="Self-managed">

Self-managed infrastructure allows for more flexibility in the setup, by separating the infrastructure setup from the Constellation cluster management.
This provides flexibility in DevOps and can meet potential regulatory requirements.
It's recommended to use Terraform for infrastructure management, but you can use any tool of your choice.

:::info

  When using Terraform, you can use the [Terraform module](./terraform-module.md) for ease of use to manage the entire Constellation cluster.

:::

You can refer to the Terraform files for the selected CSP from the [Constellation GitHub repository](https://github.com/edgelesssys/constellation/tree/main/terraform/infrastructure) for a minimum Constellation cluster configuration. From this base, you can now add, edit, or substitute resources per your own requirements with the infrastructure
management tooling of your choice. You need to keep the essential functionality of the base configuration in order for your cluster to function correctly.

<!-- vale off -->

:::info

  On Azure, a manual update to the MAA provider's policy is necessary.
  You can apply the update with the following command after creating the infrastructure, with `<URL>` being the URL of the MAA provider (i.e., `$(terraform output attestationURL | jq -r)`, when using the minimal Terraform configuration).

  ```bash
  constellation maa-patch <URL>
  ```

:::

<!-- vale on -->

Make sure all necessary resources are created, e.g., through checking your CSP's portal and retrieve the necessary values, aligned with the outputs (specified in `outputs.tf`) of the base configuration.

Fill these outputs into the corresponding fields of the `Infrastructure` block inside the `constellation-state.yaml` file. For example, fill the IP or DNS name your cluster can be reached at into the `.Infrastructure.ClusterEndpoint` field.

Continue with [initializing your cluster](#the-apply-step).

</TabItem>
</Tabs>

## The *apply* step

The following command initializes and bootstraps your cluster:

```bash
constellation apply
```

Next, configure `kubectl` for your cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
```

🏁 That's it. You've successfully created a Constellation cluster.

### Troubleshooting

In case `apply` fails, the CLI collects logs from the bootstrapping instance and stores them inside `constellation-cluster.log`.
