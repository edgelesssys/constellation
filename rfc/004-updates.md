# RFC 004: Constellation updates

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

The CLI will use a lookup table to map the Kubernetes version from the config to URLs and hashes. Those are sent over during `constellation init` and used by the first Bootstrapper. Then, the URLs and hashes are pushed to the `k8s-components-1.23.12` ConfigMap and the `k8s-components-1.23.12` ConfigMap is referenced by the `NodeVersion` CR named `constellation-version`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-components-1.23.12-sha256-8ae09b7e922a90fea7a4259fb096f73e9efa948ea2f09349618102a328c44b8b
  namespace: kube-system
immutable: true
data:
  components:
    '[{"URL":"https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz","Hash":"sha256:b275772da4026d2161bf8a8b41ed4786754c8a93ebfb6564006d5da7f23831e5","InstallPath":"/opt/cni/bin","Extract":true},{"URL":"https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.25.0/crictl-v1.25.0-linux-amd64.tar.gz","Hash":"sha256:86ab210c007f521ac4cdcbcf0ae3fb2e10923e65f16de83e0e1db191a07f0235","InstallPath":"/run/state/bin","Extract":true},{"URL":"https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubelet","Hash":"sha256:2da0b93857cf352bff5d1eb42e34d398a5971b63a53d8687b45179a78540d6d6","InstallPath":"/run/state/bin/kubelet","Extract":false},{"URL":"https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubeadm","Hash":"sha256:9fea42b4fb5eb2da638d20710ebb791dde221e6477793d3de70134ac058c4cc7","InstallPath":"/run/state/bin/kubeadm","Extract":false},{"URL":"https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubectl","Hash":"sha256:f93c18751ec715b4d4437e7ece18fe91948c71be1f24ab02a2dde150f5449855","InstallPath":"/run/state/bin/kubectl","Extract":false}]'
```

The JoinService will look at the `k8s-components-1.23.12` ConfigMap in addition to the `NodeVersion` CR named `constellation-version`. Currently, the `k8s-version` ConfigMap is mounted into the JoinService pod. We will change that so that the JoinService requests the `kubernetesComponentsReference` from `constellation-version` and then uses this to look up the Kubernetes components.
Those components are then sent to any node requesting to join the cluster.

Additionally, each node trying to join the cluster is tracked as a `JoiningNode` CR.
The JoinService creates a `JoiningNode` CRD for each issued JoinTicket with the node's name and reference to the Kubernetes components ConfigMap it was sent. This `JoiningNode` CRD is consumed by the node operator.

## Extending the Bootstrapper

During the cluster initialization we need to create the first ConfigMap with components and hashes.
We receive all necessary information from the CLI in the first place, since we need to download them to create a initialize the cluster in the first place.

To be able to even update singular components, we need to know if the set of components of a node is the desired one. To achieve that, the Bootstrapper calculates a hash of all the components' hashes.
Because of the length restriction for labels, we need to attach this information as an annotation to the node.
Annotations cannot be set during the join process (in contrast to node-labels).
Therefore, for every JoinRequest, the JoinService will create a JoiningNode CR.
This CRD will later be consumed by the node operator.
The JoiningNode CRD will contain a `componentsreference` in its spec.

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: JoiningNode
metadata:
  name: leo-1645f3a5-worker000001
spec:
  name: leo-1645f3a5-worker000001
  iscontrolplane: false
  componentsreferece: k8s-components-sha256-4054c3597f2ff5c582aaaf212db56db2b14037e79148d82d95dc046f4fc6d92e
  deadline: "2023-01-04T10:30:35Z"
```

## Creating an upgrade agent

We somehow need to download and execute `kubeadm upgrade plan` and `kubeadm upgrade apply vX.Y.Z` on the host system of a control plane node. For security reasons, we don't want those capabilities attached to any pod. Therefore, we opted for a simple and small agent, which exposes a narrow and predefined API as a socket on the control-plane host. This socket can then be mounted into the node operator pod running on a control plane node.

The agent will expose the following service:

```proto
service Update {
    rpc ExecuteUpdate(ExecuteUpdateRequest) returns (ExecuteUpdateResponse);
}

message ExecuteUpdateRequest {
    string kubeadm_url = 1;
    string kubeadm_hash = 2;
    string wanted_kubernetes_version = 3;
}

message ExecuteUpdateResponse {
}
```

The dependency and usage of the upgrade agent by the node operator is explained in the next section.

## Extending the node operator

First, the node operator consumes the JoiningNode CRD. It watches on changes in the CRD list as well as changes in the node list. The controller reconciles the JoiningNode CRDs by trying to annotate the corresponding node. If successful, the controller deletes the CRD.

