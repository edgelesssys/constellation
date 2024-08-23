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

### Configuration

Generate a configuration file for your cloud service provider (CSP):

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

This creates the file `constellation-conf.yaml` in the current directory. [Fill in your CSP-specific information](../getting-started/first-steps.md#create-a-cluster) before you continue.

Next, download the trusted measurements for your configured image.

```bash
constellation config fetch-measurements
```

For details, see the [verification section](../workflows/verify-cluster.md).

### Create

Choose the initial size of your cluster.
The following command creates a cluster with one control-plane and two worker nodes:

```bash
constellation create --control-plane-nodes 1 --worker-nodes 2
```

For details on the flags, consult the command help via `constellation create -h`.

*create* stores your cluster's state into a [`terraform.tfstate`](../architecture/orchestration.md#cluster-creation-process) file in your workspace.

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
