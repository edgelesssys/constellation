# Product features

Constellation is a confidential orchestration platform, designed to be the most secure way to run Kubernetes.
It leverages confidential computing to isolate entire Kubernetes deployments and all workloads from the infrastructure.
From the inside, a Constellation cluster feels 100% like Kubernetes as you know it.
But for everyone else, from the outside, it’s runtime-encrypted VMs talking over encrypted channels and writing encrypted data.

Constellation provides confidential computing enhancements to Kubernetes, including the following:

* Leveraging confidential VMs (CVMs) available in all major clouds to isolate and encrypt the Kubernetes control-plane and worker nodes.
* Node attestation including a [verified boot](../architecture/images.md#measured-boot) that roots in hardware-measured attestation provided by CVM technologies.
* Operating a [container network interface (CNI) plugin](../architecture/networking.md) between CVMs for encrypted network communications in your cluster. Enabling TLS offloading.
* [CVM-level persistent volume encryption](../architecture//encrypted-storage.md) ensures the confidentiality and integrity of persistent data outside of the Kubernetes cluster.
* [Confidential key management](../architecture//keys.md).
* Verifiable, measured, and authenticated [updates](../architecture/orchestration.md#upgrades) of node OS images and Kubernetes components.

Constellation provides an enterprise-ready Kubernetes environment with key features such as:

* Multi-cloud deployments. You can deploy Constellation clusters to all major cloud platforms for a consistent confidential orchestration platform.
* Highly available (HA) Confidential Kubernetes cluster with [stacked etcd topology](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology).
* Integrating with the Kubernetes cloud controller manager (CCM) to securely provide cloud services such as [cluster autoscaling](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler), [dynamic persistent volumes](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/), and [service load balancing](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer).
