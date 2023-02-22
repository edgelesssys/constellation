terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.41.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.4.3"
    }
  }
}

resource "random_password" "password" {
  length      = 16
  min_lower   = 1
  min_upper   = 1
  min_numeric = 1
  min_special = 1
}

resource "azurerm_linux_virtual_machine_scale_set" "scale_set" {
  name                            = var.name
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
  tags                            = var.tags

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
