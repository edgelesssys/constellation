# Architecture

This section offers a comprehensive overview of Constellation's inner workings. It details the chain of trust between various components and how they work together to ensure robust protection for your workloads. The main chapters include:

- [**Protocol overview**](./overview.md): The recommended **starting point** for exploring the architecture. This chapter overviews Constellation's architecture and explains the security protocol ensuring confidentiality and strong protection for your workloads.

- [**Key components**](./components/cli.md): This chapter outlines Constellation's main components, their roles, and how users interact with them:

  - [The CLI](./components/cli.md): A command-line tool to efficiently create and manage your cluster.
  - [Constellation's core services](./components/microservices.md): These services run on the control planes, enabling secure protocols for cluster scaling and performing integrity checks.
  - [Operating system images](./components/node-images.md): Constellation offers optimized OS images for Kubernetes control-plane and worker nodes, tailored for containerized workloads and ready for confidential VMs.

- [**Security concept**](./security/attestation.md): A detailed exploration of the concepts that provide strong protection for your Kubernetes clusters, including:

  - [Attestation](./security/attestation.md): Describes the process of verifying that your workloads are operating in a secure and protected state.
  - [Encrypted networking](./security/encrypted-networking.md): Explains how Constellation ensures strong encryption for all cluster traffic.
  - [Encrypted persistent storage](./security/encrypted-storage.md): Covers Constellation's approach to securely handling data in persistent storage.
  - [Cryptographic keys and primitives](./security/keys.md): Provides an overview of how Constellation manages cryptographic keys and the primitives used to protect workloads and data.

- [**Observability**](./observability.md): Observability is essential for identifying and resolving issues efficiently in a Kubernetes environment. This chapter highlights Constellation's observability features and capabilities.

- [**Versions**](./versions.md): A comprehensive overview of Constellation's versions and support policy.
