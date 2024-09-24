# Constellation - Security Overview

The security of Constellation is based on a set of protocols. The protocols are outlined in the following. The following diagram sketches the basic trust relationships between the entities in a Constellation cluster.

![chain of trust](chain-of-trust.jpg)

## Software components

Abstractly, the Constellation software comprises three core components:
1. The command-line interface (CLI) executable, which is run by the cluster administrator on their local machine.
2. The node images, which are run inside Confidential VMs (CVMs). Each node image contains a Linux kernel and a user mode program called Bootstrapper.
3. The Constellation service containers, which are run on Kubernetes. There are three Constellation services: KeyService, JoinService, and VerificationService.

## CLI: root of trust

The CLI executable is signed by Edgeless Systems. To ensure non-repudiability for CLI releases, Edgeless Systems publishes corresponding signatures to the public ledger of the [Sigstore project](https://www.sigstore.dev/). There's a [step-by-step guide](https://docs.edgeless.systems/constellation/workflows/verify-cli) on how to verify CLI signatures based on Sigstore.

The CLI contains the [runtime measurements of the latest Constellation node image](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/attestation/measurements/measurements_enterprise.go#L21) for all supported cloud platforms. The CLI writes these measurements to a cluster's config file as part of the ["config generate" command](https://docs.edgeless.systems/constellation/workflows/config). 

In case a different version of the node image is to be used, the corresponding runtime measurements can be fetched using the CLI's ["config fetch-measurements" command](reference/cli#constellation-config-fetch-measurements). This command downloads the runtime measurements and the corresponding signature from Edgeless Systems from https://cdn.confidential.cloud. See for example the following files corresponding to node image v2.16.3:
* [Measurements](https://cdn.confidential.cloud/constellation/v2/ref/-/stream/stable/v2.16.3/image/measurements.json)
* [Signature](https://cdn.confidential.cloud/constellation/v2/ref/-/stream/stable/v2.16.3/image/measurements.json.sig)

In addition to the runtime measurements, the CLI contains the following data in hardcoded form:
* The [long-term public key of Edgeless Systems](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/constants/constants.go#L264) to verify the signature of downloaded runtime measurements.
* The [hashes of the to-be-installed Kubernetes binaries](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/versions/versions.go#L199).
* The [Helm charts used for the installation of the three Constellation services](https://github.com/edgelesssys/constellation/tree/main/internal/constellation/helm/charts/edgeless/constellation-services), which include hashes of the respective containers. *Note: The Helm charts and the hashes are [generated at build time](https://github.com/edgelesssys/constellation/blob/main/internal/constellation/helm/imageversion/imageversion.go). A future version of the CLI will provide a command to print the Helm charts.*

## Cluster creation

When a cluster is [created](https://docs.edgeless.systems/constellation/workflows/create), the CLI interacts with the API of the respective infrastructure provider (e.g., Azure) and launches CVMs with the applicable node image. These CVMs are called *nodes*. On each node, the Bootstrapper is launched.

The CLI automatically selects one of the nodes as *first node*. The CLI automatically verifies the first node's runtime measurements using remote attestation. Based on this, the CLI and the Bootstrapper running on the first node set up a temporary TLS connection. This [aTLS](https://docs.edgeless.systems/constellation/architecture/attestation#attested-tls-atls) connection is mainly used for three things (see the the [interface definition](https://github.com/edgelesssys/constellation/blob/main/bootstrapper/initproto/init.proto) for a comprehensive list of exchanged data):
1. The CLI sends the hashes of the to-be-installed Kubernetes binaries to the first node.
2. The CLI generates the  [master secret](../architecture/keys.md#master-secret) of the to-be-created cluster and sends it to the first node.
3. The first node generates a [kubeconfig file](https://www.redhat.com/sysadmin/kubeconfig) and sends it to the CLI. The kubeconfig file contains Kubernetes credentials for the CLI and the Kubernetes cluster's public key, among others.

After this, the aTLS connection is closed and the node is marked as "initialized". This marker is irrevocably reflected in the node's remote attestation. This mechanism prevents a node from joining different clusters. 

## Kubernetes bootstrapping on the first node

The Bootstrapper on the first node downloads and verifies the Kubernetes binaries, using the hashes it received from the CLI. These binaries include for example [kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/), [kube-apiserver](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/), and [etcd](https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/). With these, the Bootstrapper creates a single-node Kubernetes cluster. Etcd is a distributed key-value store that Kubernetes uses to store configuration data for services. The etcd agent runs on each control-plane node of a cluster. The agents use mTLS for communication between them. Etcd uses the Raft protocol (over mTLS) to distribute state between nodes.

Next, the CLI the connects to the Kubernetes API server (kube-apiserver) using the kubeconfig file it received from the Bootstrapper. This results in an mTLS connection between the CLI and the Kubernetes API server. The CLI uses this connection for two essential operations at the Kubernetes level:
1. It writes the runtime measurements of the applicable node image to a specific key in the cluster's etcd store.
2. It executes the [hardcoded Helm charts](#cli-root-of-trust) of the three Constellation services KeyService, JoinService, and VerificationService.

The latter causes the first node to download, verify, and run these three services. The services' containers are hosted at https://ghcr.io/edgelesssys.

After this, the Constellation cluster is operational on the first node.

## Cluster growth

Additional nodes can now join the cluster - either as control-plane nodes or as worker nodes. For both, the process for joining the cluster is identical. First, the Bootstrapper running on a *new node* contacts the JoinService of the existing cluster. The JoinService verifies the remote-attestation statement of the new node using the runtime measurements stored in etcd. On success, an aTLS connection between the two is created, which is used mainly for the following (see the the [interface definition](https://github.com/edgelesssys/constellation/blob/main/joinservice/joinproto/join.proto) for a comprehensive list of exchanged data):
1. The new node sends a certificate signing request (CSR) to the JoinService.
1. The JoinService issues a corresponding certificate and sends it to the new node. The JoinService uses the signing key of the Kubernetes cluster for this, which is [generated by kubeadm](https://kubernetes.io/docs/setup/best-practices/certificates/).
1. The JoinService sends a [kubeadm token](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-token/) to the new node.
1. The JoinService sends the hashes of the to-be-installed Kubernetes binaries to the new node.
1. The JoinService sends encryption key for the new node's local storage. This key is generated and managed by the cluster's KeyService.
1. The JoinService sends the certificate of the cluster's control plane to the new node.

After this, the aTLS connection is closed and the node is marked as "initialized". 

The Bootstrapper on the new node uses the kubeadm token to download the configuration of the cluster from the Kubernetes API server. Subsequently, it downloads, verifies, and runs the given Kubernetes binaries accordingly. The kubeadm token is never used after this. 

The kubelet on the new node uses its own certificate and the certificate of the cluster's control plane (which the new node both received from the JoinService) to establish an mTLS connection with the cluster's Kubernetes API server. Once connected, the new node registers itself as control-plane node or worker node of the cluster. This process uses the standard Kubernetes mechanisms for adding nodes. 

In Constellation, a virtual private network (VPN) exists between all nodes of a cluster. To join this VPN, the new node generates WireGuard credentials for itself and writes the public part to etcd via the mTLS connection with the Kubernetes API server. Whenever nodes within a Constellation cluster are talking to each other for the first time, they check etcd for the other node's public WireGuard credentials.

*Note that etcd communication between nodes is an exception: This traffic always goes via mTLS based on the node certificates issued by the cluster.*

## Cluster upgrade

Whenever a cluster is [upgraded](https://docs.edgeless.systems/constellation/workflows/upgrade) to a new version of the node image, the CLI sends the corresponding runtime measurements via the Kubernetes API server. The new runtime measurements are stored in etcd within the cluster and replace any previous runtime measurements. The new runtime measurements are then used automatically by the JoinServer for the verification of new nodes. 
