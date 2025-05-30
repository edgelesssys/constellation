# Feature status of clouds

What works on which cloud? Currently, Confidential VMs (CVMs) are available in varying quality on the different clouds and software stacks.

For Constellation, the ideal environment provides the following:

1. Ability to run arbitrary software and images inside CVMs
2. CVMs based on AMD SEV-SNP (available in EPYC CPUs since the Milan generation) or Intel TDX (available in Xeon CPUs since the Sapphire Rapids generation)
3. Ability for CVM guests to obtain raw hardware attestation statements
4. Reviewable, open-source firmware inside CVMs
5. Capability of the firmware to attest the integrity of the code it passes control to, e.g., with an embedded virtual TPM (vTPM)

(1) is a functional must-have. (2)--(5) are required for remote attestation that fully keeps the infrastructure/cloud out. Constellation can work without them or with approximations, but won't protect against certain privileged attackers anymore.

The following table summarizes the state of features for different infrastructures as of June 2023.

| **Feature**                       | **Azure** | **GCP** | **AWS** | **OpenStack (Yoga)** |
|-----------------------------------|-----------|---------|---------|----------------------|
| **1. Custom images**              | Yes       | Yes     | Yes     | Yes                  |
| **2. SEV-SNP or TDX**             | Yes       | Yes     | Yes     | Depends on kernel/HV |
| **3. Raw guest attestation**      | Yes       | Yes     | Yes     | Depends on kernel/HV |
| **4. Reviewable firmware**        | No*       | No      | Yes     | Depends on kernel/HV |
| **5. Confidential measured boot** | Yes       | No      | No      | Depends on kernel/HV |

## Microsoft Azure

With its [CVM offering](https://docs.microsoft.com/en-us/azure/confidential-computing/confidential-vm-overview), Azure provides the best foundations for Constellation.
Regarding (3), Azure provides direct access to remote-attestation statements.
The firmware runs in an isolated domain inside the CVM and exposes a vTPM (5), but it's closed source (4).
On SEV-SNP, Azure uses VM Privilege Level (VMPL) isolation for the separation of firmware and the rest of the VM; on TDX, they use TD partitioning.
This firmware is signed by Azure.
The signature is reflected in the remote-attestation statements of CVMs.
Thus, the Azure closed-source firmware becomes part of Constellation's trusted computing base (TCB).

\* Recently, [Azure announced the open source paravisor OpenHCL](https://techcommunity.microsoft.com/blog/windowsosplatform/openhcl-the-new-open-source-paravisor/4273172). It's the foundation for fully open source and verifiable CVM firmware. Once Azure provides their CVM firmware with reproducible builds based on OpenHCL, (4) switches from *No* to *Yes*. Constellation will support OpenHCL based firmware on Azure in the future.

## Google Cloud Platform (GCP)

The [CVMs Generally Available in GCP](https://cloud.google.com/confidential-computing/confidential-vm/docs/confidential-vm-overview#amd_sev) are based on AMD SEV but don't have SNP features enabled.
CVMs with [SEV-SNP enabled are in public preview](https://cloud.google.com/confidential-computing/confidential-vm/docs/confidential-vm-overview#amd_sev-snp). Regarding (3), with their SEV-SNP offering Google provides direct access to remote-attestation statements.
However, regarding (5), attestation is partially based on the [Shielded VM vTPM](https://cloud.google.com/compute/shielded-vm/docs/shielded-vm#vtpm) for [measured boot](../architecture/attestation.md#measured-boot), which is a vTPM managed by Google's hypervisor.
Hence, the hypervisor is currently part of Constellation's TCB.
Regarding (4), the CVMs still include closed-source firmware.

In the past, Intel and Google have [collaborated](https://cloud.google.com/blog/products/identity-security/rsa-google-intel-confidential-computing-more-secure) to enhance the security of TDX.
Recently, Google has announced a [private preview for TDX](https://cloud.google.com/blog/products/identity-security/confidential-vms-on-intel-cpus-your-datas-new-intelligent-defense?hl=en).
With TDX on Google, Constellation has a similar TCB and attestation flow as with the current SEV-SNP offering.

## Amazon Web Services (AWS)

Amazon EC2 [supports AMD SEV-SNP](https://aws.amazon.com/de/about-aws/whats-new/2023/04/amazon-ec2-amd-sev-snp/).
Regarding (3), AWS provides direct access to remote-attestation statements.
However, regarding (5), attestation is partially based on the [NitroTPM](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nitrotpm.html) for [measured boot](../architecture/attestation.md#measured-boot), which is a vTPM managed by the Nitro hypervisor.
Hence, the hypervisor is currently part of Constellation's TCB.
Regarding (4), the [firmware is open source](https://github.com/aws/uefi) and can be reproducibly built.

## OpenStack

OpenStack is an open-source cloud and infrastructure management software. It's used by many smaller CSPs and datacenters. In the latest *Yoga* version, OpenStack has basic support for CVMs. However, much depends on the employed kernel and hypervisor. Features (2)--(4) are likely to be a *Yes* with Linux kernel version 6.2. Thus, going forward, OpenStack on corresponding AMD or Intel hardware will be a viable underpinning for Constellation.

## Conclusion

The different clouds and software like the Linux kernel and OpenStack are in the process of building out their support for state-of-the-art CVMs. Azure has already most features in place. For Constellation, the status quo means that the TCB has different shapes on different infrastructures. With broad SEV-SNP support coming to the Linux kernel, we soon expect a normalization of features across infrastructures.
