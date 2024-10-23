# Key concepts

Constellation is a cloud-based confidential orchestration platform.
The foundation of Constellation is Kubernetes and therefore shares the same technology stack and architecture principles.
To learn more about Constellation and Kubernetes, see [product overview](../../overview/product.md).

## Orchestration and updates

As a cluster administrator, you can use the [Constellation CLI](orchestration.md) to install and deploy a cluster.
Updates are provided in accordance with the [support policy](versions.md).

## Microservices and attestation

Constellation manages the nodes and network in your cluster. All nodes are bootstrapped by the [_Bootstrapper_](microservices.md#bootstrapper). They're verified and authenticated by the [_JoinService_](microservices.md#joinservice) before being added to the cluster and the network. Finally, the entire cluster can be verified via the [_VerificationService_](microservices.md#verificationservice) using [remote attestation](attestation.md).

## Node images and verified boot

Constellation comes with operating system images for Kubernetes control-plane and worker nodes.
They're highly optimized for running containerized workloads and specifically prepared for running inside confidential VMs.
You can learn more about [the images](images.md) and how verified boot ensures their integrity during boot and beyond.

## Key management and cryptographic primitives

Encryption of data at-rest, in-transit, and in-use is the fundamental building block for confidential computing and Constellation. Learn more about the [keys and cryptographic primitives](keys.md) used in Constellation, [encrypted persistent storage](encrypted-storage.md), and [network encryption](networking.md).

## Observability

Observability in Kubernetes refers to the capability to troubleshoot issues using telemetry signals such as logs, metrics, and traces.
In the realm of Confidential Computing, it's crucial that observability aligns with confidentiality, necessitating careful implementation.
Learn more about the [observability capabilities in Constellation](./observability.md).
