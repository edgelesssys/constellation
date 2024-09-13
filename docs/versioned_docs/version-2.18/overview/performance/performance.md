# Performance analysis of Constellation

This section provides a comprehensive examination of the performance characteristics of Constellation.

## Runtime encryption

Runtime encryption affects compute performance. [Benchmarks by Azure and Google](compute.md) show that the performance degradation of Confidential VMs (CVMs) is small, ranging from 2% to 8% for compute-intensive workloads.

## I/O performance benchmarks

We evaluated the [I/O performance](io.md) of Constellation, utilizing a collection of synthetic benchmarks targeting networking and storage.
We further compared this performance to native managed Kubernetes offerings from various cloud providers, to better understand how Constellation stands in relation to standard practices.

## Application benchmarking

To gauge Constellation's applicability to well-known applications, we performed a [benchmark of HashiCorp Vault](application.md) running on Constellation.
The results were then compared to deployments on the managed Kubernetes offerings from different cloud providers, providing a tangible perspective on Constellation's performance in actual deployment scenarios.
