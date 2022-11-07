# Constellation updates

Things we manage for the user:

1. VM Image
2. Kubernetes
3. Cilium
4. Constellation microservices

## VM Image

Updating the VM Image is already implemented.

## Kubernetes

The current manual approach is:

* ssh into a control-plane node
* curl new kubeadm version (see `versions.go`) to `/var/run/state/bin/kubeadmNew`
* `mv /var/run/state/bin/kubeadmNew /var/run/state/bin/kubeadm && chmod +x /var/run/state/bin/kubeadm`
* `export PATH=$PATH:/var/run/state/bin/`
* `kubeadm upgrade plan`
* `kubeadm upgrade apply vX.Y.Z`
* bump K8s version in configmap/k8s-version to `vX.Y`
* [upgrade your cluster](https://docs.edgeless.systems/constellation/workflows/upgrade) with a new VM image

For more details on the first steps see the [official K8s documentation](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/). This upgrade [will create new kubelet certificates](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#automatic-certificate-renewal) but does [not rotate Kubernetes CA certificate](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/#certificate-authority-rotation).

## Cilium

Cilium is installed via helm. In the long term, we don't need to maintain our fork and Cilium can be updated independently using the official releases.

## Constellation microservices

All Constellation microservices will be bundled into and therefore updated via one helm chart.

# Automatic Updates

## Extending the JoinService

The CLI will use a lookup table to map the Kubernetes version from the config to URLs and hashes. Those are sent over during `constellation init` and used by the first Bootstrapper. Then, the URLs and hashes are pushed to the `k8s-components-1.23.12` ConfigMap and the Kubernetes version with a reference to the `k8s-components-1.23.12` ConfigMap is pushed to `k8s-versions`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-version
  namespace: kube-system
data:
  k8s-version: "1.23.12"
  components: "k8s-components-1.23.12-sha256-8ae09b7e922a90fea7a4259fb096f73e9efa948ea2f09349618102a328c44b8b" # This references the ConfigMap below.
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-components-1.23.12-sha256-8ae09b7e922a90fea7a4259fb096f73e9efa948ea2f09349618102a328c44b8b
  namespace: kube-system
immutable: true
data:
  kubeadm:
    url: "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubeadm"
    hash: "sha256:7dc4799eb1ae31daae0f4d1d09a643f7687dcc78c562e0b453d01799d183d6a0"
  kubelet:
    url: "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubelet"
    hash: "sha256:100413d0badd8b4273052bae43656d2167dc67f613b019adb7c61bd49f37283a"
  kubectl:
    url: "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubectl"
    hash: "sha256:d8a5fb9e2c2b633894648c97d402bc138987c9904212cd88386954e7b2c09865"
  kubeadmConf:
    url: "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf"
    hash: "sha256:fa1dc2cc3fa48fcb83ddcf5e4f2b6853f2f13f2be6507c6fc80364f2e4b0ad6a"
  kubeletService:
    url: "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service"
    hash: "sha256sum:0d96a4ff68ad6d4b6f1f30f713b18d5184912ba8dd389f86aa7710db079abcb0"
  crictl:
    url: "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz"
    hash: "sha256:ebaea1c7b914cdd548012c6cba44f8d760fd0c7915274ffecd6d764957aac83c"
  cniplugins:
    url: "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz"
    hash: "sha256:2f04ce10b514912da87fc9979aa2e82e3a0e431c14fc0ea5c7c7209de74a1491"
```

The JoinService will consume the `k8s-components-1.23.12` ConfigMap in addition to the `k8s-version` ConfigMap. Currently, the `k8s-version` ConfigMap is mounted into the JoinService pod. We will change that so that the JoinService requests the ConfigMap values via the Kubernetes API. If a new node wants to join the cluster, the JoinService looks up the current Kubernetes version and all the component download URLs and hashes and sends them to the joining node.

## Extending the Bootstrapper

During the cluster initialization we need to create the first ConfigMap with components and hashes.
We receive all necessary information from the CLI in the first place, since we need to download them to create a initialize the cluster in the first place.

To be able to even update singular components, we need to know if the set of components of a node is the desired one. To achieve that, the Bootstrapper will calculate a hash of all the components' hashes. The node then labels itself during `kubeadm join` with this hash. The hash will later be read by the node operator. The first Bootstrapper needs to be labeled during `kubeadm init`.

```yaml
apiVersion: kubeadm.k8s.io/v1beta2
kind: JoinConfiguration
discovery:
  bootstrapToken:
    token: XXX
    apiServerEndpoint: "XXX"
    caCertHashes: ["sha256:123456789012345678901234567890564b934be406f13e28f118b32cc0b6e6db"]
nodeRegistration:
  name: worker-node-1
  kubeletExtraArgs:
    node-labels: "edgeless.systems/kubernetes-components-hash="sha256:8ae09b7e922a90fea7a4259fb096f73e9efa948ea2f09349618102a328c44b8b""
```

## Creating an upgrade agent

We somehow need to download and execute `kubeadm upgrade plan` and `kubeadm upgrade apply vX.Y.Z` on the host system of a control plane node. For security reasons, we don't want those capabilities attached to any pod. Therefore, we opted for a simple and small agent, which exposes a narrow and predefined API as a socket on the control-plane host. This socket can then be mounted into the node operator pod running on a control plane node.

The agent will expose the following service:

```proto
service Update {
    rpc ExecuteUpdate(ExecuteUpdateRequest) returns (ExecuteUpdateResponse);
}

message ExecuteUpdateRequest {
    string kubeadm_URL = 1;
    string kubeadm_Hash = 2;
    string wanted_Kubernetes_Version = 3;
}

message ExecuteUpdateResponse {
}
```

The dependency and usage of the upgrade agent by the node operator is explained in the next section.

## Extending the node operator

We need to extend the node operator to also handle Kubernetes updates. The operator already receives information about the Kubernetes version of each node.

The CLI hands users the same mechanism to deliver the Kubernetes version to the operator as we [currently use for the image reference](https://github.com/edgelesssys/constellation/blob/main/operators/constellation-node-operator/api/v1alpha1/nodeimage_types.go#L14):

```patch
// NodeImageSpec defines the desired state of NodeImage.
-type NodeImageSpec struct {
+type NodeSpec struct {
    // ImageReference is the image to use for all nodes.
    ImageReference string `json:"image,omitempty"`
+   // KubernetesVersion defines the Kubernetes version for all nodes.
+   KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}
```

Additionally, we will change the `NodeImageStatus` to `NodeStatus` (see `nodeimage_types.go`) along with the corresponding controllers.

The Controller will need to take the following steps to update the Kubernetes version:

* disable autoscaling
* get the kubeadm download URL and hash from the `k8s-components-1.23.12` ConfigMap
* pass the URL and hash over a socket mounted into its container to the local update agent running on the same node
  * The agent downloads the new kubeadm binary, checks its hash and executes `kubeadm upgrade plan` and `kubeadm upgrade apply v1.23.12`
* After the agent returned successfully, update the Kubernetes version to `1.23.12` and components reference to `k8s-components-1.23.12` in the `k8s-version` ConfigMap
* Now, iterate over all nodes, and replace them if their Kubernetes version is outdated

## Extending the `constellation upgrade` command

Currently, `constellation upgrade` allows us to upgrade the VM image via the following entry in the constellation-config.yaml:

```yaml
upgrade:
  image: /communityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.3.0
  measurements:
    11: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
    12: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
```

Instead of having a separate `upgrade` section, we will opt for a declarative approach by updating the existing values of the config file. Since only parts of the config behave in a declarative way,
we should add comments to those fields who will not update the cluster.

```yaml
kubernetesVersion: 1.24.3
kubernetesServicesVersion: 1.24.5 # Bundled Kubernetes components (Autoscaler, CloudControllerManager, CloudNodeManager, GCP Guest Agent, Konnectivity)
constellationVersion: 2.2.0 # or microserviceVersion: (KMS, JoinService, NodeMaintainanceOperator, NodeOperator, OLM, Verification, Cilium)
provider:
  azure:
    image: /communityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.3.0
    measurements:
    11: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
    12: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
```

Note that:

* `constellationVersion` 2.2.0 contains components which are all released in version 2.2.0
* `kubernetesServicesVersion` 1.24.5 could contain Autoscaler 1.24.2, CCM 1.24.8 since their patch versions are not in sync with Kubernetes. Moreover, those component versions will be bundled by us. Think: public lookup table from `kubernetesServicesVersion` -> component version.

When `constellation upgrade apply` is called the CLI needs to perform the following steps:

1. warn the user to create a Constellation/etcd backup before updating as documented in the [official K8s update docs](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/#before-you-begin)
2. create a new `k8s-components-1.24.3` ConfigMap with the corresponding URLs and hashes from the lookup table in the CLI
3. update the measurements in the `join-config` ConfigMap
4. update the Kubernetes version and VM image in the `nodeimage` CRD
5. update Cilium + Constellation microservices

The actual update in step 2. and 3. will be handled by the node-operator inside Constellation. Step 4. will be done via client side helm deployments.

Since the actual Kubernetes components and Constellation microservice versions are hidden, we will show the user for the actual changes taking place. We also print a warning to back up any important components when the upgrade necessitates a node replacement, i.e. on Kubernetes and VM image upgrades.

```bash
$ constellation upgrade apply
Upgrading Kubernetes: 1.24.2 --> 1.24.3 ...
Upgrading VM image: /communityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.3.0 --> /communityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.3.0 (not updated)

Updating Kubernetes services version to 1.24.5:
  Autoscaler: 1.24.3 --> 1.24.3 (not updated)
  CloudControllerManager: 1.24.5 --> 1.24.8
  CloudNodeManager: 1.24.1 --> 1.24.2

Updating Constellation microservices to 2.2.0:
  KMS: 2.1.3 --> 2.2.0
  joinService: 2.1.3 --> 2.2.0
  nodeOperator: 2.1.3 --> 2.2.0

Warning: Please backup any important components before upgrading Kubernetes
Apply change [yes/No]?
```
