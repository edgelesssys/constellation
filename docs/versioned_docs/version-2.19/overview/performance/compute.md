# Impact of runtime encryption on compute performance

All nodes in a Constellation cluster are executed inside Confidential VMs (CVMs). Consequently, the performance of Constellation is inherently linked to the performance of these CVMs.

## AMD and Azure benchmarking

AMD and Azure have collectively released a [performance benchmark](https://community.amd.com/t5/business/microsoft-azure-confidential-computing-powered-by-3rd-gen-epyc/ba-p/497796) for CVMs that utilize 3rd Gen AMD EPYC processors (Milan) with SEV-SNP. This benchmark, which included a variety of mostly compute-intensive tests such as SPEC CPU 2017 and CoreMark, demonstrated that CVMs experience only minor performance degradation (ranging from 2% to 8%) when compared to standard VMs. Such results are indicative of the performance that can be expected from compute-intensive workloads running with Constellation on Azure.

## AMD and Google benchmarking

Similarly, AMD and Google have jointly released a [performance benchmark](https://www.amd.com/system/files/documents/3rd-gen-epyc-gcp-c2d-conf-compute-perf-brief.pdf) for CVMs employing 3rd Gen AMD EPYC processors (Milan) with SEV-SNP. With high-performance computing workloads such as WRF, NAMD, Ansys CFS, and Ansys LS_DYNA, they observed analogous findings, with only minor performance degradation (between 2% and 4%) compared to standard VMs. These outcomes are reflective of the performance that can be expected for compute-intensive workloads running with Constellation on GCP.
