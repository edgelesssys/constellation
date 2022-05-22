terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.14"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///session"
}

module "control_plane" {
  source          = "./modules/instance_group"
  role            = "control-plane"
  amount          = var.control_plane_count
  vcpus           = var.vcpus
  memory          = var.memory
  state_disk_size = var.state_disk_size
  ip_range_start  = var.ip_range_start
  cidr            = "10.42.1.0/24"
  network_id      = libvirt_network.constellation.id
  pool            = libvirt_pool.cluster.name
  boot_volume_id  = libvirt_volume.constellation_coreos_image.id
}

module "worker" {
  source          = "./modules/instance_group"
  role            = "worker"
  amount          = var.worker_count
  vcpus           = var.vcpus
  memory          = var.memory
  state_disk_size = var.state_disk_size
  ip_range_start  = var.ip_range_start
  cidr            = "10.42.2.0/24"
  network_id      = libvirt_network.constellation.id
  pool            = libvirt_pool.cluster.name
  boot_volume_id  = libvirt_volume.constellation_coreos_image.id
}

resource "libvirt_pool" "cluster" {
  name = "constellation"
  type = "dir"
  path = "/var/lib/libvirt/images"
}

resource "libvirt_volume" "constellation_coreos_image" {
  name   = "constellation-coreos-image"
  pool   = libvirt_pool.cluster.name
  source = var.constellation_coreos_image_qcow2
  format = "qcow2"
}

resource "libvirt_network" "constellation" {
  name      = "constellation"
  mode      = "nat"
  addresses = ["10.42.0.0/16"]
  dhcp {
    enabled = true
  }
  dns {
    enabled = true
  }
}
