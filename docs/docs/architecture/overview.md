# Overview

**FS: OK'ish but not great. Do we need this section at all? Probably not.**
Constellation is a cloud-based confidential orchestration platform.
The foundation of Constellation is Kubernetes and therefore shares the same technology stack and architecture principles.
To learn more about Constellation and Kubernetes, see [product overview](../overview/product.md).

## About orchestration and updates

**FS: this is more like How-To**
As a cluster administrator, you can use the [Constellation CLI](orchestration.md) to install and deploy a cluster.
Updates are provided in accordance with the [support policy](versions.md).

## About the components and attestation

Constellation manages the nodes and network in your cluster. All nodes are bootstrapped by the [*Bootstrapper*](components.md#bootstrapper). They're verified and authenticated by the [*JoinService*](components.md#joinservice) before being added to the cluster and the network. Finally, the entire cluster can be verified via the [*VerificationService*](components.md#verificationservice) using [remote attestation](attestation.md).

## About node images and verified boot

Constellation comes with operating system images for Kubernetes control-plane and worker nodes.
They're highly optimized for running containerized workloads and specifically prepared for running inside confidential VMs.
You can learn more about [the images](images.md) and how verified boot ensures their integrity during boot and beyond.

## About key management and cryptographic primitives

Encryption of data at-rest, in-transit, and in-use is the fundamental building block for confidential computing and Constellation. Learn more about the [keys and cryptographic primitives](keys.md) used in Constellation, [encrypted persistent storage](encrypted-storage.md), and [network encryption](networking.md).
