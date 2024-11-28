# Command-line interface

The command-line interface (CLI) is one of the key components of Constellation and is used for **orchestrating constellation clusters**. It is run by the cloud administrator on a local machine and is used to **create, manage, and update confidential clusters** directly from the command line.

You can use the CLI to create a cluster on the [supported cloud platforms](../../overview/product.md).

The CLI uses a set of parameters and an optional configuration file to manage your cluster installation. You can also use the CLI to update your cluster.

## Workspaces

Each Constellation cluster has an associated _workspace_.
The workspace is where data such as the Constellation state and config files are stored.
Each workspace is associated with a single cluster and configuration.
The CLI stores state in the local filesystem making the current directory the active workspace.
Multiple clusters require multiple workspaces, hence, multiple directories.
Note that every operation on a cluster always has to be performed from the directory associated with its workspace.

You may copy files from the workspace to other locations,
but you shouldn't move or delete them while the cluster is still being used.
The Constellation CLI takes care of managing the workspace.
Only when a cluster was terminated, and you are sure the files aren't needed anymore, should you remove a workspace.

## Cluster creation

The `apply` command creates the necessary resources in your cloud environment, bootstraps the Constellation cluster, and securely installs all Kubernetes binaries to ensure the cluster's integrity. For a detailed description, please visit our [Protocol overview](../overview.md#cluster-creation).

## Post-installation configuration

After installation, the CLI provides a configuration for [accessing the cluster using the Kubernetes API](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/).

The `kubeconfig` file contains the credentials and configuration needed for connecting and authenticating to the API server. Once configured, you can orchestrate the Kubernetes cluster via `kubectl`.

After the initialization, the CLI will present you with a couple of tokens:

- The [_master secret_](../security/keys.md#master-secret) (stored in the `constellation-mastersecret.json` file by default)
- The [_clusterID_](../security/keys.md#cluster-identity) of your cluster in Base64 encoding

You can read more about these values and their meaning in the guide on [cluster identity](../security/keys.md#cluster-identity).

The _master secret_ must be kept secret and can be used to [recover your cluster](../../workflows/recovery.md).
Instead of managing this secret manually, you can [use your key management solution of choice](../security/keys.md#user-managed-key-management) with Constellation.

The _clusterID_ uniquely identifies a cluster and can be used to [verify your cluster](../../workflows/verify-cluster.md).

## Constellation upgrades

Constellation images and microservices may need to be upgraded to new versions during the lifetime of a cluster. This can be done by running the `upgrade` command. An explanation of the update protocol can be found in our [Protocol overview](../overview.md#cluster-upgrade).

<!-- soon: As the [builds of the Constellation images are reproducible](attestation.md#chain-of-trust), the updated measurements are auditable by the customer. -->
