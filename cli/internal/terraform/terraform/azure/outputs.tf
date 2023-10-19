output "out_of_cluster_endpoint" {
  value = local.out_of_cluster_endpoint
}

output "in_cluster_endpoint" {
  value = local.in_cluster_endpoint
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
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}

output "attestationURL" {
  value = var.create_maa ? azurerm_attestation_provider.attestation_provider[0].attestation_uri : ""
}

output "network_security_group_name" {
  value = azurerm_network_security_group.security_group.name
}

output "loadbalancer_name" {
  value = azurerm_lb.loadbalancer.name
}


output "user_assigned_identity_client_id" {
  value = data.azurerm_user_assigned_identity.uaid.client_id
}

output "resource_group" {
  value = var.resource_group
}

output "subscription_id" {
  value = data.azurerm_subscription.current.subscription_id
}

output "name" {
  value = local.name
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
}
