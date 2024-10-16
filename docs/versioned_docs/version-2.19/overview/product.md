# Product features

Constellation is a Kubernetes engine that aims to provide the best possible data security in combination with enterprise-grade scalability and reliability features---and a smooth user experience.

From a security perspective, Constellation implements the [Confidential Kubernetes](confidential-kubernetes.md) concept and corresponding security features, which shield your entire cluster from the underlying infrastructure.

From an operational perspective, Constellation provides the following key features:

* **Native support for different clouds**: Constellation works on Amazon Web Services (AWS), Microsoft Azure, Google Cloud Platform (GCP), and STACKIT. Support for OpenStack-based environments is coming with a future release. Constellation securely interfaces with the cloud infrastructure to provide [cluster autoscaling](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler), [dynamic persistent volumes](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/), and [service load balancing](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer).
* **High availability**: Constellation uses a [multi-master architecture](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/high-availability/) with a [stacked etcd topology](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology) to ensure high availability.
* **Integrated Day-2 operations**: Constellation lets you securely [upgrade](../workflows/upgrade.md) your cluster to a new release. It also lets you securely [recover](../workflows/recovery.md) a failed cluster. Both with a single command.
* **Support for Terraform**: Constellation includes a [Terraform provider](../workflows/terraform-provider.md) that lets you manage the full lifecycle of your cluster via Terraform.
