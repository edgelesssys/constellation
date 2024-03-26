/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# GCP SEV-ES Attestation

Google offers [confidential VMs], utilizing AMD SEV-ES to provide memory encryption.

AMD SEV-ES doesn't offer much in terms of remote attestation, and following that the VMs don't offer much either, see [their docs] on how to validate a confidential VM for some insights.
However, each VM comes with a [virtual Trusted Platform Module (vTPM)].
This module can be used to generate VM unique encryption keys or to attest the platform's chain of boot. We can use the vTPM to verify the VM is running on AMD SEV-ES enabled hardware, allowing us to bootstrap a constellation cluster.

# Issuer

Generates a TPM attestation key using a Google provided attestation key.
Additionally project ID, zone, and instance name are fetched from the metadata server and attached to the attestation document.

# Validator

Verifies the TPM attestation by using a public key provided by Google's API corresponding to the project ID, zone, instance name tuple attached to the attestation document.

# Problems

  - SEV-ES is somewhat limited when compared to the newer version SEV-SNP

    Comparison of SEV, SEV-ES, and SEV-SNP can be seen on page seven of [AMD's SNP whitepaper]

  - We have to trust Google

    Since the vTPM is provided by Google, and they could do whatever they want with it, we have no save proof of the VMs actually being confidential.

  - The provided vTPM has no endorsement certificate for its attestation key

    Without a certificate signing the authenticity of any endorsement keys we have no way of establishing a chain of trust.
    Instead, we have to rely on Google's API to provide us with the public key of the vTPM's endorsement key.

[confidential VMs]: https://cloud.google.com/compute/confidential-vm/docs/about-cvm
[their docs]: https://cloud.google.com/compute/confidential-vm/docs/monitoring
[virtual Trusted Platform Module (vTPM)]: https://cloud.google.com/security/shielded-cloud/shielded-vm#vtpm
[AMD's SNP whitepaper]: https://www.amd.com/system/files/TechDocs/SEV-SNP-strengthening-vm-isolation-with-integrity-protection-and-more.pdf#page=7
*/
package es
