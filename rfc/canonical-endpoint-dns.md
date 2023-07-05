# Canonical endpoint / Constellation with custom DNS name

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

1. start adding the fallback endpoint (and the custom domain, if set) to the SAN field of the kube-apiserver certificates
    - SAN should still contain everything it contains before
      - legacy public ip we use before switching to dns
      - `kubernetes`, `kubernetes.default`, `kubernetes.default.svc`, `kubernetes.default.svc.cluster.local`
      - 127.0.0.1, ::1, localhost
    - fallback endpoint (as specified above)
    - optional customer provided hostname (if set)
2. once every apiserver uses the extended SAN field, the cluster-internal advertise-address for the apiserver can be switched to the fallback endpoint
    - no-op on GCP (this is already the status quo)
    - ensure every existing kubelet uses the new endpoint
    - ensure newly joining kubelet use the new endpoint
    - k8s configmap uses new endpoint (`kube-system/kubeadm-config` -> `.data.ClusterConfiguration`)
3. switch id file `constellation-id.json` to use either fallback endpoint or customer endpoint (Q: which is better if both are available?)
4. switch setting in kubernetes config (`constellation-admin.conf`) to use either fallback endpoint or customer endpoint (Q: which is better if both are available?)
    - in existing clusters: patch config
    - in new clusters: write correct endpoint on `constellation init`


### Details

Some design details that are worth specifying in advance.

#### How does a Constellation node learn the canonical endpoint?

- Cloud metadata attached to every node

#### Should we still try to discover individual nodes?

Debugd and bootstrapper currently list individual instances and connect to the private vpc ips of the instances. Instead, we might be able to use the canonical endpoint and simply target the loadbalancer.
Possible issues: this may fail on GCP, since every control-plane node adds the public ip to the primary network interface.
