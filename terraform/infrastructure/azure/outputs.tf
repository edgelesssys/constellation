# Outputs common to all CSPs

output "out_of_cluster_endpoint" {
  value       = local.out_of_cluster_endpoint
  description = "External endpoint for the Kubernetes API server. Only varies from the `in_cluster_endpoint` when using an internal load balancer."
}

output "in_cluster_endpoint" {
  value       = local.in_cluster_endpoint
  description = "Internal endpoint for the Kubernetes API server."
}

output "extra_api_server_cert_sans" {
  value = sort(
    distinct(
      concat(
        [
          local.in_cluster_endpoint,
          local.out_of_cluster_endpoint,
        ],
        var.custom_endpoint == "" ? [] : [var.custom_endpoint],
        var.internal_load_balancer ? [] : [local.wildcard_lb_dns_name],
      )
    )
  )
  description = "List of Subject Alternative Names (SANs) for the API server certificate."
}

output "uid" {
  value       = local.uid
  description = "Unique Identifier (UID) of the cluster."
}

output "init_secret" {
  value       = random_password.init_secret.result
  sensitive   = true
  description = "Initialization secret to authenticate the bootstrapping node."
}

output "name" {
  value       = local.name
  description = "Unique name of the Constellation cluster, comprised by name and UID."
}

output "ip_cidr_node" {
  value       = local.cidr_vpc_subnet_nodes
  description = "CIDR block of the node network."
}

# Azure-specific outputs

output "attestation_url" {
  value       = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : ""
  description = "URL of the cluster's Microsoft Azure Attestation (MAA) provider."
}

output "network_security_group_name" {
  value       = azurerm_network_security_group.security_group.name
  description = "Name of the cluster's network security group."
}

output "loadbalancer_name" {
  value       = azurerm_lb.loadbalancer.name
  description = "Name of the cluster's load balancer."
}

output "user_assigned_identity_client_id" {
  value       = data.azurerm_user_assigned_identity.uaid.client_id
  description = "Client ID of the user assigned identity used within the cluster."
}

output "resource_group" {
  value       = var.resource_group
  description = "Name of the resource group the cluster resides in."
}

output "subscription_id" {
  value       = data.azurerm_subscription.current.subscription_id
  description = "ID of the Azure subscription the cluster resides in."
}
