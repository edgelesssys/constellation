/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
--------- WARNING! ---------

THIS PACKAGE DOES CURRENTLY NOT IMPLEMENT ANY SNP ATTESTATION.
It exists to implement required interfaces while implementing other parts of the AWS SNP attestation variant within Constellation.

----------------------------

# SNP

Attestation based on TPMs and AMD SEV-SNP.
The TPM is used to generate runtime measurements and sign them with an attestation key.
The TPM currently runs outside the confidential context. This is a limitation imposed by the AWS implementation.

# Issuer

Generates a TPM attestation using an attestation key saved inside the TPM.
Additionally loads the SEV-SNP attestation report and AMD VLEK certificate chain, and adds them to the attestation document.
The SNP report includes a measurement of the initial firmware inside the CVM, which can be precalculated independently for verification.
The report also includes the attestation key.

# Validator

Verifies the SNP report by verifying the VLEK certificate chain and the report's signature.
This estabilishes trust in the attestation key and the CVM's initial firmware.
However, since the TPM is outside the confidential context, it has to be trusted without verification.
Thus, the hypervisor is still included in the trusted computing base.

# Glossary

This section explains abbreviations used in SNP implementation.

  - Platform Security Processor (PSP)

  - Certificate Revocation List (CRL)

  - Attestation Key (AK)

  - AMD Root Key (ARK)

  - AMD Signing Key (ASK)

  - Versioned Chip Endorsement Key (VCEK)

  - Versioned Loaded Endorsement Key (VLEK)
    For more information see [SNP WhitePaper]

[SNP WhitePaper]: https://www.amd.com/system/files/TechDocs/SEV-SNP-strengthening-vm-isolation-with-integrity-protection-and-more.pdf
*/
package snp
