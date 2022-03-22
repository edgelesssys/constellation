# CLI to spawn a confidential kubernetes cluster

## Usage

0. (optional) replace the responsible in `cli/cmd/defaults.go` with yourself.
1. Build the CLI and authenticate with <AWS/Azure/GCP> according to the [README.md](https://github.com/edgelesssys/constellation-coordinator/blob/main/README.md#cloud-credentials).
2. Execute `constellation create <aws/azure/gcp> 2 <4xlarge|n2d-standard-2>`.
3. Execute `wg genkey | tee privatekey | wg pubkey > publickey` to generate a WireGuard keypair.
4. Execute `constellation init --publickey publickey`. Since the CLI waits for all nodes to be ready, this step can take up to 5 minutes.
5. Use the output from `constellation init` and the wireguard template below to create `/etc/wireguard/wg0.conf`, then execute `wg-quick up wg0`.
6. Execute `export KUBECONFIG=<path/to/admin.conf>`.
7. Use `kubectl get nodes` to inspect your cluster.
8. Execute `constellation terminate` to terminate your Constellation.

```bash
[Interface]
Address = <address from the init output>
PrivateKey = <your base64 encoded private key>
ListenPort = 51820

[Peer]
PublicKey = <public key from the init output>
AllowedIPs = 10.118.0.1/32 # IP set on the peer's wg interface
Endpoint = <public IPv4 address from the activated coordinator>:51820  # address where the peer listens on
PersistentKeepalive = 10
```

Note: Skip the manual configuration of WireGuard by executing Step 2 as root. Then, replace steps 4 and 5 with `sudo constellation init --privatekey <path/to/your/privatekey>`. This will automatically configure a new WireGuard interface named wg0 with the coordinator as peer.
