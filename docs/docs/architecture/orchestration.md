# Orchestrating Constellation clusters

You can use the CLI to create a cluster on the supported cloud platforms.
The CLI provisions the resources in your cloud environment and initiates the initialization of your cluster.
It uses a set of parameters and an optional configuration file to manage your cluster installation.
The CLI is also used for updating your cluster.

## Workspaces

Each Constellation cluster has an associated *workspace*.
The workspace is where data such as the Constellation state, config, and ID files are stored.
Each workspace is associated with a single cluster and configuration.
The CLI stores state in the local filesystem making the current directory the active workspace.
Multiple clusters require multiple workspaces, hence, multiple directories.
Note that every operation on a cluster always has to be performed from the directory associated with its workspace.

## Cluster creation process

To allow for fine-grained configuration of your cluster and cloud environment, Constellation supports an extensive configuration file with strong defaults. [Generating the configuration file](../workflows/create.md#configuration) is typically the first thing you do in the workspace.

Altogether, the following files are generated during the creation of a Constellation cluster and stored in the current workspace:

* a configuration file
* an ID file
* a Base64-encoded master secret
* Terraform artifacts such as `terraform.tfstate`, stored in the `constellation-terraform` subdirectory
* a Kubernetes `kubeconfig` file.

Constellation uses Terraform for infrastructure management. No setup of Terraform is needed. The CLI automatically fetches a copy of Terraform when required. The policy on usage of Terraform by Constellation can be found [here](../reference/terraform.md).

After the creation of your cluster, the CLI will provide you with a Kubernetes `kubeconfig` file.
This file grants you access to your Kubernetes cluster and configures the [kubectl](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) tool.
In addition, the cluster's [identifier](orchestration.md#post-installation-configuration) is returned and stored in a file called `constellation-id.json`

### Creation process details

1. The CLI `create` command creates the confidential VM (CVM) resources in your cloud environment and configures the network
2. Each CVM boots the Constellation node image and measures every component in the boot chain
3. The first component launched in each node is the [*Bootstrapper*](components.md#bootstrapper)
4. The *Bootstrapper* waits until it either receives an initialization request or discovers an initialized cluster
5. The CLI `init` command connects to the *Bootstrapper* of a selected node, sends the configuration, and initiates the initialization of the cluster
6. The *Bootstrapper* of **that** node [initializes the Kubernetes cluster](components.md#bootstrapper) and deploys the other Constellation [components](components.md) including the [*JoinService*](components.md#joinservice)
7. Subsequently, the *Bootstrappers* of the other nodes discover the initialized cluster and send join requests to the *JoinService*
8. As part of the join request each node includes an attestation statement of its boot measurements as authentication
9. The *JoinService* verifies the attestation statements and joins the nodes to the Kubernetes cluster
10. This process is repeated for every node joining the cluster later (e.g., through autoscaling)

## Post-installation configuration

Post-installation the CLI provides a configuration for [accessing the cluster using the Kubernetes API](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/).
The `kubeconfig` file provides the credentials and configuration for connecting and authenticating to the API server.
Once configured, orchestrate the Kubernetes cluster via `kubectl`.

Make sure to keep the Terraform subdirectory such as `constellation-terraform` in the workspace directory to be able to manage your cluster later on.
Without it, you won't be able to modify or terminate your cluster.

After the initialization, the CLI will present you with a couple of tokens:

* The [*master secret*](keys.md#master-secret) (stored in the `constellation-mastersecret.json` file by default)
* The [*clusterID*](keys.md#cluster-identity) of your cluster in Base64 encoding

You can read more about these values and their meaning in the guide on [cluster identity](keys.md#cluster-identity).

The *master secret* must be kept secret and can be used to [recover your cluster](../workflows/recovery.md).
Instead of managing this secret manually, you can [use your key management solution of choice](keys.md#user-managed-key-management) with Constellation.

The *clusterID* uniquely identifies a cluster and can be used to [verify your cluster](../workflows/verify-cluster.md).

## Upgrades

Constellation images and components may need to be upgraded to new versions during the lifetime of a cluster.
Constellation implements a rolling update mechanism ensuring no downtime of the control or data plane.
You can upgrade a Constellation cluster with a single operation by using the CLI.
For step-by-step instructions on how to do this, refer to [Upgrade your cluster](../workflows/upgrade.md).

### Attestation of upgrades

With every new image, corresponding measurements are released.
During an update procedure, the CLI provides the new measurements to the [JoinService](components.md#joinservice) securely.
New measurements for an updated image are automatically pulled and verified by the CLI following the [supply chain security concept](attestation.md#chain-of-trust) of Constellation.
The [attestation section](attestation.md#cluster-facing-attestation) describes in detail how these measurements are then used by the JoinService for the attestation of nodes.

<!-- soon: As the [builds of the Constellation images are reproducible](attestation.md#chain-of-trust), the updated measurements are auditable by the customer. -->
