terraform {
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.8.3"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "libvirt" {
  uri = var.libvirt_uri
}

provider "docker" {
  host = "unix:///var/run/docker.sock"
}

locals {
  cidr_vpc_subnet_nodes          = "10.42.0.0/22"
  cidr_vpc_subnet_control_planes = "10.42.1.0/24"
  cidr_vpc_subnet_worker         = "10.42.2.0/24"
  revision                       = 1
}

# A way to force replacement of resources if the provider does not want to replace them
# see: https://developer.hashicorp.com/terraform/language/resources/terraform-data#example-usage-data-for-replace_triggered_by
resource "terraform_data" "replacement" {
  input = local.revision
}

resource "random_password" "init_secret" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "docker_image" "qemu_metadata" {
  name         = var.metadata_api_image
  keep_locally = true
}

resource "docker_container" "qemu_metadata" {
  name         = "${var.name}-qemu-metadata"
  image        = docker_image.qemu_metadata.image_id
  network_mode = "host"
  rm           = true
  command = [
    "--network",
    "${var.name}-network",
    "--libvirt-uri",
    "${var.metadata_libvirt_uri}",
    "--initsecrethash",
    "${random_password.init_secret.bcrypt_hash}",
  ]
  mounts {
    source = abspath(var.libvirt_socket_path)
    target = "/var/run/libvirt/libvirt-sock"
    type   = "bind"
  }
}

module "node_group" {
  source           = "./modules/instance_group"
  base_name        = var.name
  for_each         = var.node_groups
  node_group_name  = each.key
  role             = each.value.role
  amount           = each.value.initial_count
  state_disk_size  = each.value.disk_size
  vcpus            = each.value.vcpus
  memory           = each.value.memory
  machine          = var.machine
  cidr             = each.value.role == "control-plane" ? local.cidr_vpc_subnet_control_planes : local.cidr_vpc_subnet_worker
  network_id       = libvirt_network.constellation.id
  pool             = libvirt_pool.cluster.name
  boot_mode        = var.constellation_boot_mode
  boot_volume_id   = libvirt_volume.image_id.id
  kernel_volume_id = local.kernel_volume_id
  initrd_volume_id = local.initrd_volume_id
  kernel_cmdline   = each.value.role == "control-plane" ? local.kernel_cmdline : var.constellation_cmdline
  firmware         = var.firmware
  nvram            = var.nvram
}

resource "libvirt_pool" "cluster" {
  name = "${var.name}-storage-pool"
  type = "dir"
  path = "/var/lib/libvirt/images"
}

resource "libvirt_volume" "image_id" {
  name   = "${var.name}-node-image"
  pool   = libvirt_pool.cluster.name
  source = var.image_id
  format = var.image_format
}

resource "libvirt_volume" "constellation_kernel" {
  name   = "${var.name}-kernel"
  pool   = libvirt_pool.cluster.name
  source = var.constellation_kernel
  format = "raw"
  count  = var.constellation_boot_mode == "direct-linux-boot" ? 1 : 0
}

resource "libvirt_volume" "constellation_initrd" {
  name   = "${var.name}-initrd"
  pool   = libvirt_pool.cluster.name
  source = var.constellation_initrd
  format = "raw"
  count  = var.constellation_boot_mode == "direct-linux-boot" ? 1 : 0
}

resource "libvirt_network" "constellation" {
  name      = "${var.name}-network"
  mode      = "nat"
  addresses = ["10.42.0.0/16"]
  dhcp {
    enabled = true
  }
  dns {
    enabled = true
  }
}

locals {
  kernel_volume_id = var.constellation_boot_mode == "direct-linux-boot" ? libvirt_volume.constellation_kernel[0].id : null
  initrd_volume_id = var.constellation_boot_mode == "direct-linux-boot" ? libvirt_volume.constellation_initrd[0].id : null
  kernel_cmdline   = var.constellation_boot_mode == "direct-linux-boot" ? var.constellation_cmdline : null
}
