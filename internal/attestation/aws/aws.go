/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# Amazon Web Services attestation

Constellation supports multiple attestation technologies on AWS.

  - SEV - Secure Nested Paging (SEV-SNP)

    TPM attestation verified using an SEV-SNP attestation statement.
    The TPM runs outside the confidential context.
    The initial firmware measurement included in the SNP report can be calculated idependently.
    The source code of the firmware is publicly available.

  - NitroTPM

    No confidential computing. Attestation via a TPM 2.0 compliant vTPM.
*/
package aws
