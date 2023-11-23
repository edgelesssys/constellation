# RFC 002: Proof of Ownership

A cluster owner needs a way to prove a cluster belongs to them, while a third-party needs to be able to verify the owner's claims.
For that, the owner generates a private/public key pair.

During `constellation init`, the cluster will generate its own private/public key pair, and send back a signing request for the public key.

The signed public key is measured into a PCR, so the binding of the private/public key to the cluster can be verified through remote attestation.

The cluster is now able to sign data using its own private key.

A third-party can verify a cluster belongs to a specific person in three steps:

1. Verify the signature of data provided by the third-party and signed by the cluster

2. Verify the cluster's public key was signed by the owner

3. Verify the public key is measured into a PCR by validating the cluster's attestation statement

## Workflow

1. Cluster owner generates a private/public key pair

2. The Constellation cluster generates its own private/public key pair and requests the owner to sign the public key during `constellation init`

3. Constellation measures the signed public key into PCR[11] (previously used for ownerID)

4. A third-party requests an attestation from the verification service, providing some data to be signed

5. The verification service signs the data using its private key

6. The verification service returns: attestation document + data signature + signed public key

7. The third-party verifies the public key signature using the owner's public key

8. The third-party calculates the expected PCR[11] using the signed public key and validates the attestation document

9. The data signature is verified, and if successful proving ownership of the cluster

## Encoding

The signed public key measured into PCR[11] is DER encoded.
TODO: Add exact encoding specification
