/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# SNP

Attestation based on TPM and SEV-SNP attestation.
The TPM is used to generate runtime measurements and signed by an attestation key that can be verified using the SEV-SNP attestation report.

# Issuer

Generates a TPM attestation using an attestation key saved in the TPM.
Additionally loads the SEV-SNP attestation report and AMD VCEK certificate chain, and adds them to the attestation document.

# Validator

Verifies the attestation key used by first verifying the VCEK certificate chain and the SNP attestation report.

# Glossary

This section explains abbreviations used in SNP implementation.

  - Attestation Key (AK)

  - AMD Root Key (ARK)

  - AMD Signing Key (ASK)

  - Versioned Chip Endorsement Key (VCEK)

    For more information see [SNP WhitePaper]

  - Host (Hardware?) Compatibility Layer (HCL)

    No public information. Azure compute API has a field `isHostCompatibilityLayerVm`, with only a [single sentence of documentation].

[SNP WhitePaper]: https://www.amd.com/system/files/TechDocs/SEV-SNP-strengthening-vm-isolation-with-integrity-protection-and-more.pdf
[single sentence of documentation]: https://learn.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=windows
*/
package snp
