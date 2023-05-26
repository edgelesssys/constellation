/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# NitroTPM Attestation.

Uses NitroTPM to enable a TPM based measured boot Constellation deployment.
The origin of the attesation statement can not be verified.

# Issuer

The TPM attestation is signed by the NitroTPM's RSA attestation key.
Additionally to the TPM attestation, we attach a node's [instance identity document] to the attestation document.

# Validator

Currently, the NitroTPM provides no endorsement certificate for its attestation key, nor does AWS offer an alternative way of verifying it.
For now we have to blindly trust the key.

Additionally to verifying the TPM attestation, we also check the instance identity document for consistency.

[instance identity document]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html

[NitroTPM]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nitrotpm.html
*/
package nitrotpm
