# Upgrading Kubernetes

Constellation is a Kubernetes distribution. As such, dependencies on Kubernetes versions exist in multiple places:

- The desired Kubernetes version deployed by `kubeadm init`
- Kubernetes resources (deployments made while initializing Kubernetes, including the `cloud-controller-manager`, `cluster-autoscaler` and more)
- Kubernetes go dependencies for the bootstrapper code

## Understand what has changed

Before adding support for a new Kubernetes version, it is a very good idea to [read the release notes](https://kubernetes.io/releases/notes/) and to identify breaking changes.

## Upgrading Kubernetes resources

Everything related to Kubernetes versions is tracked in [the versions file](/internal/versions/versions.go). Add a new `ValidK8sVersion` and fill out the `VersionConfigs` entry for that version.
During cluster initialization, multiple Kubernetes resources are deployed. Some of these should be upgraded with Kubernetes.
You can check available version tags for container images using [the container registry tags API](https://docs.docker.com/registry/spec/api/#listing-image-tags):

```sh
curl -qL https://registry.k8s.io/v2/autoscaling/cluster-autoscaler/tags/list | jq .tags
curl -qL https://registry.k8s.io/v2/cloud-controller-manager/tags/list | jq .tags
curl -qL https://registry.k8s.io/v2/provider-aws/cloud-controller-manager/tags/list | jq .tags
curl -qL https://mcr.microsoft.com/v2/oss/kubernetes/azure-cloud-controller-manager/tags/list | jq .tags
curl -qL https://mcr.microsoft.com/v2/oss/kubernetes/azure-cloud-node-manager/tags/list | jq .tags
# [...]
```

Normally renovate will handle the upgrading of Kubernetes dependencies.

## Test the new Kubernetes version

- Setup a Constellation cluster using the new image with the new bootstrapper binary and check if Kubernetes is deployed successfully.

    ```sh
    # should print the new k8s version for every node
    kubectl get nodes -o wide
    # read the logs for pods deployed in the kube-system namespace and ensure they are healthy
    kubectl -n kube-system get pods
    kubectl -n kube-system logs [...]
    kubectl -n kube-system describe pods
    ```

- Read the logs of the main Kubernetes components by getting a shell on the nodes and scan for errors / deprecation warnings:

    ```sh
    journalctl -u kubelet
    journalctl -u containerd
    ```

- Conduct e2e tests
  - [Run the sonobuoy test suite against your branch](https://sonobuoy.io/)
  - [Run CI e2e tests](github-actions.md)
