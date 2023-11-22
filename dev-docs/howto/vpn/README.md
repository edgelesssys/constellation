# Constellation VPN

This Helm chart deploys a VPN server to your Constellation cluster.

## Installation

1. Create and populate the configuration.

   ```sh
   helm inspect values . >config.yaml
   ```

2. Install the Helm chart.

   ```sh
   helm install -f config.yaml vpn . 
   ```

3. Follow the post-installation instructions displayed by the CLI.

## Architecture

The VPN server is deployed as a `StatefulSet` to the cluster. It hosts the VPN frontend component, which is responsible for relaying traffic between the pod and the on-prem network, and the routing components that provide access to Constellation resources. The frontend supports IPSec and Wireguard.

The VPN frontend is exposed with a public LoadBalancer to be accessible from the on-prem network. Traffic that reaches the VPN server pod is split into two categories: pod IPs and service IPs.

The pod IP range is NATed with an iptables rule. On-prem worklaods can establish connections to a pod IP, but the Constellation workloads will see the client IP translated to that of the VPN frontend pod.

The service IP range is handed to a transparent proxy running in the VPN frontend pod, which relays the connection to a backend pod. This is necessary because of the load-balancing mechanism of Cilium, which assumes service IP traffic to originate from the Constellation cluster itself. As for pod IP ranges, Constellation pods will only see the translated client address.

## Limitations

* Service IPs need to be proxied by the VPN frontend pod. This is a single point of failure, and it may become a bottleneck.
* IPs are NATed, so the Constellation pods won't see the real on-prem IPs.
* NetworkPolicy can't be applied selectively to the on-prem ranges.
* No connectivity from Constellation to on-prem workloads.
