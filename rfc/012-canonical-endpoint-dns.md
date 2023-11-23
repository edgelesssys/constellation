# RFC 012: Canonical endpoint / Constellation with custom DNS name

Customers may want to use a DNS name for the cluster endpoint instead of the public ip address.
The public ip may be bound to a zonal loadbalancer (AWS), limiting the availability of the cluster to a single zone.
Additionally, the public ip address may change if the ip is freed and re-allocated accidentally.
For branding or organizational reasons, customers may want to use a custom DNS name instead of a bare ip address or csp-provided hostname.
Additionally, a dual-stack setup with IPv4 and IPv6 can be achieved more easily with a DNS name.

## State of loadbalancing / DNS on the relevant CSPs

### AWS

On AWS, each loadbalancer [has a public dns name](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-internet-facing-load-balancers.html) in the format `name-1234567890.region.elb.amazonaws.com`.
This hostname can be used as a default endpoint for Constellation for interactions with the bootstrapper, cluster joining and the kubernetes API.

Additionally, customers may register custom domains and point them to the loadbalancer hostname.

### Azure

On Azure, each public ip [can have a public dns name](https://learn.microsoft.com/en-us/azure/virtual-network/ip-services/public-ip-addresses#dns-name-label) in the format `name-1234567890.location.cloudapp.azure.com`.
This hostname can be used as a default endpoint for Constellation for interactions with the bootstrapper, cluster joining and the kubernetes API.

Additionally, customers may register custom domains and point them to either the public ip hostname or address.

### GCP

On Google Cloud, neither loadbalancers nor public ips have a default, public dns name.

We have to use the (single) public ipv4 address as the default endpoint for Constellation for interactions with the bootstrapper, cluster joining and the kubernetes API.

Additionally, customers may register custom domains and point them to the public ip address.

### Conclusion

**Fallback endpoint**:

- loadbalancer dns name on AWS (`name-1234567890.region.elb.amazonaws.com`)
- public ip dns name on Azure (`name-1234567890.location.cloudapp.azure.com`)
- public ip on GCP (`203.0.113.1`)

**(Optional) customer provided dns**:

- any custom domain that points to the fallback endpoint

## Steps for enabling DNS (aka: "canonical endpoint")

The steps for enabling DNS are split up in phases.
Each phase has to happen in a separate release, since the steps expect a certain state of the cluster that is reached when the previous phase is completed.

### Phase 1: Discover and add the new endpoints to the SAN field of the apiserver certificates

Start adding the fallback endpoint (and the custom domain, if set) to the SAN field of the kube-apiserver certificates.
- SAN should still contain everything it contains before
  - legacy public ip we use before switching to dns
  - `kubernetes`, `kubernetes.default`, `kubernetes.default.svc`, `kubernetes.default.svc.cluster.local`
  - 127.0.0.1, ::1, localhost
- fallback endpoint (as specified above)
- optional customer provided hostname (if set)
For new clusters, the cert-sans field should already be set correctly when calling `kubeadm init` (using the `ClusterConfiguration` yaml).
#### One-time step (migration for existing clusters):

In existing clusters, this can be achieved by patching the `kube-system/kubeadm-config` -> `.data.ClusterConfiguration.certSANs` field.
Newly joining nodes will use the new endpoint, while existing nodes will continue to use the old endpoint until they are replaced.
During the upgrade, each control-plane node will be replaced by a node with the new endpoints in the SAN field.
The new endpoint names can be retrieved from the cloud metadata attached to every node. The CLI can retrieve the new endpoint names from terraform output (fallback endpoint) and directly from the constellation-conf.yaml (custom domain).

### Phase 2: Use new endpoint for `ClusterConfiguration.ControlPlaneEndpoint`, switch id file and kubernetes config

Once every apiserver uses the new SAN field, the cluster-internal control-plane endpoint for the apiserver can be switched to the fallback endpoint.
For new clusters, the control-plane endpoint in `ClusterConfiguration.ControlPlaneEndpoint` should be set to the fallback endpoint when calling kubeadm init.
Write fallback endpoint to `constellation-id.json` and `constellation-admin.conf` on `constellation init`.

#### One-time step  (migration for existing clusters):

In existing clusters, this can be achieved by patching the `kube-system/kubeadm-config` -> `.data.ClusterConfiguration.ControlPlaneEndpoint` field.
The existing configuration files in the user's workspace directory should be updated to use the fallback endpoint.
- switch id file `constellation-id.json` to use fallback endpoint
- switch setting in kubernetes config (`constellation-admin.conf`) to use fallback endpoint

### Details

Some design details that are worth specifying in advance.

#### How does a Constellation node learn the canonical endpoint?

- Cloud metadata attached to every node

#### Should we still try to discover individual nodes?

Debugd and bootstrapper currently list individual instances and connect to the private vpc ips of the instances. Instead, we might be able to use the canonical endpoint and simply target the loadbalancer.
Possible issues: this may fail on GCP, since every control-plane node adds the public ip to the primary network interface.
