# RFC XXX: Trusted Kubernetes Container Images

Kubernetes control plane images should be verified by the Constellation installation.

## The Problem

When we bootstrap the Constellation Kubernetes cluster, `kubeadm` places a set
of static pods for the control plane components into the filesystem. The
manifests refer to images in a registry beyond the users' control, and the
container image content is not reproducible.

This is obviously a trust issue, because the Kubernetes control plane is
part of Constellation's TCB, but it is also a problem when Constellation is set
up in a restricted environment where this repo is not available.

## Requirements

1. In a default installation, Constellation must verify Kubernetes control plane images.

Out of scope:

- Customization of the image repository or image content.
  - This is orthogonal to image verification and will be subject of a separate
    RFC.
- Reproducibility from github.com/kubernetes/kubernetes to registry.k8s.io and
  the associated chain of trust.
  - It is not clear whether Kubernetes images can be reproduced at all [1].
    Either way, the likely threat is a machine-in-the-middle attack between
    Constellation and registry.k8s.io. A desirable addition to this proposal
    could be verification of image signatures [2].
- Container registry trust & CA certificates.
  - This is also orthogonal to image verification.

[1]: https://github.com/kubernetes/kubernetes/blob/master/build/README.md#reproducibility
[2]: https://kubernetes.io/docs/tasks/administer-cluster/verify-signed-artifacts/

## Solution

Kubernetes control plane images are going to be pinned by a hash, which is verified by
the CRI. Image hashes are added to the Constellation codebase when support for
a new version is added. During installation, the `kubeadm` configuration is
modified so that images are pinned.

### Image Hashes

We are concerned with the following control plane images (tags for v1.27.7):

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
  KubernetesImages: {
    "kube-apiserver": "registry.k8s.io/kube-apiserver:v1.27.7@sha256:<...>", 
    "kube-controller-manager": "registry.k8s.io/kube-controller-manager:v1.27.7@sha256:<...>", 
    "kube-scheduler": "registry.k8s.io/kube-scheduler:v1.27.7@sha256:<...>", 
    "etcd": "registry.k8s.io/etcd:3.5.9-0@sha256:<...>",
    "coredns": "registry.k8s.io/coredns/coredns:v1.10.1@sha256:<...>",
  },
}
```

### Cluster Init

During cluster initialization, we need to tell `kubeadm` that we want to use
the embedded image references instead of the default ones. For that, we
populate the
[`InitConfiguration.Patches`](https://pkg.go.dev/k8s.io/kubernetes@v1.27.7/cmd/kubeadm/app/apis/kubeadm/v1beta3#InitConfiguration)
with a list of [JSON Patch](https://datatracker.ietf.org/doc/html/rfc6902)
files that replace the container image with the pinned alternative.

The patches need to be written to the stateful filesystem by the
bootstrapper. This is very similar to `components.Component`, which also
place Kubernetes-related data onto the filesystem:

```go
type Component struct {
  URL         string
  Hash        string
  InstallPath string
  Extract     bool
}
```

Components are handled by the installer, where the convention currently expects
HTTP URLs that are to be downloaded. We can extend this by allowing other forms
of URI schemes:
[data URLs](https://developer.mozilla.org/en-US/docs/web/http/basics_of_http/data_urls).
A patch definition as Component would look like this:

```go
patch := &components.Component{
  URL: "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL215LWNvbnRyb2wtcGxhbmUtaW1hZ2U6djEuMjcuN0BzaGEyNTY6Li4uIn1dCg=="
  InstallPath: "/opt/kubernetes/patches/kube-apiserver+json.json"
  // InstallPath and Extract deliberately left empty.
}
```

This method does not work for coredns, which can only be customized with the
[`ClusterConfiguration.DNS`](https://pkg.go.dev/k8s.io/kubernetes@v1.27.7/cmd/kubeadm/app/apis/kubeadm/v1beta3#ClusterConfiguration)
section. However, we can split the image into repository, path and tag+hash
and use these values as `DNS.ImageMeta`.

### Upgrade

The upgrade agent currently receives a kubeadm URI and hash, and internally
assembles this to a `Component`. We change the upgrade proto to accept
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
