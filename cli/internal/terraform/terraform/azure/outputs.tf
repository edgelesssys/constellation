output "ip" {
  value = azurerm_public_ip.loadbalancer_ip.ip_address
}

output "api_server_cert_sans" {
  value = sort(concat([azurerm_public_ip.loadbalancer_ip.ip_address, local.wildcard_lb_dns_name], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
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

// new
output "network_security_group_name" {
  value = azurerm_network_security_group.security_group.name
}

output "loadbalancer_name" {
  value = azurerm_lb.loadbalancer.name
}


output "user_assigned_identity" {
  value = var.user_assigned_identity
}
