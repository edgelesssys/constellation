terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.14"
    }
    docker = {
      source = "kreuzwerker/docker"
      version = "2.17.0"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///session"
}

provider "docker" {
  host = "unix:///var/run/docker.sock"

  registry_auth {
    address     = "ghcr.io"
    config_file = pathexpand("~/.docker/config.json")
  }
}

resource "docker_image" "qemu-metadata" {
  name = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:latest"
  keep_locally = true 
}

resource "docker_container" "qemu-metadata" {
  name = "qemu-metadata"
  image = docker_image.qemu-metadata.latest
  network_mode = "host"
  rm = true
  mounts {
    source = "/var/run/libvirt/libvirt-sock"
    target = "/var/run/libvirt/libvirt-sock"
    type = "bind" 
  }
  mounts {
    source = var.metadata_api_log_dir
    target = "/pcrs"
    type = "bind"
  }
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
  machine         = var.machine
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
  machine         = var.machine
}

resource "libvirt_pool" "cluster" {
  name = "constellation"
  type = "dir"
  path = "/var/lib/libvirt/images"
}

resource "libvirt_volume" "constellation_coreos_image" {
  name   = "constellation-coreos-image"
  pool   = libvirt_pool.cluster.name
  source = var.constellation_coreos_image
  format = var.image_format
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
