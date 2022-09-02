# Configuration file

Constellation CLI reads all configuration options from `constellation-conf.yaml`.

> The Constellation CLI can generate a default configuration file. This should be the preferred way, so that the configuration matches the used CLI version.

A sample configuration for a Constellation cluster on Azure looks like this:

```yaml
version: v1 # Schema version of this configuration file.
autoscalingNodeGroupMin: 1 # Minimum number of worker nodes in autoscaling group.
autoscalingNodeGroupMax: 10 # Maximum number of worker nodes in autoscaling group.
stateDiskSizeGB: 30 # Size (in GB) of a node's disk to store the non-volatile state.
# Ingress firewall rules for node network.
ingressFirewall:
    - name: bootstrapper # Name of rule.
      description: bootstrapper default port # Description for rule.
      protocol: tcp # Protocol, such as 'udp' or 'tcp'.
      iprange: 0.0.0.0/0 # CIDR range for which this rule is applied.
      fromport: 9000 # Start port of a range.
      toport: 0 # End port of a range, or 0 if a single port is given by fromport.
    - name: ssh # Name of rule.
      description: SSH # Description for rule.
      protocol: tcp # Protocol, such as 'udp' or 'tcp'.
      iprange: 0.0.0.0/0 # CIDR range for which this rule is applied.
      fromport: 22 # Start port of a range.
      toport: 0 # End port of a range, or 0 if a single port is given by fromport.
    - name: nodeport # Name of rule.
      description: NodePort # Description for rule.
      protocol: tcp # Protocol, such as 'udp' or 'tcp'.
      iprange: 0.0.0.0/0 # CIDR range for which this rule is applied.
      fromport: 30000 # Start port of a range.
      toport: 32767 # End port of a range, or 0 if a single port is given by fromport.
    - name: kubernetes # Name of rule.
      description: Kubernetes # Description for rule.
      protocol: tcp # Protocol, such as 'udp' or 'tcp'.
      iprange: 0.0.0.0/0 # CIDR range for which this rule is applied.
      fromport: 6443 # Start port of a range.
      toport: 0 # End port of a range, or 0 if a single port is given by fromport.
# Supported cloud providers and their specific configurations.
provider:
    # Configuration for Azure as provider.
    azure:
        subscription: "" # Subscription ID of the used Azure account. See: https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-subscription
        tenant: "" # Tenant ID of the used Azure account. See: https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id#find-your-azure-ad-tenant
        location: "" # Azure datacenter region to be used. See: https://docs.microsoft.com/en-us/azure/availability-zones/az-overview#azure-regions-with-availability-zones
        image: /subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation-coreos/versions/0.0.1659453699 # Machine image used to create Constellation nodes.
        stateDiskType: StandardSSD_LRS # Type of a node's state disk. The type influences boot time and I/O performance. See: https://docs.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison
        # Expected confidential VM measurements.
        measurements:
            11: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
            12: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
        userAssignedIdentity: "" # Authorize spawned VMs to access Azure API. See: https://docs.edgeless.systems/constellation/getting-started/install#authorization
kubernetesVersion: "1.24" # Kubernetes version installed in the cluster.

# # Egress firewall rules for node network.
# egressFirewall:
#     - name: rule#1 # Name of rule.
#       description: the first rule # Description for rule.
#       protocol: tcp # Protocol, such as 'udp' or 'tcp'.
#       iprange: 0.0.0.0/0 # CIDR range for which this rule is applied.
#       fromport: 443 # Start port of a range.
#       toport: 443 # End port of a range, or 0 if a single port is given by fromport.

# # Create SSH users on Constellation nodes.
# sshUsers:
#     - username: Alice # Username of new SSH user.
#       publicKey: ssh-rsa AAAAB3NzaC...5QXHKW1rufgtJeSeJ8= alice@domain.com # Public key of new SSH user.
```

## Required customizations

Most options of a generated configuration can be kept at their default values. However, you must edit some cloud provider options.

### Azure

Set the `subscription` and `tenant` IDs of your subscription.

Set the `userAssignedIdentity` that you [created for Constellation](../getting-started/install.md#azure).

### GCP

Set the `project` that you want to use for your Constellation cluster.
