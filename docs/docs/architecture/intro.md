# Architecture

This section of the documentation offers a comprehensive overview of Constellation's inner workings. It details the chain of trust between various components and how they work together to ensure robust protection for your workloads. The main chapters include:

- [**Protocol Overview**](./overview.md): The recommended **starting point** for exploring the architecture. This chapter gives an overview of Constellation's architecture and explains the security protocol that underpins confidentiality and strong protection for your workloads.
- [**Key Components**](./components/cli.md): This chapter outlines Constellation's key components, their purposes, and how users interact with them:
  - The [CLI](./components/cli.md)is used to create and orchestrate your cluster.
  - Constellation's [core services](./components/microservices.md) run on control planes to ensure secure protocols for cluster expansion and integrity checks.
  - Constellation provides [operating system images](./components/node-images.md) for Kubernetes control-plane and worker nodes, optimized for containerized workloads and prepared for confidential VMs.
- [**Protection Mechanisms**](./security/attestation.md): A deeper dive into the various concepts that deliver strong protection guarantees for your Kubernetes clusters.

- [**Observability**](./observability.md): In a Kubernetes context, observability is crucial for efficiently identifying and resolving issues. This chapter covers Constellation's observability capabilities.

- [**Versions**](./versions.md): An overview of Constellation's versions and support policy.
