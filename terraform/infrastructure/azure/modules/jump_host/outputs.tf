output "ip" {
  value       = azurerm_linux_virtual_machine.jump_host.public_ip_address
  description = "Public IP of the jump host."
}
