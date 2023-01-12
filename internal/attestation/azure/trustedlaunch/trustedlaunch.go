/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Trusted Launch

Use Azure's trusted launch vTPM to enable a TPM based measure boot Constellation.

# Issuer

Generates a TPM attestation using an attestation key saved in the TPM.
Additionally an endorsement certificate of the key, and corresponding CA certificate chain are added to the attestation document.

# Validator

Verifies the TPM attestation statement using the public key of the endorsement certificate.
The certificate is verified by first verifying its CA certificate chain.
*/
package trustedlaunch
