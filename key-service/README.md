# KeyService

The KeyService is one of Constellation's Kubernetes components, responsible for distributing keys and secrets to other services.
This includes the JoinService, which contacts the KeyService to derive state disk keys and measurement secrets for newly-joining, and rejoining nodes,
and Constellation's CSI drivers, which contact the KeyService for disk encryption keys.

The service is not exposed outside the cluster, and should be kept for internal usage only.

## gRPC API

Keys can be requested through simple gRPC API based on an ID and key length.

## Backends

The KeyService supports multiple backends to store keys and manage crypto operations.
The default option holds a master secret in memory. Keys are derived on demand from this secret, and not stored anywhere.
Other backends make use of external Key Management Systems (KMS) for key derivation and securing a master secret.
When using an external KMS backend, encrypted keys are stored in cloud buckets.
