terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.56.0"
    }
  }
}

locals {
  role_dashed     = var.role == "ControlPlane" ? "control-plane" : "worker"
  # migration: allow the old node group names to work since they were created without the uid
  # and without multiple node groups in mind
  # node_group: worker_default => name == "<base>-1-worker"
  # node_group: control_plane_default => name:  "<base>-control-plane"
  # new names:
  # node_group: foo, role: Worker => name == "<base>-worker-<uid>"
  # node_group: bar, role: ControlPlane => name == "<base>-control-plane-<uid>"
  group_uid       = random_id.uid.hex
  maybe_uid       = (var.node_group_name == "control_plane_default" || var.node_group_name == "worker_default") ? "" : "-${local.group_uid}"
  maybe_one       = var.node_group_name == "worker_default" ? "-1" : ""
  name = "${var.base_name}${local.maybe_one}-${local.role_dashed}${local.maybe_uid}"
}

resource "random_id" "uid" {
  byte_length = 4
}
resource "azurerm_lb_backend_address_pool" "backend_pool" {
  loadbalancer_id = var.loadbalancer_id
  name            = local.name
}

resource "azurerm_lb_probe" "health_probes" {
  for_each = { for port in var.ports : port.name => port }

  loadbalancer_id     = var.loadbalancer_id
  name                = each.value.name
  port                = each.value.port
  protocol            = each.value.protocol
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
  frontend_ip_configuration_name = "PublicIPAddress"
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.backend_pool.id]
  probe_id                       = each.value.id
  disable_outbound_snat          = true
}
