# Command-line interface

The command-line interface (CLI) is one of the key components of Constellation and is used for **orchestrating constellation clusters**. It is run by the cloud administrator on a local machine and is used to **create, manage, and update confidential clusters** directly from the command line.

You can use the CLI to create a cluster on the supported cloud platforms.
The CLI provisions the resources in your cloud environment and initiates the initialization of your cluster.
It uses a set of parameters and an optional configuration file to manage your cluster installation.
The CLI is also used for updating your cluster.

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

## Cluster creation process

To allow for fine-grained configuration of your cluster and cloud environment, Constellation supports an extensive configuration file with strong defaults. [Generating the configuration file](../../workflows/config.md) is typically the first thing you do in the workspace.

Altogether, the following files are generated during the creation of a Constellation cluster and stored in the current workspace:

- a configuration file
- a state file
- a Base64-encoded master secret
- [Terraform artifacts](../../reference/terraform.md), stored in subdirectories
- a Kubernetes `kubeconfig` file.

After the initialization of your cluster, the CLI will provide you with a Kubernetes `kubeconfig` file.
This file grants you access to your Kubernetes cluster and configures the [kubectl](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) tool.
In addition, the cluster's [identifier](../old/orchestration.md#post-installation-configuration) is returned and stored in the state file.

### Creation flow

1. The CLI `apply` command first creates the confidential VM (CVM) resources in your cloud environment and configures the network
2. Each CVM boots the Constellation node image and measures every component in the boot chain
3. The first microservice launched in each node is the [_Bootstrapper_](../old/microservices.md#bootstrapper)
4. The _Bootstrapper_ waits until it either receives an initialization request or discovers an initialized cluster
5. The CLI then connects to the _Bootstrapper_ of a selected node, sends the configuration, and initiates the initialization of the cluster
6. The _Bootstrapper_ of **that** node [initializes the Kubernetes cluster](../old/microservices.md#bootstrapper) and deploys the other Constellation [microservices](microservices.md) including the [_JoinService_](microservices.md#joinservice)
7. Subsequently, the _Bootstrappers_ of the other nodes discover the initialized cluster and send join requests to the _JoinService_
8. As part of the join request each node includes an attestation statement of its boot measurements as authentication
9. The _JoinService_ verifies the attestation statements and joins the nodes to the Kubernetes cluster
10. This process is repeated for every node joining the cluster later (e.g., through autoscaling)

## Post-installation configuration

Post-installation the CLI provides a configuration for [accessing the cluster using the Kubernetes API](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/).
The `kubeconfig` file provides the credentials and configuration for connecting and authenticating to the API server.
Once configured, orchestrate the Kubernetes cluster via `kubectl`.

After the initialization, the CLI will present you with a couple of tokens:

- The [_master secret_](../old/keys.md#master-secret) (stored in the `constellation-mastersecret.json` file by default)
- The [_clusterID_](../old/keys.md#cluster-identity) of your cluster in Base64 encoding

You can read more about these values and their meaning in the guide on [cluster identity](../old/keys.md#cluster-identity).

The _master secret_ must be kept secret and can be used to [recover your cluster](../../workflows/recovery.md).
Instead of managing this secret manually, you can [use your key management solution of choice](../old/keys.md#user-managed-key-management) with Constellation.

The _clusterID_ uniquely identifies a cluster and can be used to [verify your cluster](../../workflows/verify-cluster.md).

## Upgrades

Constellation images and microservices may need to be upgraded to new versions during the lifetime of a cluster.
Constellation implements a rolling update mechanism ensuring no downtime of the control or data plane.
You can upgrade a Constellation cluster with a single operation by using the CLI.
For step-by-step instructions on how to do this, refer to [Upgrade your cluster](../../workflows/upgrade.md).

### Attestation of upgrades

With every new image, corresponding measurements are released.
During an update procedure, the CLI provides new measurements to the [JoinService](../old/microservices.md#joinservice) securely.
New measurements for an updated image are automatically pulled and verified by the CLI following the [supply chain security concept](../old/attestation.md#chain-of-trust) of Constellation.
The [attestation section](../old/attestation.md#cluster-facing-attestation) describes in detail how these measurements are then used by the JoinService for the attestation of nodes.

<!-- soon: As the [builds of the Constellation images are reproducible](attestation.md#chain-of-trust), the updated measurements are auditable by the customer. -->
