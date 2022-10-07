# Network encryption

Constellation encrypts all pod communication using the [container network interface (CNI)](https://github.com/containernetworking/cni).
To that end, Constellation deploys, configures, and operates the [Cilium](https://cilium.io/) CNI plugin.
Cilium provides [transparent encryption](https://docs.cilium.io/en/stable/gettingstarted/encryption) for all cluster traffic using either IPSec or [WireGuard](https://www.wireguard.com/).
Currently, Constellation only supports WireGuard as the encryption engine.
You can read more about the cryptographic soundness of WireGuard [in their white paper](https://www.wireguard.com/papers/wireguard.pdf).

Cilium is actively working on implementing a feature called [`host-to-host`](https://github.com/cilium/cilium/pull/19401) encryption mode for WireGuard.
With `host-to-host`, all traffic between nodes will be tunneled via WireGuard (host-to-host, host-to-pod, pod-to-host, pod-to-pod).
Until the `host-to-host` feature is released, Constellation enables `pod-to-pod` encryption.
This mode encrypts all traffic between Kubernetes pods using WireGuard tunnels.
Constellation uses an extended version of `pod-to-pod` called *strict* mode.

When using Cilium in the default setup but with encryption enabled, there is a [known issue](https://docs.cilium.io/en/v1.12/gettingstarted/encryption/#egress-traffic-to-not-yet-discovered-remote-endpoints-may-be-unencrypted)
that can cause pod-to-pod traffic to be unencrypted.
Constellation uses strict mode to mitigates this issue.
We change the default behavior of traffic that's destined for an unknown endpoint to not be send out in plaintext to instead being dropped.
The strict mode can distinguish between traffic that's send to an pod from traffic that's destined for an cluster-external endpoint, since it knows the pod's CIDR range.

The last remaining traffic that's not encrypted is traffic originating from hosts.
This mainly includes health checks from Kubernetes API server.
Also, traffic proxied over the API server via e.g. `kubectl port-forward` isn't encrypted.
