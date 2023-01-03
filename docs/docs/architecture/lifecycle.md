# Cluster lifecycle

The lifecycle of a Constellation cluster consist of three phases: *creation*, *upgrade*, and *termination*. The following describes each phase and links to detailed descriptions of key components and concepts.

## Cluster creation

(**FS: this is an intro for everyone. Details on attestation etc. will be given on specialized pages.**)

The cluster administrator uses the [`constellation create`](../reference/cli.md#constellation-create) command to create a Constellation cluster. The process is as follows:

1. The CLI (i.e., the `constellation` program) uses the cloud provider's API to create the initial set of Confidential VMs (CVMs). It writes TODO
2. Each CVM boots the [node image](images.md) configured in `constellation-conf.yaml`. 
3. On each CVM, the [Bootstrapper](components.md#bootstrapper) is automatically launched. The Bootstrapper waits until it either receives an initialization request from the CLI or discovers an initialized cluster.

The cluster administrator then uses the [`constellation init`](../reference/cli.md#constellation-init) command to initialize the cluster. This triggers the following steps:

1. The CLI sets up an [aTLS](./attestation.md) ("attested TLS") connection with the Bootstrapper on one of the previously created CVMs. During the aTLS handshake, the CLI verifies the CVM (and the software running in it) based on the policy specified in `constellation-conf.yaml`.
2. The CLI sends `constellation-conf.yaml` to the Bootstrapper over the aTLS connection and triggers the initialization of Kubernetes.
3. The Bootstrapper downloads the official Kubernetes release specified in `constellation-conf.yaml` and the [Constellation services](./services.md). The Bootstrapper verifies the integrity of the downloaded  binaries (based on their SHA-256 hashes. For this, the Bootstrapper contains a hardcoded list of the SHA-256 hashes of all supported binaries). (**FS: Probably move explanation somewhere else.**)
4. The Bootstrapper initializes the Kubernetes cluster and launches the Constellation services. 

The Kubernetes cluster is now operational and consists of a single CVM-based node. Additional nodes automatically join the cluster as is described in the section on [cluster scaling](#cluster-scaling).

## Cluster scaling and failover

New nodes automatically boot the configured [node image](images.md) and 

7. The Bootstrappers of the other nodes discover the initialized cluster via the cloud provider's API. 
8. Each node's Bootstraper sends a request to join the newly created cluster to the JoinService.
8. As part of the join request each node includes an attestation statement of its boot measurements as authentication
9. The *JoinService* verifies the attestation statements and joins the nodes to the Kubernetes cluster
10. This process is repeated for every node joining the cluster later (e.g., through autoscaling)

## Cluster upgrade

## Cluster termination

## Post-installation configuration

Post installation, the CLI provides a configuration for [accessing the cluster using the Kubernetes API](https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/).
The `kubeconfig` file provides the credentials and configuration for connecting and authenticating to the API server.
Once configured, you can orchestrate the Kubernetes cluster via `kubectl`.

Make sure to keep the state files such as `terraform.tfstate` in the workspace directory to be able to manage your cluster later on.
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
