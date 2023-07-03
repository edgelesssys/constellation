output "public_ip" {
  value      = azurerm_public_ip.main.ip_address
  sensitive  = false
  depends_on = [azurerm_public_ip.main]
}

output "ssh_private_key" {
  value      = tls_private_key.ssh_key.private_key_openssh
  sensitive  = true
  depends_on = [tls_private_key.ssh_key]
}
