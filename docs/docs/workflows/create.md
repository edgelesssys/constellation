# Create your cluster

Creating your cluster requires two steps:

1. Creating the necessary resources in your cloud environment
2. Bootstrapping the Constellation cluster and setting up a connection

See the [architecture](../architecture/orchestration.md) section for details on the inner workings of this process.

:::tip
If you don't have a cloud subscription, check out [MiniConstellation](../getting-started/mini-constellation.md), which lets you set up a local Constellation cluster using virtualization.
:::

## The *create* step

This step creates the necessary resources for your cluster in your cloud environment.

### Configuration

Generate a configuration file for your cloud service provider (CSP):

<tabs groupId="csp">
<tabItem value="azure" label="Azure">

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

*create* stores your cluster's configuration to a file named [`constellation-state.json`](../architecture/orchestration.md#installation-process) in your current directory.

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
