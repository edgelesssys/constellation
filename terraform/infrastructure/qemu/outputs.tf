# Outputs common to all CSPs

output "out_of_cluster_endpoint" {
  value       = module.node_group["control_plane_default"].instance_ips[0]
  description = "External endpoint for the Kubernetes API server. Only varies from the `in_cluster_endpoint` when using an internal load balancer."
}

output "in_cluster_endpoint" {
  value       = module.node_group["control_plane_default"].instance_ips[0]
  description = "Internal endpoint for the Kubernetes API server."
}

output "api_server_cert_sans" {
  value       = sort(concat([module.node_group["control_plane_default"].instance_ips[0]], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
  description = "List of Subject Alternative Names (SANs) for the API server certificate."
}

output "uid" {
  value       = "qemu" // placeholder
  description = "Unique Identifier (UID) of the cluster."
}

output "init_secret" {
  value       = random_password.init_secret.result
  sensitive   = true
  description = "Initialization secret to authenticate the bootstrapping node."
}

output "name" {
  value       = "${var.name}-qemu" // placeholder, as per "uid" output
  description = "Unique name of the Constellation cluster, comprised by name and UID."
}

output "ip_cidr_node" {
  value       = local.cidr_vpc_subnet_nodes
  description = "CIDR block of the node network."
}

# QEMU-specific outputs

output "validate_constellation_kernel" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_kernel) > 0
    error_message = "constellation_kernel must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
  description = "Validation placeholder. Do not consume as output."
}

output "validate_constellation_initrd" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_initrd) > 0
    error_message = "constellation_initrd must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
  description = "Validation placeholder. Do not consume as output."
}

output "validate_constellation_cmdline" {
  value = null
  precondition {
    condition     = var.constellation_boot_mode != "direct-linux-boot" || length(var.constellation_cmdline) > 0
    error_message = "constellation_cmdline must be set if constellation_boot_mode is 'direct-linux-boot'"
  }
  description = "Validation placeholder. Do not consume as output."
}
