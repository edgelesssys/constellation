terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.56.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

locals {
  role_dashed = var.role == "ControlPlane" ? "control-plane" : "worker"
  tags = merge(
    var.tags,
    { constellation-role = local.role_dashed },
    { constellation-node-group = var.node_group_name },
  )
  # migration: allow the old node group names to work since they were created without the uid
  # and without multiple node groups in mind
  # node_group: worker_default => name == "<base>-worker"
  # node_group: control_plane_default => name:  "<base>-control-plane"
  # new names:
  # node_group: foo, role: Worker => name == "<base>-worker-<uid>"
  # node_group: bar, role: ControlPlane => name == "<base>-control-plane-<uid>"
  group_uid = random_id.uid.hex
  maybe_uid = (var.node_group_name == "control_plane_default" || var.node_group_name == "worker_default") ? "" : "-${local.group_uid}"
  name      = "${var.base_name}-${local.role_dashed}${local.maybe_uid}"
}

resource "random_id" "uid" {
  byte_length = 4
}
resource "random_password" "password" {
  length      = 16
  min_lower   = 1
  min_upper   = 1
  min_numeric = 1
  min_special = 1
}

resource "azurerm_linux_virtual_machine_scale_set" "scale_set" {
  name                            = local.name
  resource_group_name             = var.resource_group
  location                        = var.location
  sku                             = var.instance_type
  instances                       = var.instance_count
  admin_username                  = "adminuser"
  admin_password                  = random_password.password.result
  overprovision                   = false
  provision_vm_agent              = false
  vtpm_enabled                    = true
  disable_password_authentication = false
  upgrade_mode                    = "Manual"
  secure_boot_enabled             = var.secure_boot
  source_image_id                 = var.image_id
  tags                            = local.tags
  zones                           = var.zones
  identity {
    type         = "UserAssigned"
    identity_ids = [var.user_assigned_identity]
  }

  boot_diagnostics {}

  dynamic "os_disk" {
    for_each = var.confidential_vm ? [1] : [] # if confidential_vm is true
    content {
      security_encryption_type = "VMGuestStateOnly"
      caching                  = "ReadWrite"
      storage_account_type     = "Premium_LRS"
    }
  }
  dynamic "os_disk" {
    for_each = var.confidential_vm ? [] : [1] # else
    content {
      caching              = "ReadWrite"
      storage_account_type = "Premium_LRS"
    }
  }

  data_disk {
    storage_account_type = var.state_disk_type
    disk_size_gb         = var.state_disk_size
    caching              = "ReadWrite"
    lun                  = 0
  }

  network_interface {
    name                      = "node-network"
    primary                   = true
    network_security_group_id = var.network_security_group_id

    ip_configuration {
      name                                   = "node-network"
      primary                                = true
      subnet_id                              = var.subnet_id
      load_balancer_backend_address_pool_ids = var.backend_address_pool_ids
    }
  }

  lifecycle {
    ignore_changes = [
      instances,       # required. autoscaling modifies the instance count externally
      source_image_id, # required. update procedure modifies the image id externally
    ]
  }
}