Second, we need to extend the node operator to also handle Kubernetes updates. The operator already receives information about the Kubernetes version of each node.

The CLI hands users the same mechanism to deliver the Kubernetes version to the operator as we currently use for the image reference:

```patch
// NodeImageSpec defines the desired state of NodeImage.
-type NodeImageSpec struct {
+type NodeVersionSpec struct {
    // ImageReference is the image to use for all nodes.
    ImageReference string `json:"image,omitempty"`
    // ImageVersion is the CSP independent version of the image to use for all nodes.
    ImageVersion string `json:"imageVersion,omitempty"`
+   // KubernetesComponentsReference is a reference to the ConfigMap containing the Kubernetes components to use for all nodes.
+   KubernetesComponentsReference string `json:"kubernetesComponentsReference,omitempty"`
}
```

Additionally, we will change the `NodeImageStatus` to `NodeVersionStatus` (see `nodeimage_types.go`) along with the corresponding controllers.

The Controller will need to take the following steps to update the Kubernetes version:

* disable autoscaling
* get the kubeadm download URL and hash from the `k8s-components-1.23.12` ConfigMap
* pass the URL and hash over a socket mounted into its container to the local update agent running on the same node
  * The agent downloads the new kubeadm binary, checks its hash and executes `kubeadm upgrade plan` and `kubeadm upgrade apply v1.23.12`
* After the agent returned successfully, update the components reference to `k8s-components-1.23.12` in the `NodeVersion` CRD named `constellation-version`.
* Now, iterate over all nodes, and replace them if their Kubernetes version is outdated

## Extending the `constellation upgrade` command

Currently, `constellation upgrade` allows us to upgrade the VM image via the following entry in the constellation-config.yaml:

```yaml
upgrade:
  image: v2.3.0
  measurements:
    11:
      expected: "0000000000000000000000000000000000000000000000000000000000000000"
      warnOnly: false
    12:
      expected: "0000000000000000000000000000000000000000000000000000000000000000"
      warnOnly: false
```

Instead of having a separate `upgrade` section, we will opt for a declarative approach by updating the existing values of the config file. Since only parts of the config behave in a declarative way,
we should add comments to those fields that will not update the cluster.

```yaml
kubernetesVersion: 1.24.2
microserviceVersion: 2.1.3 # All services deployed as part of installing Constellation
image: v2.3.0
provider:
  azure:
    measurements:
      11:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
      12:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
```

Note that:

* `microserviceVersion` is a bundle of component versions which can be conceptually separated into two groups:
  * Services that are versioned based on the constellation version: : KMS, JoinService, NodeMaintainanceOperator, NodeOperator, OLM, Verification, Cilium (for now).
  There only exists one version of each service and it is compatible with all Kubernetes versions currently supported by Constellation.
  The deployment and image version are the same for all three Kubernetes versions.
  * Services that are versioned based on the kubernetes version: Autoscaler, CloudControllerManager, CloudNodeManager, GCP Guest Agent, Konnectivity.
  There exist one version for each Kubernetes version.
  The deployment is the same for all three Kuberenetes version, but the image is specific to the version.
  Images are specified by the CLI upon loading the Helm chart, by inspeciting `constellation-conf.yaml`.
  Deployment variations could be introduced into the Helm charts if they become necessary in the future.

### CLI commands

`constellation upgrade` has 2 sub commands:

* `constellation upgrade check`
* `constellation upgrade apply`

When `constellation upgrade check` is called it checks if the current CLI includes helm charts and kubernetes components that are newer than the ones configured in `constellation-config.json`.
If this is the case, the CLI prints a list of all components that will be updated.
Moreover, it checks for new image patch versions via the update API (see: rfc/update-api.json).
Image patch versions are forward compatible within one minor version.

Lastly, the CLI checks if a newer CLI version is available via the update API (see: rfc/update-api.json). If this is the case, it will print the latest CLI version instead of the output described above.
If the current version and latest version diverge more than one minor version, it will also show the latest CLI of the next minor version, and suggest a way to download it.
An example:
- current CLI version: `2.3.2`
- available CLI version with minor version `2.4.0`: `2.4.1`, `2.4.2`, `2.4.3`
- latest CLI version: `2.5.0`

`constellation upgrade check` will show `2.5.0` as latest, and suggest that the next step in the upgrade process is `2.4.3`.
Since any CLI can only upgrade from one minor version below to its own version, we need to perform the upgrade to `2.4.3` before upgrading to `2.5.0`.
If there are still microservice updates needed with the current CLI, we need to prompt the user to first install those before continuing with the next minor release.

