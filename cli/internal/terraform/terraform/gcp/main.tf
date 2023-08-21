terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.83.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = "4.72.0"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

provider "google-beta" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

locals {
  uid            = random_id.uid.hex
  name           = "${var.name}-${local.uid}"
  initSecretHash = random_password.initSecret.bcrypt_hash
  labels = {
    constellation-uid = local.uid,
  }
  ports_node_range      = "30000-32767"
  ports_kubernetes      = "6443"
  ports_bootstrapper    = "9000"
  ports_konnectivity    = "8132"
  ports_verify          = "30081"
  ports_recovery        = "9999"
  ports_join            = "30090"
  ports_debugd          = "4000"
  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  cidr_vpc_subnet_pods  = "10.10.0.0/16"
  kube_env              = "AUTOSCALER_ENV_VARS: kube_reserved=cpu=1060m,memory=1019Mi,ephemeral-storage=41Gi;node_labels=;os=linux;os_distribution=cos;evictionHard="
  control_plane_named_ports = flatten([
    { name = "kubernetes", port = local.ports_kubernetes },
    { name = "bootstrapper", port = local.ports_bootstrapper },
    { name = "verify", port = local.ports_verify },
    { name = "konnectivity", port = local.ports_konnectivity },
    { name = "recovery", port = local.ports_recovery },
    { name = "join", port = local.ports_join },
    var.debug ? [{ name = "debugd", port = local.ports_debugd }] : [],
  ])
  node_groups_by_role = {
    for name, node_group in var.node_groups : node_group.role => name...
  }
  control_plane_instance_groups = [
    for control_plane in local.node_groups_by_role["control-plane"] : module.instance_group[control_plane].instance_group
  ]
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "initSecret" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "google_compute_network" "vpc_network" {
  name                    = local.name
  description             = "Constellation VPC network"
  auto_create_subnetworks = false
  mtu                     = 8896
}

resource "google_compute_subnetwork" "vpc_subnetwork" {
  name          = local.name
  description   = "Constellation VPC subnetwork"
  network       = google_compute_network.vpc_network.id
  ip_cidr_range = local.cidr_vpc_subnet_nodes
  secondary_ip_range = [
    {
      range_name    = local.name,
      ip_cidr_range = local.cidr_vpc_subnet_pods,
    }
  ]
}

resource "google_compute_router" "vpc_router" {
  name        = local.name
  description = "Constellation VPC router"
  network     = google_compute_network.vpc_network.id
}

resource "google_compute_router_nat" "vpc_router_nat" {
  name                               = local.name
  router                             = google_compute_router.vpc_router.name
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_firewall" "firewall_external" {
  name          = local.name
  description   = "Constellation VPC firewall"
  network       = google_compute_network.vpc_network.id
  source_ranges = ["0.0.0.0/0"]
  direction     = "INGRESS"

  allow {
    protocol = "tcp"
    ports = flatten([
      local.ports_node_range,
      local.ports_bootstrapper,
      local.ports_kubernetes,
      local.ports_konnectivity,
      local.ports_recovery,
      local.ports_join,
      var.debug ? [local.ports_debugd] : [],
    ])
  }

}

resource "google_compute_firewall" "firewall_internal_nodes" {
  name          = "${local.name}-nodes"
  description   = "Constellation VPC firewall"
  network       = google_compute_network.vpc_network.id
  source_ranges = [local.cidr_vpc_subnet_nodes]
  direction     = "INGRESS"

  allow { protocol = "tcp" }
  allow { protocol = "udp" }
  allow { protocol = "icmp" }
}

resource "google_compute_firewall" "firewall_internal_pods" {
  name          = "${local.name}-pods"
  description   = "Constellation VPC firewall"
  network       = google_compute_network.vpc_network.id
  source_ranges = [local.cidr_vpc_subnet_pods]
  direction     = "INGRESS"

  allow { protocol = "tcp" }
  allow { protocol = "udp" }
  allow { protocol = "icmp" }
}


module "instance_group" {
  source              = "./modules/instance_group"
  for_each            = var.node_groups
  base_name           = local.name
  node_group_name     = each.key
  role                = each.value.role
  zone                = each.value.zone
  uid                 = local.uid
  instance_type       = each.value.instance_type
  initial_count       = each.value.initial_count
  image_id            = var.image_id
  disk_size           = each.value.disk_size
  disk_type           = each.value.disk_type
  network             = google_compute_network.vpc_network.id
  subnetwork          = google_compute_subnetwork.vpc_subnetwork.id
  alias_ip_range_name = google_compute_subnetwork.vpc_subnetwork.secondary_ip_range[0].range_name
  kube_env            = local.kube_env
  debug               = var.debug
  named_ports         = each.value.role == "control-plane" ? local.control_plane_named_ports : []
  labels              = local.labels
  init_secret_hash    = local.initSecretHash
  custom_endpoint     = var.custom_endpoint
}

resource "google_compute_global_address" "loadbalancer_ip" {
  name = local.name
}

module "loadbalancer_kube" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "HTTPS"
  backend_port_name       = "kubernetes"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_kubernetes
  frontend_labels         = merge(local.labels, { constellation-use = "kubernetes" })
}

module "loadbalancer_boot" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "bootstrapper"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_bootstrapper
  frontend_labels         = merge(local.labels, { constellation-use = "bootstrapper" })
}

module "loadbalancer_verify" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "verify"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_verify
  frontend_labels         = merge(local.labels, { constellation-use = "verify" })
}

module "loadbalancer_konnectivity" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "konnectivity"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_konnectivity
  frontend_labels         = merge(local.labels, { constellation-use = "konnectivity" })
}

module "loadbalancer_recovery" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "recovery"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_recovery
  frontend_labels         = merge(local.labels, { constellation-use = "recovery" })
}

module "loadbalancer_join" {
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "join"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_join
  frontend_labels         = merge(local.labels, { constellation-use = "join" })
}

module "loadbalancer_debugd" {
  count                   = var.debug ? 1 : 0 // only deploy debugd in debug mode
  source                  = "./modules/loadbalancer"
  name                    = local.name
  health_check            = "TCP"
  backend_port_name       = "debugd"
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip.self_link
  port                    = local.ports_debugd
  frontend_labels         = merge(local.labels, { constellation-use = "debugd" })
}

moved {
  from = module.instance_group_control_plane
  to   = module.instance_group["control_plane_default"]
}

moved {
  from = module.instance_group_worker
  to   = module.instance_group["worker_default"]
}
