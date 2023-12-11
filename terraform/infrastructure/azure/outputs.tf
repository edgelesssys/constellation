output "out_of_cluster_endpoint" {
  value = local.out_of_cluster_endpoint
  description = "The external endpoint for the Kubernetes API server. This only varies from the `in_cluster_endpoint` when using an internal load balancer setup."
}

output "in_cluster_endpoint" {
  value = local.in_cluster_endpoint
  description = "The internal endpoint for the Kubernetes API server. This only varies from the `in_cluster_endpoint` when using an internal load balancer setup."
}

output "api_server_cert_sans" {
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
  value = local.uid
  description = "The UID of the cluster."
}

output "init_secret" {
  value     = random_password.initSecret.result
  sensitive = true
  description = "The init secret for the cluster."
}

output "attestation_url" {
  value = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : ""
  description = "The attestation URL for the cluster."
}

output "network_security_group_name" {
  value = azurerm_network_security_group.security_group.name
  description = "The name of the network security group."
}

output "loadbalancer_name" {
  value = azurerm_lb.loadbalancer.name
  description = "The name of the load balancer."
}


output "user_assigned_identity_client_id" {
  value = data.azurerm_user_assigned_identity.uaid.client_id
  description = "The client ID of the user assigned identity."
}

output "resource_group" {
  value = var.resource_group
  description = "The name of the resource group."
}

output "subscription_id" {
  value = data.azurerm_subscription.current.subscription_id
  description = "Azure subscription ID."
}

output "name" {
  value = local.name
  description = "The name of the cluster."
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
  description = "The CIDR block for the VPC subnet for nodes."
}
