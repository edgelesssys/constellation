# Key management and cryptographic primitives

Constellation protects and isolates your cluster and workloads.
To that end, cryptography is the foundation that ensures the confidentiality and integrity of all components.
Evaluating the security and compliance of Constellation requires a precise understanding of the cryptographic primitives and keys used.
The following gives an overview of the architecture and explains the technical details.

## Confidential VMs

Confidential VM (CVM) technology comes with hardware and software components for memory encryption, isolation, and remote attestation.
For details on the implementations and cryptographic soundness, refer to the hardware vendors' documentation and advisories.

## Master secret

The master secret is the cryptographic material used for deriving the [*clusterID*](#cluster-identity) and the *key encryption key (KEK)* for [storage encryption](#storage-encryption).
It's generated during the bootstrapping of a Constellation cluster.
It can either be managed by [Constellation](#constellation-managed-key-management) or an [external key management system](#user-managed-key-management).
In case of [recovery](#recovery-and-migration), the master secret allows to decrypt the state and recover a Constellation cluster.

## Cluster identity

The identity of a Constellation cluster is represented by cryptographic [measurements](attestation.md#runtime-measurements):

The **base measurements** represent the identity of a valid, uninitialized Constellation node.
They depend on the node image, but are otherwise the same for every Constellation cluster.
On node boot, they're determined using the CVM's attestation mechanism and [measured boot up to the read-only root filesystem](images.md).

The **clusterID** represents the identity of a single initialized Constellation cluster.
It's derived from the master secret and a cryptographically random salt and unique for every Constellation cluster.
The [Bootstrapper](microservices.md#bootstrapper) measures the *clusterID* into its own PCR before executing any code not measured as part of the *base measurements*.
See [Node attestation](attestation.md#node-attestation) for details.

The remote attestation statement of a Constellation cluster combines the *base measurements* and the *clusterID* for a verifiable, unspoofable, unique identity.

## Network encryption

Constellation encrypts all cluster network communication using the [container network interface (CNI)](https://github.com/containernetworking/cni).
See [network encryption](networking.md) for more details.

The Cilium agent running on each node establishes a secure [WireGuard](https://www.wireguard.com/) tunnel between it and all other known nodes in the cluster.
Each node creates its own [Curve25519](http://cr.yp.to/ecdh.html) encryption key pair and distributes its public key via Kubernetes.
A node uses another node's public key to decrypt and encrypt traffic from and to Cilium-managed endpoints running on that node.
Connections are always encrypted peer-to-peer using [ChaCha20](http://cr.yp.to/chacha.html) with [Poly1305](http://cr.yp.to/mac.html).
WireGuard implements [forward secrecy with key rotation every 2 minutes](https://lists.zx2c4.com/pipermail/wireguard/2017-December/002141.html).
Cilium supports [key rotation](https://docs.cilium.io/en/stable/security/network/encryption-ipsec/#key-rotation) for the long-term node keys via Kubernetes secrets.

## Storage encryption

Constellation supports transparent encryption of persistent storage.
The Linux kernel's device mapper-based encryption features are used to encrypt the data on the block storage level.
Currently, the following primitives are used for block storage encryption:

* [dm-crypt](https://www.kernel.org/doc/html/latest/admin-guide/device-mapper/dm-crypt.html)
* [dm-integrity](https://www.kernel.org/doc/html/latest/admin-guide/device-mapper/dm-integrity.html)

Adding primitives for integrity protection in the CVM attacker model are under active development and will be available in a future version of Constellation.
See [encrypted storage](encrypted-storage.md) for more details.

As a cluster administrator, when creating a cluster, you can use the Constellation [installation program](orchestration.md) to select one of the following methods for key management:

* Constellation-managed key management
* User-managed key management

### Constellation-managed key management

#### Key material and key derivation

During the creation of a Constellation cluster, the cluster's master secret is used to derive a KEK.
This means creating two clusters with the same master secret will yield the same KEK.
Any data encryption key (DEK) is derived from the KEK via HKDF.
Note that the master secret is recommended to be unique for every cluster and shouldn't be reused (except in case of [recovering](../workflows/recovery.md) a cluster).

#### State and storage

The KEK is derived from the master secret during the initialization.
Subsequently, all other key material is derived from the KEK.
Given the same KEK, any DEK can be derived deterministically from a given identifier.
Hence, there is no need to store DEKs. They can be derived on demand.
After the KEK was derived, it's stored in memory only and never leaves the CVM context.

#### Availability

Constellation-managed key management has the same availability as the underlying Kubernetes cluster.
Therefore, the KEK is stored in the [distributed Kubernetes etcd storage](https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/) to allow for unexpected but non-fatal (control-plane) node failure.
The etcd storage is backed by the encrypted and integrity protected [state disk](images.md#state-disk) of the nodes.

#### Recovery

Constellation clusters can be recovered in the event of a disaster, even when all node machines have been stopped and need to be rebooted.
For details on the process see the [recovery workflow](../workflows/recovery.md).

### User-managed key management

User-managed key management is under active development and will be available soon.
In scenarios where constellation-managed key management isn't an option, this mode allows you to keep full control of your keys.
For example, compliance requirements may force you to keep your KEKs in an on-prem key management system (KMS).

During the creation of a Constellation cluster, you specify a KEK present in a remote KMS.
This follows the common scheme of "bring your own key" (BYOK).
Constellation will support several KMSs for managing the storage and access of your KEK.
Initially, it will support the following KMSs:

* [AWS KMS](https://aws.amazon.com/kms/)
* [GCP KMS](https://cloud.google.com/security-key-management)
* [Azure Key Vault](https://azure.microsoft.com/en-us/services/key-vault/#product-overview)
* [KMIP-compatible KMS](https://www.oasis-open.org/committees/tc_home.php?wg_abbrev=kmip)

Storing the keys in Cloud KMS of AWS, Azure, or GCP binds the key usage to the particular cloud identity access management (IAM).
In the future, Constellation will support remote attestation-based access policies for Cloud KMS once available.
Note that using a Cloud KMS limits the isolation and protection to the guarantees of the particular offering.

KMIP support allows you to use your KMIP-compatible on-prem KMS and keep full control over your keys.
This follows the common scheme of "hold your own key" (HYOK).

The KEK is used to encrypt per-data "data encryption keys" (DEKs).
DEKs are generated to encrypt your data before storing it on persistent storage.
After being encrypted by the KEK, the DEKs are stored on dedicated cloud storage for persistence.
Currently, Constellation supports the following cloud storage options:

* [AWS S3](https://aws.amazon.com/s3/)
* [GCP Cloud Storage](https://cloud.google.com/storage)
* [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/#overview)

The DEKs are only present in plaintext form in the encrypted main memory of the CVMs.
Similarly, the cryptographic operations for encrypting data before writing it to persistent storage are performed in the context of the CVMs.

#### Recovery and migration

In the case of a disaster, the KEK can be used to decrypt the DEKs locally and subsequently use them to decrypt and retrieve the data.
In case of migration, configuring the same KEK will provide seamless migration of data.
Thus, only the DEK storage needs to be transferred to the new cluster alongside the encrypted data for seamless migration.
