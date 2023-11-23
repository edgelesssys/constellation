# RFC 005: External KMS

Currently, Constellation only supports [Constellation-managed key management](https://docs.edgeless.systems/constellation/2.0/architecture/keys#constellation-managed-key-management).
The owner provides a master secret to the cluster on initialization.
The cluster holds this secret during its lifetime and uses it to derive DEKs for disk encryption.

In addition, Constellation should support [user-managed key management](https://docs.edgeless.systems/constellation/2.0/architecture/keys#user-managed-key-management) with an external KMS.
Main goals are:

* The KEK (master secret) never leaves the KMS. Unencrypted DEKs only exist temporarily.
* Auto recovery of a cluster without manual intervention.

We need to decide on how to implement the KMIP client and what reference KMIP servers we want to implement against in the first iteration.

## Details on status quo

The CLI generates the master secret and passes it to the Bootstrapper on init.
The CLI also injects it into the Helm charts as a Kubernetes secret for the Constellation KMS service (referred to as CKMS in the following to avoid confusion).

The Bootstrapper uses the master secret to

* derive the *measurement salt* and the *clusterID*.
* derive the DEK and set up the state disk.

Note that the Bootstrapper performs these operations itself.
Particularly, it replicates the DEK derivation of the CKMS in ClusterKMS mode.

The CKMS serves a gRPC API to get DEKs.
It's designed to support different (cloud) KMS services as backend.
Currently, only the ClusterKMS backend is used, which uses the master secret from the Kubernetes secrets to derive the DEKs.
The clients of the CKMS are the join-service and the CSI drivers.

### Recovery

When a node boots, the disk-mapper will setup the encrypted state disk during initramfs.
To decrypt an already existing disks, the disk-mapper will ask all available joinservice instances for a decryption key for it's current disk UUID.
In case there is no join-service instance that can provide the decryption key, manual recovery becomes necessary.

To enable manual recovery the disk-mapper starts a recovery server.
The recovery server waits for a `recover` gRPC call.
During a `recover` call the CLI will attest the measurements of the node.
After successful attestation the CLI will provide a disk decryption key and measurement secret for the given disk UUID.
The measurement secret, together with a measurement salt (not secret) is used to derive the clusterID.

*Changes for eKMS; regarding disk decryption:*
* Recovery server accepts KMS URI, storage URI and kms/storage IAM secret instead of a masterSecret. During normal operation the KMS service has access to the IAM secrets through a mounted k8s secret. This secret is not available during initramfs.
* For eKMS backends the two URIs can be used directly to request new DEKs.
* For the cKMS backend the KMS URI can include an optional parameter that holds the masterSecret: `kms://cluster-kms?masterSecret=<masterSecret>`.

The above approach allows us to integrate with the existing setup code in `keyservice/setup/setup.go` with only minimal changes (parse masterSecret in case of cluster-kms).
This code is used to setup CloudKMS objects.
The `setup.go` code will have to be refactored to live in `internal` so that the disk-mapper pkg can directly communicate with the respective external KMS.

### Implemented, but yet unused features

There are CKMS backends for Azure, GCP, and AWS.
These should be working, but aren't battle-tested.

The init gRPC API has the following fields:

```
  bytes master_secret = 2;
  string kms_uri = 3;
  string storage_uri = 4;
  string key_encryption_key_id = 5;
  bool use_existing_kek = 6;
  bytes salt = 10;
```

* `master_secret` (and `salt`) are used as described above. The other fields are unused.
* `kms_uri` and `storage_uri` contain the type and configuration of the (external) KMS and storage to use. The CKMS already has logic to parse them and create corresponding backends.
* `key_encryption_key_id` and `use_existing_kek` are supposed to control how the external KMS is used.

## Overview of Hashicorp Vault's KMS integrations

Multiple features of Vault integrate with KMSs.

### KMS secrets engine

A KMS secrets engine uses an external KMS to perform operations like encryption and decryption.
Vault provides a [KMS engine for GCP](https://developer.hashicorp.com/vault/docs/secrets/gcpkms) only.

### Auto unseal

When you start Vault, it's in a sealed state.
You must provide the seal key to unseal it.
This is similar to recovering a Constellation cluster.

Instead of manually providing the key, you can configure Vault to [automatically use a KMS to decrypt its root key](https://developer.hashicorp.com/vault/docs/concepts/seal#auto-unseal).
You can use Azure, GCP, and AWS KMSs among others, but not KMIP.

Vault enterprise features [seal wrap](https://developer.hashicorp.com/vault/docs/enterprise/sealwrap).
This is an "extra layer of protection" that uses a KMS to encrypt single values before storing them in Vault.
The docs promote this as a means of meeting compliance requirements rather than improving security.

### KMIP server

Vault can act as a [KMIP server](https://developer.hashicorp.com/vault/docs/secrets/kmip).
This means that clients can connect to Vault via both the native API and a KMIP API.
Vault does *not* provide any KMIP client functionality.

### Go-KMS-Wrapping

https://github.com/hashicorp/go-kms-wrapping

This library wraps various KMS services and provides a common interface to them.
It's similar to the CKMS implementation.
We may consider to replace our own implementation with this.
However, it also doesn't have support for KMIP.

### Conclusion

We can't use Vault as a component of Constellation to add KMIP support.
We also can't use it as a cloud KMS proxy because it only supports GCP.

We could use Vault as DEK storage and use auto unseal with cloud KMS to replace the master secret.
However, we would need a separate way for KMIP support.

We should probably go with adding native support for external KMSs.

We may use Vault as a reference KMIP server to implement against, but we need to procure Vault Enterprise and the ADP module for it.

## Other tools and libraries

### PyKMIP

[PyKMIP](https://github.com/OpenKMIP/PyKMIP) is a KMIP client library for Python.
Part of the project is a [KMIP server](https://pykmip.readthedocs.io/en/latest/server.html) for testing.
We may use it as a reference KMIP server to implement against.

### kmip-go

[kmip-go](https://github.com/ThalesGroup/kmip-go) provides [KMIP protocol structures and can encode](https://pkg.go.dev/github.com/gemalto/kmip-go#example-package-Client) and decode them.
We can use this as a building block of the Constellation KMIP client.
It also has a [server type](https://pkg.go.dev/github.com/gemalto/kmip-go#Server) that may be useful in unit tests.

## Tasks

* Decide whether we want to replace our own implementations with Go-KMS-Wrapping
* Plan recovery with external KMS
* The code of the Bootstrapper (and recovery) dealing with the master secret should be centralized and shared with the CKMS implementation
  * Implement the variant that uses an external KMS instead of the master key for these functionalities
* Expose CKMS configuration to the user
* Implement KMIP as a backend for the CKMS
  * Use kmip-go for the client logic
  * Use kmip-go's server type for unit tests
  * Use PyKMIP's server for integration tests
  * Use Vault Enterprise+ADP for extended integration and/or e2e tests

## Issues

Can we achieve recovery of a cluster without manual intervention with KMIP?
How to authenticate?
