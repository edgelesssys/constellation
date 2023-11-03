output "backendpool_id" {
  value       = azurerm_lb_backend_address_pool.backend_pool.id
  description = "The ID of the created backend pool."
}
