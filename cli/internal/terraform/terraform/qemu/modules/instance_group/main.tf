terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.7.1"
    }
  }
}

resource "libvirt_domain" "instance_group" {
  name     = "${var.name}-${var.role}-${count.index}"
  count    = var.amount
  memory   = var.memory
  vcpu     = var.vcpus
  machine  = var.machine
  firmware = local.firmware
  dynamic "nvram" {
    for_each = var.boot_mode == "uefi" ? [1] : []
    content {
      file     = "/var/lib/libvirt/qemu/nvram/${var.role}-${count.index}_VARS.fd"
      template = var.nvram
    }
  }
  dynamic "xml" {
    for_each = var.boot_mode == "uefi" ? [1] : []
    content {
      xslt = file("${path.module}/domain.xsl")
    }
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
    scsi      = true
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
  name           = "constellation-${var.role}-${count.index}-boot"
  count          = var.amount
  pool           = var.pool
  base_volume_id = var.boot_volume_id
}

resource "libvirt_volume" "state_volume" {
  name   = "constellation-${var.role}-${count.index}-state"
  count  = var.amount
  pool   = var.pool
  size   = local.state_disk_size_byte
  format = "qcow2"
}

locals {
  state_disk_size_byte = 1073741824 * var.state_disk_size
  ip_range_start       = 100
  kernel               = var.boot_mode == "direct-linux-boot" ? var.kernel_volume_id : null
  initrd               = var.boot_mode == "direct-linux-boot" ? var.initrd_volume_id : null
  cmdline              = var.boot_mode == "direct-linux-boot" ? [{ "_" = var.kernel_cmdline }] : null
  firmware             = var.boot_mode == "uefi" ? var.firmware : null
}
