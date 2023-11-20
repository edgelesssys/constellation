# Constellation VPN

This Helm chart deploys a VPN server to your Constellation cluster.

## Prerequisites

* Constellation >= v2.14.0
* A publicly routable VPN endpoint on premises that supports IPSec in IKEv2
  tunnel mode with NAT traversal enabled.
* A list of on-prem CIDRs that should be reachable from Constellation.

## Setup

1. Configure Cilium to route services for the VPN (see [Architecture](#architecture) for details).
   * Edit the Cilium config: `kubectl -n kube-system edit configmap cilium-config`.
   * Set the config item `enable-sctp: "true"`.
   * Restart the Cilium agents: `kubectl -n kube-system rollout restart daemonset/cilium`.

2. Create the Constellation VPN configuration file.

   ```sh
   helm inspect values . >config.yaml
   ```

3. Populate the Constellation VPN configuration file. At least the following
   need to be configured:
   * The list of on-prem CIDRs (`peerCIDRs`).
   * The `ipsec` subsection.

4. Install the Helm chart.

   ```sh
   helm install -f config.yaml vpn . 
   ```

5. Configure the on-prem gateway with Constellation's pod and service CIDR
   (see `config.yaml`).

## Things to try

Ask CoreDNS about its own service IP:

```sh
dig +notcp @10.96.0.10 kube-dns.kube-system.svc.cluster.local
```

Ask the Kubernetes API server about its wellbeing:

```sh
curl --insecure https://10.96.0.1:6443/healthz
```

Ping a pod:

```sh
ping $(kubectl get pods vpn-frontend-0 -o go-template --template '{{ .status.podIP }}')
```

## Architecture

The VPN server is deployed as a `StatefulSet` to the cluster. It hosts the VPN
frontend component, which is responsible for relaying traffic between the pod
and the on-prem network over an IPSec tunnel.

The VPN frontend is exposed with a public LoadBalancer so that it becomes
accessible from the on-prem network.

An init container sets up IP routes on the frontend host and inside the
frontend pod. All routes are bound to the frontend pod's lxc interface and thus
deleted together with it.

A VPN operator deployment is added that configures the `CiliumEndpoint` with
on-prem IP ranges, thus configuring routes on non-frontend hosts. The endpoint
shares the frontend pod's lifecycle.

In Cilium's default configuration, service endpoints are resolved in cgroup
eBPF hooks that are not applicable to VPN traffic. We force Cilium to apply
service NAT at the LXC interface by enabling SCTP support.

## Limitations

* VPN traffic is handled by a single pod, which may become a bottleneck.
* Frontend pod restarts / migrations invalidate IPSec connections.
* Only pre-shared key authentication is supported.
