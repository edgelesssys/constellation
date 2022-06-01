# Google Cloud Platform attestation

Google offers [confidential VMs](https://cloud.google.com/compute/confidential-vm/docs/about-cvm), utilizing AMD SEV-ES to provide memory encryption.

AMD SEV-ES doesn't offer much in terms of remote attestation, and following that the VMs don't offer much either, see [their docs](https://cloud.google.com/compute/confidential-vm/docs/monitoring) on how to validate a confidential VM for some insights.
However, each VM comes with a [virtual Trusted Platform Module (vTPM)](https://cloud.google.com/security/shielded-cloud/shielded-vm#vtpm). 
This module can be used to generate VM unique encryption keys or to attest the platform's chain of boot. We can use the vTPM to verify the VM is running on AMD SEV-ES enabled hardware, allowing us to bootstrap a constellation cluster.

## vTPM components

For attestation we make use of multiple vTPM features:

* Endorsement Key

    Asymemetric key used to establish trust in other keys issued by the TPM or used directly for attestation. The private part never leaves the TPM, while the public part, referred to as Endorsement Public Key (EPK), is available to remote parties. The TPM can issue new keys, signed by its endorsement key, which can then be verified by a remote party using the EPK.

* Endorsement Publc Key Certificate (EPKC)

    A Certificate signed by the TPM manufacturer verifying the authenticity of the EPK. The public key of the Certificate is the EPK.

* Event Log

    A log of events over the boot process.

* [Platform Control Register (PCR)](https://link.springer.com/chapter/10.1007/978-1-4302-6584-9_12)

    Registers holding measurements of software and configuration data. PCR values are not directly written, but updated: a new value is the digest of the old value concatenated with the to be added data.
    Contents of the PCRs can be signed for attestation. Providing proof to a remote party about software running on the system.

## Attestation flow

1. The VM boots and writes its measured software state to the PCRs.

2. The PCRs are hashed and signed by the EPK.

3. An attestation statement is created, containing the EPK, the original PCR values, the hashed PCRs, the signature, and the event log.

4. A remote party establishes trust in the TPMs EPK by verifying its EPKC with the TPM manufactures CA certificate chain.

    Google's vTPMs have no EPKC, instead we querie the GCE API to retrieve a VMs public signing key. This is the public part of the endorsment key used to sign the attestation document.
    The downside to this is the verifying party requiring permissions to access the GCE API.

5. The remote party verifies the signature was created by the TPM, and the hash matches the PCRs.

6. The remote party reads the event log and verifies measuring the event log results in the given PCR values

7. The software state is now verified, the only thing left to do is to decide if the state is good or not. This is done by comparing the given PCR values to a set of expected PCR values.


## Problems

* SEV-ES is somewhat limited when compared to the newer version SEV-SNP

    Comparison of SEV, SEV-ES, and SEV-SNP can be seen on page seven of [AMD's SNP whitepaper](https://www.amd.com/system/files/TechDocs/SEV-SNP-strengthening-vm-isolation-with-integrity-protection-and-more.pdf#page=7)

* We have to trust Google

    Since the vTPM is provided by Google, and they could do whatever they want with it, we have no save proof of the VMs actually being confidential.

* The provided vTPM has no EPKC

    Without a certificate signing the authenticity of any endorsement keys we have no way of establishing a chain of trust.


## Code testing

Running tests for GCP attestation requires root access on a confidential VM.
The VM needs to have read access for the Compute Engine API. This is not an IAM role, but has to be set on the VM itself.

Build and run the tests:

```shell
go test -c -tags gcp
sudo ./gcp.test
```