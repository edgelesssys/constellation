# Constellation - Security Overview

The security of Constellation is based on a set of protocols. The protocols are outlined in the following. The following diagram sketches the basic trust relationships between the entities in a Constellation cluster.

![chain of trust](chain-of-trust.jpg)

## Software components

Abstractly, the Constellation software comprises three core components:
1. The command-line interface (CLI) executable, which is run by the cluster administrator on their local machine.
2. The node images, which are run inside Confidential VMs (CVMs). Among other, each node image contains a Linux kernel and a user mode program called Bootstrapper.
3. The Constellation service containers, which are run on Kubernetes. There are three Constellation services: KeyService, JoinService, and VerificationService.

## CLI: root of trust

The CLI executable is signed by Edgeless Systems. To ensure non-repudiability for CLI releases, Edgeless Systems publishes corresponding signatures to the public ledger of the [Sigstore project](https://www.sigstore.dev/). There's a [step-by-step guide](https://docs.edgeless.systems/constellation/workflows/verify-cli) on how to verify CLI signatures based on Sigstore.

The CLI contains the [measurements](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/attestation/measurements/measurements_enterprise.go#L21) of the latest Constellation node image for all supported cloud platforms. The CLI writes these measurements to a cluster's config file as part of the ["config generate" command](https://docs.edgeless.systems/constellation/workflows/config). Note that currently Constellation uses 16 TPM-based runtime measurements for each cloud platform. The purpose and source of the measurements are described in the [next section](#remote-attestation-of-nodes).

In case a different version of the node image is to be used, the corresponding measurements can be fetched using the CLI's ["config fetch-measurements" command](reference/cli#constellation-config-fetch-measurements). This command downloads the measurements and the corresponding signature from Edgeless Systems from https://cdn.confidential.cloud. See for example the following files corresponding to node image v2.16.3:
* [Measurements](https://cdn.confidential.cloud/constellation/v2/ref/-/stream/stable/v2.16.3/image/measurements.json)
* [Signature](https://cdn.confidential.cloud/constellation/v2/ref/-/stream/stable/v2.16.3/image/measurements.json.sig)

In addition to the measurements, the CLI contains the following data in hardcoded form:
* The [long-term public key of Edgeless Systems](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/constants/constants.go#L264) to verify the signature of downloaded measurements.
* The [hashes of the to-be-installed Kubernetes binaries](https://github.com/edgelesssys/constellation/blob/edc0c7068ef4527efeaf584a2a35e0f51f58c426/internal/versions/versions.go#L199).
* The [Helm charts used for the installation of the three Constellation services](https://github.com/edgelesssys/constellation/tree/main/internal/constellation/helm/charts/edgeless/constellation-services), which include hashes of the respective containers. *Note: The Helm charts and the hashes are [generated at build time](https://github.com/edgelesssys/constellation/blob/main/internal/constellation/helm/imageversion/imageversion.go). A future version of the CLI will provide a command to print the Helm charts.*

## Remote attestation of nodes

To identify themselves, nodes use the remote-attestation functionality of the underlying CVM platform. Constellation supports Intel TDX and AMD SEV-SNP based platforms. Abstractly, Intel TDX and AMD SEV-SNP hash the initial memory contents of the CVMs. This hash is also called `launch digest`. The launch digest is reflected in each remote-attestation statement that is requested by the software inside the CVM. Abstractly, a remote-attestation statement `R` from a CVM looks as follows: 

```
R = Sig-CPU(<launch digest>, <auxiliary data>, <payload>)
```

The `payload` is controlled by the software running inside the CVM. In the case of a Constellation node, the `payload` is always the public key of the respective Bootstrapper running inside the CVM. Thus, `R` can be seen as a certificate for that public key issued by the CPU. Based on this, nodes establish so called "attested TLS" (aTLS) connections. aTLS is used during [cluster creation](#cluster-creation) and when [growing a cluster](#cluster-growth).

### Measurements

In the ideal case, the underlying CVM platform does not inject any of its own software into a CVM. In that case, a Constellation node image can contain its own firmware/UEFI. This allows for the creation of node images for which the launch digest covers all defining parts of a node, including the firmware, the kernel, the kernel command line, and the disk image. In this case, the launch digest is the only measurement that's required to verify the identity and integrity of a node. 

### Measured Boot as fallback

However, currently, all supported CVM platforms (AWS, Azure, and GCP) inject custom firmware into CVMs. Thus, in practice, Constellation relies on conventional [measured boot](https://docs.edgeless.systems/constellation/architecture/images#measured-boot) to reflect the identity and integrity of nodes. In measured boot, in general, the software components involved in the boot process of a system are "measured" into the 16 registers of a Trusted Platform Module (TPM). The values of theses registers are also called "runtime measurements".

All supported CVM platforms provide TPMs to CVMs. Constellation nodes use these to measure their boot process. They include the 16 runtime measurements as `auxiliary data` in `R`. On each CVM platform, runtime measurements are taken differently. Details on this are given in the [Constellation documentation](https://docs.edgeless.systems/constellation/architecture/attestation#runtime-measurements). 

With measured boot, Constellation only checks the 16 runtime measurements during the verification of a node's remote-attestation statement. The launch digest is not considered, because it only covers the firmware injected by the CVM platform and may change whenever the CVM platform is updated. 

Currently, on AWS and GCP the TPM implementation resides outside the CVM. On Azure, the TPM implementation is part of the injected firmware and resides inside the CVM. More information can be found in the [Constellation documentation](https://docs.edgeless.systems/constellation/overview/clouds).

## Cluster creation

When a cluster is [created](https://docs.edgeless.systems/constellation/workflows/create), the CLI interacts with the API of the respective infrastructure provider (e.g., Azure) and launches CVMs with the applicable node image. These CVMs are called *nodes*. On each node, the Bootstrapper is launched.

The CLI automatically selects one of the nodes as *first node*. The CLI automatically verifies the first node's measurements using [remote attestation](#remote-attestation-and-runtime-measurements). Based on this, the CLI and the Bootstrapper running on the first node set up a temporary aTLS connection. This connection is mainly used for three things (see the the [interface definition](https://github.com/edgelesssys/constellation/blob/main/bootstrapper/initproto/init.proto) for a comprehensive list of exchanged data):
1. The CLI sends the hashes of the to-be-installed Kubernetes binaries to the first node.
2. The CLI generates the  [master secret](../architecture/keys.md#master-secret) of the to-be-created cluster and sends it to the first node.
3. The first node generates a [kubeconfig file](https://www.redhat.com/sysadmin/kubeconfig) and sends it to the CLI. The kubeconfig file contains Kubernetes credentials for the CLI and the Kubernetes cluster's public key, among others.

After this, the aTLS connection is closed and the node is marked as "initialized" by extending. This marker is irrevocably reflected in the node's remote attestation. This mechanism prevents a node from (accidentally or maliciously) joining different clusters. 

## Kubernetes bootstrapping on the first node

The Bootstrapper on the first node downloads and verifies the Kubernetes binaries, using the hashes it received from the CLI. These binaries include for example [kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/), [kube-apiserver](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/), and [etcd](https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/). With these, the Bootstrapper creates a single-node Kubernetes cluster. Etcd is a distributed key-value store that Kubernetes uses to store configuration data for services. The etcd agent runs on each control-plane node of a cluster. The agents use mTLS for communication between them. Etcd uses the Raft protocol (over mTLS) to distribute state between nodes.

Next, the CLI the connects to the Kubernetes API server (kube-apiserver) using the kubeconfig file it received from the Bootstrapper. This results in an mTLS connection between the CLI and the Kubernetes API server. The CLI uses this connection for two essential operations at the Kubernetes level:
1. It writes the measurements of the applicable node image to a specific key in the cluster's etcd store.
2. It executes the [hardcoded Helm charts](#cli-root-of-trust) of the three Constellation services KeyService, JoinService, and VerificationService.

The latter causes the first node to download, verify, and run these three services. The services' containers are hosted at https://ghcr.io/edgelesssys.

After this, the Constellation cluster is operational on the first node.

## Cluster growth

Additional nodes can now join the cluster - either as control-plane nodes or as worker nodes. For both, the process for joining the cluster is identical. First, the Bootstrapper running on a *new node* contacts the JoinService of the existing cluster. The JoinService verifies the remote-attestation statement of the new node using the measurements stored in etcd. On success, an aTLS connection between the two is created, which is used mainly for the following (see the the [interface definition](https://github.com/edgelesssys/constellation/blob/main/joinservice/joinproto/join.proto) for a comprehensive list of exchanged data):
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

Whenever a cluster is [upgraded](https://docs.edgeless.systems/constellation/workflows/upgrade) to a new version of the node image, the CLI sends the corresponding measurements via the Kubernetes API server. The new measurements are stored in etcd within the cluster and replace any previous measurements. The new measurements are then used automatically by the JoinServer for the verification of new nodes. 
