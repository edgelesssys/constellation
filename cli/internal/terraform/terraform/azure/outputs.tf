output "ip" {
  value = azurerm_public_ip.loadbalancer_ip.ip_address
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
