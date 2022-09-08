# Create your cluster

Creating your cluster requires two steps:

1. Creating the necessary resources in your cloud environment
2. Bootstrapping the Constellation cluster and setting up a connection

See the [architecture](../architecture/orchestration.md) section for details on the inner workings of this process.

## The *create* step

This step creates the necessary resources for your cluster in your cloud environment.

### Prerequisites

Before creating your cluster you need to decide on

* the size of your cluster (the number of control-plane and worker nodes)
* the machine type of your nodes (depending on the availability in your cloud environment)
* whether to enable autoscaling for your cluster (automatically adding and removing nodes depending on resource demands)

You can find the currently supported machine types for your cloud environment in the [installation guide](../architecture/orchestration.md).

### Configuration

Constellation can generate a configuration file for your cloud provider:

<tabs groupId="csp">
<tabItem value="azure" label="Azure" default>

```bash
constellation config generate azure
```

</tabItem>
<tabItem value="gcp" label="GCP" default>

```bash
constellation config generate gcp
```

</tabItem>
</tabs>

This creates the file `constellation-conf.yaml` in the current directory. You must edit it before you can execute the next steps. See the [reference section](../reference/config.md) for details.

Next, download the latest trusted measurements for your configured image.

```bash
constellation config fetch-measurements
```

For more details, see the [verification section](../workflows/verify.md).

### Create

The following command creates a cluster with one control-plane and two worker nodes:

```bash
constellation create --control-plane-nodes 1 --worker-nodes 2 -y
```

For details on the flags and a list of supported instance types, consult the command help via `constellation create -h`.

*create* will store your cluster's configuration to a file named [`constellation-state.json`](../architecture/orchestration.md#installation-process) in your current directory.

## The *init* step

This step bootstraps your cluster and configures your Kubernetes client.

### Init

The following command initializes and bootstraps your cluster:

```bash
constellation init
```

To enable autoscaling in your cluster, add the `--autoscale` flag:

```bash
constellation init --autoscale
```

Next, configure kubectl for your Constellation cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
kubectl get nodes -o wide
```

That's it. You've successfully created a Constellation cluster.
