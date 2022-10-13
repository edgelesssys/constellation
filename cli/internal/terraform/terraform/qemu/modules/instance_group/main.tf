terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.14"
    }
  }
}

locals {
  state_disk_size_byte = 1073741824 * var.state_disk_size
  ip_range_start       = 100
}

resource "libvirt_domain" "instance_group" {
  name    = "${var.name}-${var.role}-${count.index}"
  count   = var.amount
  memory  = var.memory
  vcpu    = var.vcpus
  machine = var.machine
  tpm {
    backend_type    = "emulator"
    backend_version = "2.0"
  }
  disk = [
    {
      volume_id = element(libvirt_volume.boot_volume.*.id, count.index)
      scsi : true,
      // fix for https://github.com/dmacvicar/terraform-provider-libvirt/issues/728
      block_device : null,
      file : null,
      url : null,
      wwn : null
    },
    {
      volume_id = element(libvirt_volume.state_volume.*.id, count.index)
      // fix for https://github.com/dmacvicar/terraform-provider-libvirt/issues/728
      block_device : null,
      file : null,
      scsi : null,
      url : null,
      wwn : null
    },
  ]
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
  xml {
    xslt = file("${path.module}/domain.xsl")
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