We also print `In newer CLI versions there are even newer versions available.` if e.g. there is a newer patch version of Kubernetes available in one of the proposed minor versions.

Executing `constellation upgrade check --update-config` updates all new version values to `constellation-conf.json`.
This allows the user to execute `constellation upgrade apply` without manually modifying `constellation-conf.json`.

```bash
$ constellation upgrade check
Possible Kubernetes upgrade:
  1.24.2 --> 1.24.3 (or 1.25.2)
  In newer CLIs there are even newer patch versions available.
Possible VM image upgrade:
  2.3.0 --> 2.3.0 (not updated)

Possible Kubernetes services upgrade to 1.24.5:
  Autoscaler: 1.24.3 --> 1.24.3 (not updated)
  CloudControllerManager: 1.24.5 --> 1.24.8
  CloudNodeManager: 1.24.1 --> 1.24.2

Possible Constellation microservices upgrade to 2.2.0:
  KMS: 2.1.3 --> 2.2.0
  joinService: 2.1.3 --> 2.2.0
  nodeOperator: 2.1.3 --> 2.2.0

There is a new CLI available: v2.6.0

Your current CLI version is more than one minor version apart from the latest release.
Please upgrade your cluster with the current CLI and then download the CLI of the next minor version
and upgrade your cluster again.

You can download it via:

$ wget <CDN link to the CLI with the latest patch version of the next minor version>
```

When `constellation upgrade apply` is called the CLI needs to perform the following steps:

1. warn the user to create a Constellation/etcd backup before updating as documented in the [official K8s update docs](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/#before-you-begin)
2. create a new `k8s-components-1.24.3` ConfigMap with the corresponding URLs and hashes from the lookup table in the CLI
3. update the measurements in the `join-config` ConfigMap
4. update the Kubernetes version and VM image in the `NodeVersion` CRD named `constellation-verison`
5. update Constellation microservices

Since the service versions bundled inside a `microserviceVersion` are hidden, the CLI will print the changes taking place. We also print a warning to back up any important components when the upgrade necessitates a node replacement, i.e. on Kubernetes and VM image upgrades.

```bash
$ constellation upgrade apply
Upgrading Kubernetes: 1.24.2 --> 1.24.3 ...
Upgrading VM image: 2.3.0 --> 2.3.0 (not updated)

Upgrading Kubernetes services version to 1.24.5:
  Autoscaler: 1.24.3 --> 1.24.3 (not updated)
  CloudControllerManager: 1.24.5 --> 1.24.8
  CloudNodeManager: 1.24.1 --> 1.24.2

Upgrading Constellation microservices to 2.2.0:
  KMS: 2.1.3 --> 2.2.0
  joinService: 2.1.3 --> 2.2.0
  nodeOperator: 2.1.3 --> 2.2.0

Warning: Please backup any important components before upgrading Kubernetes
Apply change [yes/No]?

```

# Compatibility
`constellation upgrade` has to handle the version of four components: CLI, image, microservices and Kubernetes.
To do this correctly and keep the cluster in a working condition some constraints are required.

- Constellation microservices, VM image and the CLI are versioned in lockstep, i.e. each release a new version of all components is released.
  The Constellation version references all of these components.
- Each Constellation version is compatible with three or four Kubernetes versions.
  When a new Kubernetes version is released, Constellation will support four versions for one release cycle; before phasing out the oldest Kubernetes version.
  To learn if microserviceVersion A.B.C is compatible with Kubernetes version X.Y, one has to check whether the Constellation version is compatible with the Kubernetes version X.Y.
- Each Constellation version X.Y.Z is compatible with all patch versions of the next Constellation version X.Y+1.
- The individual versions of Microservice, image and CLI all have to be compatible with the targeted Kubernetes version.
  Each component can be upgraded independently.
  This means that during a Kubernetes upgrade the oldest Constellation component has to be compatible with the new Kubernetes version.
  If not, this oldest component has to be updated first.
- For Kubernetes versions the version-target of an upgrade has to be supported by the current Constellation version.
  The currently running Kubernetes version is not relevant for the Kubernetes upgrade.

These constraints are enforced by the CLI.
If users decide to change specific versions by changing the Kubernetes resources, nothing is stopping them.

The compatibility information should be separated from the enforcement code.
This way a minimal implementation can be created where the compatibility information is embedded into the CLI.
As a next step the information can be served through the [Constellation API](./008-apis.md).
By serving the compatibility information dynamically, faulty versions can be excluded from upgrade paths even after they have been released.
