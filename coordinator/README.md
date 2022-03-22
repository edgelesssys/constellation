# Coordinator
The Coordinator/Node architecture is inspired by K8s. The Coordinator resembles kube-apiserver, while the Nodes resemble kubelets.

All peers serve the *pubapi*, which is exposed publicly. Once initialized, a Coordinator additionally serves the *vpnapi*, which is exposed inside the VPN.

## pubapi
The pubapi provides APIs that are either required from outside the cluster or inside the cluster before the VPN is established.

pubapi connections are protected by attested TLS (atls): the client verifies the server. The server does *not* verify the client. The APIs must be designed to form a chain of trust, so that no additional verification is needed.

For example, to activate all peers in a new cluster, there's a chain of trust from the CLI via the Coordinator to the Nodes:
* CLI calls ActivateAsCoordinator
* Coordinator calls ActivateAsNode

If new Nodes shall be added to the cluster later, they must not activate themselves by the Coordinator, but have to ask it to activate them (using ActivateAdditionalNodes). This way, the chain of trust is preserved.

Try to keep the pubapi small. Prefer adding new functionality to the vpnapi instead.

## vpnapi
The vpnapi is served by the Coordinator and can be used by the Nodes after they joined the VPN. Most importantly, the Nodes use it to get updates about added/removed/changed peers.

A Node regularly requests an update from a Coordinator. This is required for fault tolerance: if a Node cannot be provided with updated peer infos at one time, e.g., because of a network issue, it will continue to try and will eventually converge towards the desired state. (Note that this may not be fully implemented yet.)

Peer updates are versioned. The Node sends its last known version number and the Coordinator responds with the current version number and with the updated peers if needed. Currently, updates contain full peer info, but may be changed to incremental in the future.

## Core
Both APIs use the Core to fulfill the requests. The Core implements the core logic of a peer. It doesn't know the APIs and should be kept free of any gRPC or other client/server code.

## Naming convention
We have defined additional naming conventions for the coordinator.

### Entities
* Coordinator: the thing activated by ActivateAsCoordinator
* Node: the things activated by ActivateAsNode
* peer: either Coordinator or Node
* admin: the user who calls ActivateAsCoordinator

### Network
IP addresses:
* ip: numeric IP address
* host: either IP address or hostname
* endpoint: host+port

Interfaces using the addresses:
* public
* vpn

Usage: variable namings should then be entityInterfaceKind, e.g.
* coordinatorPublicEndpoint
* nodeVPNIP

Entity and/or interface are omitted if not relevant for function contract.

### Keys
Kinds:
* key: symmetric key
* pubKey: public key
* privKey: private key

Purpose:
* *entity*
* vpn
* *entity*VPN

Example:
* nodeVPNPubKey
