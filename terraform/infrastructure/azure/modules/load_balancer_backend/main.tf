terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.27.0"
    }
  }
}

resource "azurerm_lb_backend_address_pool" "backend_pool" {
  loadbalancer_id = var.loadbalancer_id
  name            = var.name
}

resource "azurerm_lb_probe" "health_probes" {
  for_each = { for port in var.ports : port.name => port }

  loadbalancer_id     = var.loadbalancer_id
  name                = each.value.name
  port                = each.value.port
  protocol            = each.value.health_check_protocol
  request_path        = each.value.path
  interval_in_seconds = 5
}

resource "azurerm_lb_rule" "rules" {
  for_each = azurerm_lb_probe.health_probes

  loadbalancer_id                = var.loadbalancer_id
  name                           = each.value.name
  protocol                       = "Tcp"
  frontend_port                  = each.value.port
  backend_port                   = each.value.port
  frontend_ip_configuration_name = var.frontend_ip_configuration_name
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.backend_pool.id]
  probe_id                       = each.value.id
  disable_outbound_snat          = true
}
