# constellation-node-operator

The constellation node operator manages the lifecycle of constellation nodes after cluster initialization.
In particular, it is responsible for updating the OS images of nodes by replacing nodes running old images with new nodes.

## High level goals

- Admin or `constellation apply` can create custom resources for node related components
- The operator will manage nodes in the cluster by trying to ensure every node has the specified image
- If a node uses an outdated image, it will be replaced by a new node
- Admin can update the specified image at any point in time which will trigger a rolling upgrade through the cluster
- Nodes are replaced safely (cordon, drain, preservation of node labels)

## Description

The operator has multiple controllers with corresponding custom resource definitions (CRDs) that are responsible for the following high level tasks:

### NodeVersion

`NodeVersion` is the only user controlled CRD. The spec allows an administrator to update the desired image and trigger a rolling update.

Example for GCP:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: NodeVersion
metadata:
  name: constellation-version
spec:
  image: "projects/constellation-images/global/images/<image-name>"
```

Example for Azure:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: NodeVersion
metadata:
  name: constellation-version
spec:
  image: "/subscriptions/<subscription-id>/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/<image-definition-name>/versions/<image-version>"
```

### AutoscalingStrategy

`AutoscalingStrategy` is used and modified by the `NodeVersion` controller to pause the `cluster-autoscaler` while an image update is in progress.

Example:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: AutoscalingStrategy
metadata:
  name: autoscalingstrategy
spec:
  enabled: true
  deploymentName: "cluster-autoscaler"
  deploymentNamespace: "kube-system"
```

### ScalingGroup

`ScalingGroup` represents one scaling group at the CSP. Constellation uses one scaling group for worker nodes and one for control-plane nodes.
The scaling group controller will automatically set the image used for newly created nodes to be the image set in the `NodeVersion` Spec. On cluster creation, one instance of the `ScalingGroup` resource per scaling group at the CSP is created. It does not need to be updated manually.

Example for GCP:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: ScalingGroup
metadata:
  name: scalinggroup-worker
spec:
  nodeImage: "constellation-version"
  groupId: "projects/<project-id>/zones/<zone>/instanceGroupManagers/<instance-group-name>"
  autoscaling: true
```

Example for Azure:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: ScalingGroup
metadata:
  name: scalinggroup-worker
spec:
  nodeImage: "constellation-version"
  groupId: "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>"
  autoscaling: true
```

### PendingNode

`PendingNode` represents a node that is either joining or leaving the cluster. These are nodes that are not part of the cluster (they do not have a corresponding node object). Instead, they are used to track the creation and deletion of nodes.
This resource is automatically managed by the operator.
For joining nodes, the deadline is used to delete the pending node if it fails to join before the deadline ends.

Example for GCP:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: PendingNode
metadata:
  name: pendingnode-sample
spec:
  providerID: "gce://<project-id>/<zone>/<instance-name>"
  groupID: "projects/<project-id>/zones/<zone>/instanceGroupManagers/<instance-group-name>"
  nodeName: "<kubernetes-node-name>"
  goal: Join
  deadline: "2022-07-04T08:33:18+00:00"
```

Example for Azure:

```yaml
apiVersion: update.edgeless.systems/v1alpha1
kind: PendingNode
metadata:
  name: pendingnode-sample
spec:
  providerID: "azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>/virtualMachines/<instance-id>"
  groupID: "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>"
  nodeName: "<kubernetes-node-name>"
  goal: Join
  deadline: "2022-07-04T08:33:18+00:00"
```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/constellation/node-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/constellation/node-operator:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Production deployment

The operator is deployed automatically during `constellation-init`.
Prerequisite for this is that cert-manager is installed.
cert-manager is also installed during `constellation-init`.
To deploy you can use the Helm chart at `/internal/constellation/helm/charts/edgeless/operators/constellation-operator`.
