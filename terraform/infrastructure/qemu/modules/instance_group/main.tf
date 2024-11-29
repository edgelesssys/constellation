terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.8.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}
resource "libvirt_domain" "instance_group" {
  count    = var.amount
  name     = "${var.base_name}-${var.role}-${local.group_uid}-${count.index}"
  memory   = var.memory
  vcpu     = var.vcpus
  machine  = var.machine
  firmware = local.firmware
  dynamic "cpu" {
    for_each = var.boot_mode == "direct-linux-boot" ? [1] : []
    content {
      mode = "host-passthrough"
    }
  }
  dynamic "nvram" {
    for_each = var.boot_mode == "uefi" ? [1] : []
    content {
      file     = "/var/lib/libvirt/qemu/nvram/${var.role}-${count.index}_VARS.fd"
      template = var.nvram
    }
  }
  xml {
    xslt = file("${path.module}/${local.xslt_filename}")
  }
  kernel  = local.kernel
  initrd  = local.initrd
  cmdline = local.cmdline
  tpm {
    backend_type    = "emulator"
    backend_version = "2.0"
  }
  disk {
    volume_id = element(libvirt_volume.boot_volume.*.id, count.index)
  }
  disk {
    volume_id = element(libvirt_volume.state_volume.*.id, count.index)
  }
  network_interface {
    network_id     = var.network_id
    hostname       = "${var.role}-${count.index}"
    addresses      = [cidrhost(var.cidr, local.ip_range_start + count.index)]
    wait_for_lease = true
  }
  console {
    type        = "pty"
    target_port = "0"
  }
}

resource "libvirt_volume" "boot_volume" {
  count          = var.amount
  name           = "constellation-${var.role}-${local.group_uid}-${count.index}-boot"
  pool           = var.pool
  base_volume_id = var.boot_volume_id
  lifecycle {
    ignore_changes = [
      name, # required. Allow legacy scale sets to keep their old names
    ]
  }
}

resource "libvirt_volume" "state_volume" {
  count  = var.amount
  name   = "constellation-${var.role}-${local.group_uid}-${count.index}-state"
  pool   = var.pool
  size   = local.state_disk_size_byte
  format = "qcow2"
  lifecycle {
    ignore_changes = [
      name, # required. Allow legacy scale sets to keep their old names
    ]
  }
}

resource "random_id" "uid" {
  byte_length = 4
}

locals {
  group_uid            = random_id.uid.hex
  state_disk_size_byte = 1073741824 * var.state_disk_size
  ip_range_start       = 100
  kernel               = var.boot_mode == "direct-linux-boot" ? var.kernel_volume_id : null
  initrd               = var.boot_mode == "direct-linux-boot" ? var.initrd_volume_id : null
  cmdline              = var.boot_mode == "direct-linux-boot" ? [{ "_" = var.kernel_cmdline }] : null
  firmware             = var.boot_mode == "uefi" ? var.firmware : null
  xslt_filename        = var.boot_mode == "direct-linux-boot" ? "tdx_domain.xsl" : "domain.xsl"
}
