# Feature status of clouds

What works on which cloud? Currently, Confidential VMs (CVMs) are available in varying quality on the different clouds and software stacks.

For Constellation, the ideal environment provides the following:

1. Ability to run arbitrary software and images inside CVMs
2. CVMs based on AMD SEV-SNP (available in EPYC CPUs since the Milan generation) or, in the future, Intel TDX (available in Xeon CPUs from the Sapphire Rapids generation onward)
3. Ability for CVM guests to obtain raw attestation statements directly from the CPU, ideally via a TPM-like interface
4. Reviewable, open-source firmware inside CVMs

(1) is a functional must-have. (2)--(4) are required for remote attestation that fully keeps the infrastructure/cloud out. Constellation can work without them or with approximations, but won't protect against certain privileged attackers anymore.

The following table summarizes the state of features for different infrastructures as of September 2022.

| **Feature**                   | **Azure** | **GCP** | **AWS** | **OpenStack (Yoga)** |
|-------------------------------|-----------|---------|---------|----------------------|
| **1. Custom images**          | Yes       | Yes     | No      | Yes                  |
| **2. SEV-SNP or TDX**         | Yes       | No      | No      | Depends on kernel/HV |
| **3. Raw guest attestation**  | Yes       | No      | No      | Depends on kernel/HV |
| **4. Reviewable firmware**    | No*       | No      | No      | Depends on kernel/HV |

## Microsoft Azure

With its [CVM offering](https://docs.microsoft.com/en-us/azure/confidential-computing/confidential-vm-overview), Azure provides the best foundations for Constellation. Regarding (3), Azure provides direct access to remote-attestation statements. However, regarding (4), the standard CVMs still include closed-source firmware running in VM Privilege Level (VMPL) 0. This firmware is signed by Azure. The signature is reflected in the remote-attestation statements of CVMs. Thus, the Azure closed-source firmware becomes part of Constellation's trusted computing base (TCB).

\* Recently, Azure [announced](https://techcommunity.microsoft.com/blog/windowsosplatform/openhcl-the-new-open-source-paravisor/4273172) the open source paravisor OpenHCL. It is the foundation for fully open source and verifiable CVM firmware. Once, Azure provides their CVM firmware with reproducible builds based on OpenHCL, (4) switches from *No* to *Yes*. Constellation will support OpenHCL based firmware on Azure in the future.

## Google Cloud Platform (GCP)

The [CVMs available in GCP](https://cloud.google.com/confidential-computing/confidential-vm/docs/confidential-vm-overview#amd_sev) are based on AMD SEV but don't have SNP features enabled. This impacts attestation capabilities. Currently, GCP doesn't offer CVM-based attestation at all. Instead, GCP provides attestation statements based on its regular [vTPM](https://cloud.google.com/blog/products/identity-security/virtual-trusted-platform-module-for-shielded-vms-security-in-plaintext), which is managed by the hypervisor. On GCP, the hypervisor is thus currently part of Constellation's TCB.

## Amazon Web Services (AWS)

AWS currently doesn't offer CVMs. AWS proprietary Nitro Enclaves offer some related features but [are explicitly not designed to keep AWS itself out](https://aws.amazon.com/blogs/security/confidential-computing-an-aws-perspective/). Besides, they aren't suitable for running entire Kubernetes nodes inside them. Therefore, Constellation uses regular EC2 instances on AWS [Nitro](https://aws.amazon.com/ec2/nitro/) without runtime encryption. Attestation is based on the [NitroTPM](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nitrotpm.html), which is a vTPM managed by the Nitro hypervisor. Hence, the hypervisor is currently part of Constellation's TCB.

## OpenStack

OpenStack is an open-source cloud and infrastructure management software. It's used by many smaller CSPs and datacenters. In the latest *Yoga* version, OpenStack has basic support for CVMs. However, much depends on the employed kernel and hypervisor. Features (2)--(4) are likely to be a *Yes* with Linux kernel version 6.2. Thus, going forward, OpenStack on corresponding AMD or Intel hardware will be a viable underpinning for Constellation.

## Conclusion

The different clouds and software like the Linux kernel and OpenStack are in the process of building out their support for state-of-the-art CVMs. Azure has already most features in place. For Constellation, the status quo means that the TCB has different shapes on different infrastructures. With broad SEV-SNP support coming to the Linux kernel, we soon expect a normalization of features across infrastructures.
