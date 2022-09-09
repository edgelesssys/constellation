# Orchestrating Constellation clusters

You can use the CLI to create a cluster on the supported cloud platforms.
The CLI provisions the resources in your cloud environment and initiates the initialization of your cluster.
It uses a set of parameters and an optional configuration file to manage your cluster installation.
The CLI is also used for updating your cluster.

## Workspaces

Each Constellation cluster has an associated *workspace*.
The workspace is where the persistent data such as the Constellation state, config, and ID files are stored.
Each workspace is associated with a single cluster and configuration.
Currently, the CLI stores state in the local filesystem making the current directory the active workspace.
Multiple clusters require multiple workspaces, hence, multiple directories.
Note that every operation on a cluster always has to be performed from the directory associated with its workspace.

## Cluster creation process

Releases of Constellation are [published on GitHub](https://github.com/edgelesssys/constellation/releases).

To allow for fine-grained configuration of your cluster and cloud environment, Constellation supports an extensive configuration file with strong defaults.
The CLI provides you with a good default configuration which can be generated with `constellation config generate`. Some cloud account-specific information is always required and to be set by the user.
Details and examples can be found in the [reference guide](../reference/config.md).

The following files are generated during the creation of a Constellation cluster and stored in the current workspace:

* a configuration file
* a state file
* an ID file
* a base64 encoded master secret
* a Kubernetes `kubeconfig` file.

Constellation must store the state of its created infrastructure and configuration.
This state is used by Constellation to map real-world resources to your configuration and keep track of metadata.
This state is stored by default in a local file named `constellation-state.json`.

After the successful creation of your cluster, the CLI will provide you with a Kubernetes `kubeconfig` file.
This file provides you with access to your Kubernetes cluster and configures the [kubectl](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) tool.
In addition, the cluster's [identifier](orchestration.md#post-installation-configuration) is returned and stored in a file called `constellation-id.json`

### Creation process details

1. The CLI `create` command creates the confidential VMs (CVMs) resources in your cloud environment and configures the network
2. Each CVM boots the Constellation node image and measures every component in the boot chain
3. The first component launched in each node is the [*Bootstrapper*](components.md#bootstrapper)
4. The *Bootstrapper* waits until it either receives an initialization request or discovers an initialized cluster
5. The CLI `init` command connects to the *Bootstrapper* of a selected node, sends the configuration, and initiates the initialization of the cluster
6. The *Bootstrapper* of **that** node [initializes the Kubernetes cluster](components.md#bootstrapper) and deploys the other Constellation [components](components.md) including the [*JoinService*](components.md#joinservice)
7. Subsequently, the *Bootstrappers* of the other nodes discover the initialized cluster and send join requests to the *JoinService*
8. As part of the join request each node includes an attestation statement of its boot measurements as a form of authentication
9. The *JoinService* verifies the attestation statements and joins the nodes to the Kubernetes cluster
10. This process is repeated for every node joining the cluster later (e.g., through autoscaling)

## Post-installation configuration

Post-installation the CLI provides a configuration for [accessing the cluster using the Kubernetes API](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/).
The `kubeconfig` file provides the credentials and configuration for connecting and authenticating to the API server.
Once configured, orchestrate the Kubernetes cluster via `kubectl`.

Keep the state files in the workspace directory such as the `constellation-state.json` for the CLI to be able to manage your cluster.
Without it, you won't be able to modify or terminate your cluster.

After the initialization, the CLI will present you with a couple of tokens:

* The [*master secret*](keys.md#master-secret) (stored in the `constellation-mastersecret.json` file by default)
* The [*clusterID*](keys.md#cluster-identity) of your cluster in base64 format

You can read more about these values and their meaning in the guide on [cluster identity](keys.md#cluster-identity).

The *master secret* must be kept secret and can be used to [recover your cluster](../workflows/recovery.md).
Instead of managing this secret manually, you can [use your key management solution of choice](keys.md#user-managed-key-management) with Constellation.

The *clusterID* uniquely identifies a cluster and can be used to [verify your cluster](../workflows/verify.md).

## Upgrades

Constellation images and components might need to be upgraded to new versions during the lifetime of a cluster.
Constellation implements a rolling update mechanism ensuring no downtime of the control or data plane.
You can upgrade a Constellation cluster with a single operation by using the CLI.
For step-by-step instructions on how to do this, refer to [Upgrade Constellation](../workflows/upgrade.md).

### Attestation of upgrades

The new verification hashes (measurements) are released together with every image.
During an update procedure, the CLI provides the new measurements to the [JoinService](components.md#joinservice) securely.
New measurements for an updated image are automatically pulled and verified by the CLI following the [supply chain security concept](attestation.md#chain-of-trust) of Constellation.
The [attestation section](attestation.md#cluster-facing-attestation) describes in detail how these measurements are then used by the JoinService for the attestation of nodes.

The updated measurements are reproducible based on the updated node images, hence,  auditable by the customer.
