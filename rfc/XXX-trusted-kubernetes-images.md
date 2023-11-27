# RFC XXX: Trusted Kubernetes Images

Kubernetes control plane images should be verified by the Constellation installation.

## The Problem

When we bootstrap the Constellation Kubernetes cluster, `kubeadm` places a set
of static pods for the control plane components into the filesystem. The
manifests refer to images in a registry beyond the users' control, and the
image content is not reproducible.

This is obviously a trust issue, because the Kubernetes control plane is
part of Constellation's TCB, but it is also a problem when Constellation is set
up in a restricted environment where this repo is not available.

## Requirements

1. In a default installation, Constellation must verify Kubernetes control plane images.
2. Users must be able to override the image repository for Kubernetes control plane images.
3. Users must be able to override the verification mechanism for Kubernetes control plane images.

Out of scope:

- reproducibility from github.com/kubernetes/kubernetes to registry.k8s.io and
  the associated chain of trust
- container registry trust & CA certificates

## Solution

Kubernetes control plane images are going to be pinned by hash and verified by
the CRI. In a default installation, we take the upstream images from
registry.k8s.io and ship hashes with the CLI. Users can override the image
per control plane component.

### Image Hashes

We are concerned with the following control plane images (for v1.27.7):

- registry.k8s.io/kube-apiserver:v1.27.7
- registry.k8s.io/kube-controller-manager:v1.27.7
- registry.k8s.io/kube-scheduler:v1.27.7
- registry.k8s.io/coredns/coredns:v1.10.1
- registry.k8s.io/etcd:3.5.9-0

When a new Kubernetes version is added to `/internal/versions/versions.go`, we
generate the corresponding list of images with `kubeadm config images list` and
probe their hashes on registry.k8s.io. Generating the list of images this way
must happen offline to prevent `kubeadm` from being clever. These hashes are
added to `versions.go` as a mapping from component to pinned image:

```golang
V1_27: {
  ClusterVersion: "v1.27.7",
  // ...
  ControlPlaneImages: {
    "kube-apiserver": "registry.k8s.io/kube-apiserver:v1.27.7@sha256:<...>", 
    "kube-controller-manager": "registry.k8s.io/kube-controller-manager:v1.27.7@sha256:<...>", 
    "kube-scheduler": "registry.k8s.io/kube-scheduler:v1.27.7@sha256:<...>", 
    "etcd": "registry.k8s.io/etcd:3.5.9-0@sha256:<...>",
    "coredns": "registry.k8s.io/coredns/coredns:v1.10.1@sha256:<...>",
  },
}
```

### Configuration Options

Users can override the images in their `constellation-conf.yaml`, using the
same nested structure as `versions.go` (*to be elaborated on*).

### Cluster Init

During cluster initialization, we need to tell `kubeadm` that we want to use
the embedded image references instead of the default ones. For that, we
populate the
[`InitConfiguration.Patches`](https://pkg.go.dev/k8s.io/kubernetes@v1.27.7/cmd/kubeadm/app/apis/kubeadm/v1beta3#InitConfiguration)
with a list of patch files that replace the container image with the pinned
alternative.

This method does not work for coredns, which can only be customized with the
[`ClusterConfiguration.DNS`](https://pkg.go.dev/k8s.io/kubernetes@v1.27.7/cmd/kubeadm/app/apis/kubeadm/v1beta3#ClusterConfiguration)
section. It's a bit of a hack, but we can parse the registry part and the
tag/hash part from the full image and set it in that `DNS` section of the
config.

The patches need to be materialized to the stateful filesystem by the
bootstrapper. As there are a lot of similarities between these patches and
`components.Components`, we could extend that definition to also allow inline
file content as an alternative to remote URI and hash. Alternatively, we could
move the creation of the kubeadm configuration to the cli and let the
bootstrapper only override fields that are discovered in the cluster, like
cloud provider specifics.

### Upgrade

The upgrade agent currently receives a kubeadm URI and hash, and internally
assembles this to a `Component`. We could change the upgrade proto to accept
a full `components.Components`, which then would also include the new patches.
The components would be populated from the ConfigMap, as is already the case.

The CoreDNS config would need to be updated in the `kube-system/kubeadm-config`
ConfigMap.

## Alternatives Considered

### Exposing more of KubeadmConfig

We could allow users to supply their own patches to `KubeadmConfig` for finer
control over the installation. We don't want to do this because:

1. It does not solve the problem of image verification - we'd still need to
   derive image hashes from somewhere.
2. It's easy to accidentally leave charted territory when applying config
   overrides, and responsibilities are unclear in that case: should users be
   allowed to configure network, etcd, etc.?
3. The way Kubernetes exposes the configuration is an organically grown mess:
   registries are now in multiple structs, path names are hard-coded to some
   extent and versions come from somewhere else entirely (cf.
   kubernetes/kubernetes#102502).

### Ship the container images with the OS

We could bundle all control plane images in our OS image and configure kubeadm
to never pull images. This would make Constellation independent of external
image resources at the expense of flexibility: overriding the control plane
images in development setups would be harder, and we would not be able to
support user-provided images anymore.
