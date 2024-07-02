# Encrypted persistent storage

Confidential VMs provide runtime memory encryption to protect data in use.
In the context of Kubernetes, this is sufficient for the confidentiality and integrity of stateless services.
Consider a front-end web server, for example, that keeps all connection information cached in main memory.
No sensitive data is ever written to an insecure medium.
However, many real-world applications need some form of state or data-lake service that's connected to a persistent storage device and requires encryption at rest.
As described in [Use persistent storage](../workflows/storage.md), cloud service providers (CSPs) use the container storage interface (CSI) to make their storage solutions available to Kubernetes workloads.
These CSI storage solutions often support some sort of encryption.
For example, Google Cloud [encrypts data at rest by default](https://cloud.google.com/security/encryption/default-encryption), without any action required by the customer.

## Cloud provider-managed encryption

CSP-managed storage solutions encrypt the data in the cloud backend before writing it physically to disk.
In the context of confidential computing and Constellation, the CSP and its managed services aren't trusted.
Hence, cloud provider-managed encryption protects your data from offline hardware access to physical storage devices.
It doesn't protect it from anyone with infrastructure-level access to the storage backend or a malicious insider in the cloud platform.
Even with "bring your own key" or similar concepts, the CSP performs the encryption process with access to the keys and plaintext data.

In the security model of Constellation, securing persistent storage and thereby data at rest requires that all cryptographic operations are performed inside a trusted execution environment.
Consequently, using CSP-managed encryption of persistent storage usually isn't an option.

## Constellation-managed encryption

Constellation provides CSI drivers for storage solutions in all major clouds with built-in encryption support.
Block storage provisioned by the CSP is [mapped](https://guix.gnu.org/manual/en/html_node/Mapped-Devices.html) using the [dm-crypt](https://www.kernel.org/doc/html/latest/admin-guide/device-mapper/dm-crypt.html), and optionally the [dm-integrity](https://www.kernel.org/doc/html/latest/admin-guide/device-mapper/dm-integrity.html), kernel modules, before it's formatted and accessed by the Kubernetes workloads.
All cryptographic operations happen inside the trusted environment of the confidential Constellation node.

Note that for integrity-protected disks, [volume expansion](https://kubernetes.io/blog/2018/07/12/resizing-persistent-volumes-using-kubernetes/) isn't supported.

By default the driver uses data encryption keys (DEKs) issued by the Constellation [*KeyService*](microservices.md#keyservice).
The DEKs are in turn derived from the Constellation's key encryption key (KEK), which is directly derived from the [master secret](keys.md#master-secret).
This is the recommended mode of operation, and also requires the least amount of setup by the cluster administrator.

Alternatively, the driver can be configured to use a key management system to store and access KEKs and DEKs.

Refer to [keys and cryptography](keys.md) for more details on key management in Constellation.

Once deployed and configured, the CSI driver ensures transparent encryption and integrity of all persistent volumes provisioned via its storage class.
Data at rest is secured without any additional actions required by the developer.

## Cryptographic algorithms

This section gives an overview of the libraries, cryptographic algorithms, and their configurations, used in Constellation's CSI drivers.

### dm-crypt

To interact with the dm-crypt kernel module, Constellation uses [libcryptsetup](https://gitlab.com/cryptsetup/cryptsetup/).
New devices are formatted as [LUKS2](https://gitlab.com/cryptsetup/LUKS2-docs/-/tree/master) partitions with a sector size of 4096 bytes.
The used key derivation function is [Argon2id](https://datatracker.ietf.org/doc/html/rfc9106) with the [recommended parameters for memory-constrained environments](https://datatracker.ietf.org/doc/html/rfc9106#section-7.4) of 3 iterations and 64 MiB of memory, utilizing 4 parallel threads.
For encryption Constellation uses AES in XTS-Plain64. The key size is 512 bit.

### dm-integrity

To interact with the dm-integrity kernel module, Constellation uses [libcryptsetup](https://gitlab.com/cryptsetup/cryptsetup/).
When enabled, the used data integrity algorithm is [HMAC](https://datatracker.ietf.org/doc/html/rfc2104) with SHA256 as the hash function.
The tag size is 32 Bytes.

## Encrypted S3 object storage

Constellation comes with a service that you can use to transparently retrofit client-side encryption to existing applications that use S3 (AWS or compatible) for storage.
To learn more, check out the [s3proxy documentation](../workflows/s3proxy.md).
