# Upgrading Kubernetes

Constellation is a Kubernetes distribution and a sidecar. This means that the Kubernetes version is tracked in multiple places:

- Kubernetes linux binaries installed in the CoreOS image (`kubelet`, `kubectl`, `kubeadm`)
- The desired Kubernetes version deployed by `kubeadm init`
- Kubernetes resources (deployments made while initializing Kubernetes, including the `cloud-controller-manager`, `cluster-autoscaler` and more)
- Kubernetes go dependencies for the bootstrapper code


## Understand what has changed

Before changing the Kubernetes version, it is a very good idea to [read the release notes](https://kubernetes.io/releases/notes/) and to identify breaking changes.


## Prepare CoreOS images

CoreOS is the linux distribution that constellation is built on. The Kubernetes components are installed on CoreOS from [the official Kubernetes RPM sources](https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64).
The installed versions of `kubelet`, `kubeadm` and `kubectl` are pinned to a specific version by the [`manifest-lock.x86_64.json`](https://github.com/edgelesssys/constellation-fedora-coreos-config/blob/constellation/manifest-lock.x86_64.json) in the [`constellation-fedora-coreos-config` repository](https://github.com/edgelesssys/constellation-fedora-coreos-config):

```javascript
{
  "packages": {
    // [...]
    "kubeadm": {
      "evra": "1.23.1-0.x86_64"
    },
    "kubectl": {
      "evra": "1.23.1-0.x86_64"
    },
    "kubelet": {
      "evra": "1.23.1-0.x86_64"
    },
    // [...]
  }
  // [...]
}
```

New CoreOS images with the desired Kubernetes version should be prepared for testing/debugging and a PR should be opened.

## Upgrade the pinned Kubernetes version deployed by kubeadm

Kubeadm is the Kubernetes deployment tool used by constellation. During `kubeadm init`, a Kubernetes version is selected and installed. Using the flag `--kubernetes-version` or the `ClusterConfiguration` field `kubernetesVersion`, this version can be pinned.
To change this version, set the go constant in [`github.com/edgelesssys/constellation/internal/constants.KubernetesVersion`](/internal/constants/constants.go):

```go
const (
    // [...]

    // KubernetesVersion installed by kubeadm.
    KubernetesVersion = "stable-1.23"
)
```

## Upgrading Kubernetes resources

During the cluster initialization, multiple Kubernetes resources are deployed. Some of these should be upgraded with Kubernetes.
Look at [the resources folder](/bootstrapper/internal/kubernetes/k8sapi/resources) and decide what needs to be upgraded. Cloud provider specific images are defined in [`github.com/edgelesssys/constellation/bootstrapper/cloudprovider`](/bootstrapper/cloudprovider/images.go). You can check available version tags for container images using [the container registry tags API](https://docs.docker.com/registry/spec/api/#listing-image-tags):

```
curl -q https://k8s.gcr.io/v2/autoscaling/cluster-autoscaler/tags/list | jq .tags
curl -q https://k8s.gcr.io/v2/cloud-controller-manager/tags/list | jq .tags
curl -q https://us.gcr.io/v2/k8s-artifacts-prod/provider-aws/cloud-controller-manager/tags/list | jq .tags
curl -q https://mcr.microsoft.com/v2/oss/kubernetes/azure-cloud-controller-manager/tags/list | jq .tags
curl -q https://mcr.microsoft.com/v2/oss/kubernetes/azure-cloud-node-manager/tags/list | jq .tags
# [...]
```


## Upgrade go dependencies

The [`go.mod`](/go.mod) and [`go.sum`](/go.sum) files pin versions of the Kubernetes go packages. While these do not need to be on the exact versions used in the Kubernetes deployment, it is a good idea to keep them updated and on a similar version.
Upgrade Kubernetes go dependencies by changing the versions of all packages in the `k8s.io` namespace from the old version to the new version in [`go.mod`](/go.mod) and run `go mod tidy`. Ensure that there are no other conflicts and test your changes.
See the diff of [this PR](https://github.com/edgelesssys/constellation/pull/110) as an example of updating the go dependencies.

## Test the new Kubernetes version

- Setup a Constellation cluster using the new image with the new bootstrapper binary and check if Kubernetes is deployed successfully.

    ```
    # should print the new k8s version for every node
    kubectl get nodes -o wide
    # read the logs for pods deployed in the kube-system namespace and ensure they are healthy
    kubectl -n kube-system get pods
    kubectl -n kube-system logs [...]
    kubectl -n kube-system describe pods
    ```
- Read the logs of the main Kubernetes components by getting a shell on the nodes and scan for errors / deprecation warnings:

    ```
    journalctl -u kubelet
    journalctl -u containerd
    ```

- Conduct e2e tests
  - [Run the sonobuoy test suite against your branch](https://sonobuoy.io/)
  - [Run CI e2e tests](/.github/docs/README.md)
