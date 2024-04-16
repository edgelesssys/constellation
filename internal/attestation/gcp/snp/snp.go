/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# GCP SEV-SNP attestation

Google offers [confidential VMs], utilizing AMD SEV-SNP to provide memory encryption.

Each SEV-SNP VM comes with a [virtual Trusted Platform Module (vTPM)].
This vTPM can be used to generate encryption keys unique to the VM or to attest the platform's boot chain.
We can use the vTPM to verify the VM is running on AMD SEV-SNP enabled hardware and booted the expected OS image, allowing us to bootstrap a constellation cluster.

# Issuer

Retrieves an SEV-SNP attestation statement for the VM it's running in. Then, it generates a TPM attestation statement, binding the SEV-SNP attestation statement to it by including its hash in the TPM attestation statement.
Without binding the SEV-SNP attestation statement to the TPM attestation statement, the SEV-SNP attestation statement could be used in a different VM. Furthermore, it's important to first create the SEV-SNP attestation statement
and then the TPM attestation statement, as otherwise, a non-CVM could be used to create a valid TPM attestation statement, and then later swap the SEV-SNP attestation statement with one from a CVM.
Additionally project ID, zone, and instance name are fetched from the metadata server and attached to the attestation statement.

# Validator

First, it verifies the SEV-SNP attestation statement by checking the signatures and claims. Then, it verifies the TPM attestation by using a
public key provided by Google's API corresponding to the project ID, zone, instance name tuple attached to the attestation document, and confirms whether the SEV-SNP attestation statement is bound to the TPM attestation statement.

# Problems

  - We have to trust Google

    Since the vTPM is provided by Google, and they could do whatever they want with it, we have no save proof of the VMs actually being confidential.

  - The provided vTPM has no endorsement certificate for its attestation key

    Without a certificate signing the authenticity of any endorsement keys we have no way of establishing a chain of trust.
    Instead, we have to rely on Google's API to provide us with the public key of the vTPM's endorsement key.

[confidential VMs]: https://cloud.google.com/compute/confidential-vm/docs/about-cvm
[virtual Trusted Platform Module (vTPM)]: https://cloud.google.com/security/shielded-cloud/shielded-vm#vtpm
*/
package snp
