terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "3.0.0"
    }

    stackit = {
      source  = "stackitcloud/stackit"
      version = "0.51.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "openstack" {
  cloud = var.cloud
}

provider "stackit" {
  default_region = "eu01"
}


data "openstack_identity_auth_scope_v3" "scope" {
  name = "scope"
}

locals {
  uid                    = random_id.uid.hex
  name                   = "${var.name}-${local.uid}"
  init_secret_hash       = random_password.init_secret.bcrypt_hash
  ports_node_range_start = "30000"
  ports_node_range_end   = "32767"
  control_plane_named_ports = flatten([
    { name = "kubernetes", port = "6443", health_check = "HTTPS" },
    { name = "bootstrapper", port = "9000", health_check = "TCP" },
    { name = "verify", port = "30081", health_check = "TCP" },
    { name = "recovery", port = "9999", health_check = "TCP" },
    { name = "join", port = "30090", health_check = "TCP" },
    var.debug ? [{ name = "debugd", port = "4000", health_check = "TCP" }] : [],
    var.emergency_ssh ? [{ name = "ssh", port = "22", health_check = "TCP" }] : [],
  ])
  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  cidr_vpc_subnet_lbs   = "192.168.177.0/24"
  tags                  = concat(["constellation-uid-${local.uid}"], var.additional_tags)
  identity_service = [
    for entry in data.openstack_identity_auth_scope_v3.scope.service_catalog :
    entry if entry.type == "identity"
  ][0]
  identity_endpoint = [
    for endpoint in local.identity_service.endpoints :
    endpoint if(endpoint.interface == "public")
  ][0]
  identity_internal_url = local.identity_endpoint.url
  cloudsyaml_path       = length(var.openstack_clouds_yaml_path) > 0 ? var.openstack_clouds_yaml_path : "~/.config/openstack/clouds.yaml"
  cloudsyaml            = yamldecode(file(pathexpand(local.cloudsyaml_path)))
  cloudyaml             = local.cloudsyaml.clouds[var.cloud]
  revision              = 1
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

data "openstack_networking_network_v2" "floating_ip_pool" {
  network_id = var.floating_ip_pool_id
}

resource "openstack_networking_network_v2" "vpc_network" {
  name        = local.name
  description = "Constellation VPC network"
  tags        = local.tags
}

resource "openstack_networking_subnet_v2" "vpc_subnetwork" {
  name        = local.name
  description = "Constellation VPC subnetwork"
  network_id  = openstack_networking_network_v2.vpc_network.id
  cidr        = local.cidr_vpc_subnet_nodes
  dns_nameservers = [
    "1.1.1.1",
    "8.8.8.8",
    "9.9.9.9",
  ]
  tags = local.tags
}

resource "openstack_networking_subnet_v2" "lb_subnetwork" {
  name        = "${var.name}-${local.uid}-lb"
  description = "Constellation LB subnetwork"
  network_id  = openstack_networking_network_v2.vpc_network.id
  cidr        = local.cidr_vpc_subnet_lbs
  dns_nameservers = [
    "1.1.1.1",
    "8.8.8.8",
    "9.9.9.9",
  ]
  tags = local.tags
}

resource "openstack_networking_router_v2" "vpc_router" {
  name                = local.name
  external_network_id = data.openstack_networking_network_v2.floating_ip_pool.network_id
}

resource "openstack_networking_router_interface_v2" "vpc_router_interface" {
  router_id = openstack_networking_router_v2.vpc_router.id
  subnet_id = openstack_networking_subnet_v2.vpc_subnetwork.id
}

resource "openstack_networking_router_interface_v2" "lbs_router_interface_lbs" {
  router_id = openstack_networking_router_v2.vpc_router.id
  subnet_id = openstack_networking_subnet_v2.lb_subnetwork.id
}

resource "openstack_networking_secgroup_v2" "vpc_secgroup" {
  name        = local.name
  description = "Constellation VPC security group"
}

resource "openstack_networking_secgroup_rule_v2" "icmp_in" {
  description       = "icmp ingress"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  port_range_min    = 0
  port_range_max    = 0
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "icmp_out" {
  description       = "icmp egress"
  direction         = "egress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  port_range_min    = 0
  port_range_max    = 0
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "tcp_out" {
  description       = "tcp egress"
  direction         = "egress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 0
  port_range_max    = 0
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "udp_out" {
  description       = "udp egress"
  direction         = "egress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 0
  port_range_max    = 0
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "tcp_between_nodes" {
  description       = "tcp between nodes"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 0
  port_range_max    = 0
  remote_ip_prefix  = local.cidr_vpc_subnet_nodes
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "udp_between_nodes" {
  description       = "udp between nodes"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 0
  port_range_max    = 0
  remote_ip_prefix  = local.cidr_vpc_subnet_nodes
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "nodeport_tcp" {
  description       = "nodeport tcp"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = local.ports_node_range_start
  port_range_max    = local.ports_node_range_end
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "nodeport_udp" {
  description       = "nodeport udp"
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = local.ports_node_range_start
  port_range_max    = local.ports_node_range_end
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}

resource "openstack_networking_secgroup_rule_v2" "tcp_ingress" {
  for_each          = { for item in local.control_plane_named_ports : item.name => item }
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = each.value.port
  port_range_max    = each.value.port
  security_group_id = openstack_networking_secgroup_v2.vpc_secgroup.id
}


module "instance_group" {
  source                           = "./modules/instance_group"
  for_each                         = var.node_groups
  base_name                        = local.name
  node_group_name                  = each.key
  role                             = each.value.role
  initial_count                    = each.value.initial_count
  disk_size                        = each.value.state_disk_size
  state_disk_type                  = each.value.state_disk_type
  availability_zone                = each.value.zone
  image_id                         = var.image_id
  flavor_id                        = each.value.flavor_id
  security_groups                  = [openstack_networking_secgroup_v2.vpc_secgroup.id]
  tags                             = local.tags
  uid                              = local.uid
  network_id                       = openstack_networking_network_v2.vpc_network.id
  subnet_id                        = openstack_networking_subnet_v2.vpc_subnetwork.id
  init_secret_hash                 = local.init_secret_hash
  identity_internal_url            = local.identity_internal_url
  openstack_username               = local.cloudyaml["auth"]["username"]
  openstack_password               = local.cloudyaml["auth"]["password"]
  openstack_user_domain_name       = local.cloudyaml["auth"]["user_domain_name"]
  openstack_region_name            = local.cloudyaml["region_name"]
  openstack_load_balancer_endpoint = openstack_networking_floatingip_v2.public_ip.address
}

resource "openstack_networking_floatingip_v2" "public_ip" {
  pool        = data.openstack_networking_network_v2.floating_ip_pool.name
  description = "Public ip for first control plane node"
  tags        = local.tags
}

resource "openstack_networking_floatingip_associate_v2" "public_ip_associate" {
  count       = var.cloud == "stackit" ? 0 : 1
  floating_ip = openstack_networking_floatingip_v2.public_ip.address
  port_id     = module.instance_group["control_plane_default"].port_ids.0
  depends_on = [
    openstack_networking_router_v2.vpc_router,
    openstack_networking_router_interface_v2.vpc_router_interface,
  ]
}

module "stackit_loadbalancer" {
  count              = var.cloud == "stackit" ? 1 : 0
  source             = "./modules/stackit_loadbalancer"
  name               = local.name
  stackit_project_id = var.stackit_project_id
  member_ips         = module.instance_group["control_plane_default"].ips
  network_id         = openstack_networking_network_v2.vpc_network.id
  external_address   = openstack_networking_floatingip_v2.public_ip.address
  ports = {
    for port in local.control_plane_named_ports : port.name => port.port
  }
}


moved {
  from = module.instance_group_control_plane
  to   = module.instance_group["control_plane_default"]
}

moved {
  from = module.instance_group_worker
  to   = module.instance_group["worker_default"]
}

# TODO(malt3): get LoadBalancer API enabled in the test environment
# resource "openstack_lb_loadbalancer_v2" "loadbalancer" {
#   name          = local.name
#   description   = "Constellation load balancer"
#   vip_subnet_id = openstack_networking_subnet_v2.vpc_subnetwork.id
# }


# resource "openstack_networking_floatingip_v2" "loadbalancer_ip" {
#   pool        = data.openstack_networking_network_v2.floating_ip_pool.name
#   description = "Loadbalancer ip for ${local.name}"
#   tags        = local.tags
# }

# module "loadbalancer_kube" {
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-kube"
#   member_ips      = module.instance_group_control_plane.ips.value
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_kubernetes
# }

# module "loadbalancer_boot" {
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-boot"
#   member_ips      = module.instance_group_control_plane.ips
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_bootstrapper
# }

# module "loadbalancer_verify" {
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-verify"
#   member_ips      = module.instance_group_control_plane.ips
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_verify
# }

# module "loadbalancer_konnectivity" {
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-konnectivity"
#   member_ips      = module.instance_group_control_plane.ips
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_konnectivity
# }

# module "loadbalancer_recovery" {
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-recovery"
#   member_ips      = module.instance_group_control_plane.ips
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_recovery
# }

# module "loadbalancer_debugd" {
#   count           = var.debug ? 1 : 0 // only deploy debugd in debug mode
#   source          = "./modules/loadbalancer"
#   name            = "${local.name}-debugd"
#   member_ips      = module.instance_group_control_plane.ips
#   loadbalancer_id = openstack_lb_loadbalancer_v2.loadbalancer.id
#   subnet_id       = openstack_networking_subnet_v2.vpc_subnetwork.id
#   port            = local.ports_debugd
# }
