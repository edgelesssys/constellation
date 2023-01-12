/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Amazon Web Services attestation

Attestation for AWS using [NitroTPM].

AWS currently does not support confidential VMs, but offers a TPM 2.0 compliant vTPM integration.
We use this to enable a TPM based measured boot Constellation deployment.

# Issuer

The TPM attestation is signed by the NitroTPM's RSA attestation key.
Additionally to the TPM attestation, we attach a node's [instance identity document] to the attestation document.

# Validator

Currently, the NitroTPM provides no endorsement certificate for its attestation key, nor does AWS offer a secondary of of verifying it.
For now we have to blindly trust the key.

Additionally to verifying the TPM attestation, we also check the instance identity document for consistency.

[NitroTPM]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nitrotpm.html
[instance identity document]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
*/
package aws
