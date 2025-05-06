terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.33.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

locals {
  uid              = random_id.uid.hex
  name             = "${var.name}-${local.uid}"
  init_secret_hash = random_password.init_secret.bcrypt_hash
  labels = merge(
    var.additional_labels,
    { constellation-uid = local.uid }
  )
  ports_node_range      = "30000-32767"
  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  cidr_vpc_subnet_pods  = "10.10.0.0/16"
  cidr_vpc_subnet_proxy = "192.168.179.0/24"
  cidr_vpc_subnet_ilb   = "192.168.180.0/24"
  kube_env              = "AUTOSCALER_ENV_VARS: kube_reserved=cpu=1060m,memory=1019Mi,ephemeral-storage=41Gi;node_labels=;os=linux;os_distribution=cos;evictionHard="
  control_plane_named_ports = flatten([
    { name = "kubernetes", port = "6443", health_check = "HTTPS" },
    { name = "bootstrapper", port = "9000", health_check = "TCP" },
    { name = "verify", port = "30081", health_check = "TCP" },
    { name = "konnectivity", port = "8132", health_check = "TCP" },
    { name = "recovery", port = "9999", health_check = "TCP" },
    { name = "join", port = "30090", health_check = "TCP" },
    var.debug ? [{ name = "debugd", port = "4000", health_check = "TCP" }] : [],
    var.emergency_ssh ? [{ name = "ssh", port = "22", health_check = "TCP" }] : [],
  ])
  node_groups_by_role = {
    for name, node_group in var.node_groups : node_group.role => name...
  }
  control_plane_instance_groups = [
    for control_plane in local.node_groups_by_role["control-plane"] : module.instance_group[control_plane].instance_group_url
  ]
  in_cluster_endpoint     = var.internal_load_balancer ? google_compute_address.loadbalancer_ip_internal[0].address : google_compute_global_address.loadbalancer_ip[0].address
  out_of_cluster_endpoint = var.debug && var.internal_load_balancer ? module.jump_host[0].ip : local.in_cluster_endpoint
  revision                = 1
}

# A way to force replacement of resources if the provider does not want to replace them
# see: https://developer.hashicorp.com/terraform/language/resources/terraform-data#example-usage-data-for-replace_triggered_by
resource "terraform_data" "replacement" {
  input = local.revision
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "init_secret" {
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
  secondary_ip_range {
    range_name    = local.name
    ip_cidr_range = local.cidr_vpc_subnet_pods
  }
}

resource "google_compute_subnetwork" "proxy_subnet" {
  count         = var.internal_load_balancer ? 1 : 0
  name          = "${local.name}-proxy"
  ip_cidr_range = local.cidr_vpc_subnet_proxy
  region        = var.region
  purpose       = "REGIONAL_MANAGED_PROXY"
  role          = "ACTIVE"
  network       = google_compute_network.vpc_network.id
}

resource "google_compute_subnetwork" "ilb_subnet" {
  count         = var.internal_load_balancer ? 1 : 0
  name          = "${local.name}-ilb"
  ip_cidr_range = local.cidr_vpc_subnet_ilb
  region        = var.region
  network       = google_compute_network.vpc_network.id
  depends_on    = [google_compute_subnetwork.proxy_subnet]
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
      [for port in local.control_plane_named_ports : port.port],
      [local.ports_node_range],
      var.internal_load_balancer ? [22] : [],
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
  source                 = "./modules/instance_group"
  for_each               = var.node_groups
  base_name              = local.name
  node_group_name        = each.key
  role                   = each.value.role
  zone                   = each.value.zone
  uid                    = local.uid
  instance_type          = each.value.instance_type
  initial_count          = each.value.initial_count
  image_id               = var.image_id
  disk_size              = each.value.disk_size
  disk_type              = each.value.disk_type
  network                = google_compute_network.vpc_network.id
  subnetwork             = google_compute_subnetwork.vpc_subnetwork.id
  alias_ip_range_name    = google_compute_subnetwork.vpc_subnetwork.secondary_ip_range[0].range_name
  kube_env               = local.kube_env
  debug                  = var.debug
  named_ports            = each.value.role == "control-plane" ? local.control_plane_named_ports : []
  labels                 = local.labels
  init_secret_hash       = local.init_secret_hash
  custom_endpoint        = var.custom_endpoint
  cc_technology          = var.cc_technology
  iam_service_account_vm = var.iam_service_account_vm
}

resource "google_compute_address" "loadbalancer_ip_internal" {
  count        = var.internal_load_balancer ? 1 : 0
  name         = local.name
  region       = var.region
  subnetwork   = google_compute_subnetwork.ilb_subnet[0].id
  purpose      = "SHARED_LOADBALANCER_VIP"
  address_type = "INTERNAL"
  labels       = local.labels
}

resource "google_compute_global_address" "loadbalancer_ip" {
  count = var.internal_load_balancer ? 0 : 1
  name  = local.name
}

module "loadbalancer_public" {
  // for every port in control_plane_named_ports if internal lb is disabled
  for_each                = var.internal_load_balancer ? {} : { for port in local.control_plane_named_ports : port.name => port }
  source                  = "./modules/loadbalancer"
  name                    = local.name
  backend_port_name       = each.value.name
  port                    = each.value.port
  health_check            = each.value.health_check
  backend_instance_groups = local.control_plane_instance_groups
  ip_address              = google_compute_global_address.loadbalancer_ip[0].self_link
  frontend_labels         = merge(local.labels, { constellation-use = each.value.name })
}

module "loadbalancer_internal" {
  for_each               = var.internal_load_balancer ? { for port in local.control_plane_named_ports : port.name => port } : {}
  source                 = "./modules/internal_load_balancer"
  name                   = local.name
  backend_port_name      = each.value.name
  port                   = each.value.port
  health_check           = each.value.health_check
  backend_instance_group = local.control_plane_instance_groups[0]
  ip_address             = google_compute_address.loadbalancer_ip_internal[0].self_link
  frontend_labels        = merge(local.labels, { constellation-use = each.value.name })

  region         = var.region
  network        = google_compute_network.vpc_network.id
  backend_subnet = google_compute_subnetwork.ilb_subnet[0].id
}

module "jump_host" {
  count          = var.internal_load_balancer && var.debug ? 1 : 0
  source         = "./modules/jump_host"
  base_name      = local.name
  zone           = var.zone
  subnetwork     = google_compute_subnetwork.vpc_subnetwork.id
  labels         = var.additional_labels
  lb_internal_ip = google_compute_address.loadbalancer_ip_internal[0].address
  ports          = [for port in local.control_plane_named_ports : port.port]
}
moved {
  from = module.loadbalancer_boot
  to   = module.loadbalancer_public["bootstrapper"]
}

moved {
  from = module.loadbalancer_kube
  to   = module.loadbalancer_public["kubernetes"]
}

moved {
  from = module.loadbalancer_verify
  to   = module.loadbalancer_public["verify"]
}

moved {
  from = module.loadbalancer_konnectivity
  to   = module.loadbalancer_public["konnectivity"]
}

moved {
  from = module.loadbalancer_recovery
  to   = module.loadbalancer_public["recovery"]
}

moved {
  from = module.loadbalancer_debugd[0]
  to   = module.loadbalancer_public["debugd"]
}
