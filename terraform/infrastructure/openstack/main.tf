terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "1.52.1"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.6.0"
    }
  }
}

provider "openstack" {
  cloud = var.cloud
}

data "openstack_identity_auth_scope_v3" "scope" {
  name = "scope"
}

locals {
  uid                    = random_id.uid.hex
  name                   = "${var.name}-${local.uid}"
  initSecretHash         = random_password.initSecret.bcrypt_hash
  ports_node_range_start = "30000"
  ports_node_range_end   = "32767"
  ports_kubernetes       = "6443"
  ports_bootstrapper     = "9000"
  ports_konnectivity     = "8132"
  ports_verify           = "30081"
  ports_recovery         = "9999"
  ports_debugd           = "4000"
  cidr_vpc_subnet_nodes  = "192.168.178.0/24"
  tags                   = ["constellation-uid-${local.uid}"]
  identity_service = [
    for entry in data.openstack_identity_auth_scope_v3.scope.service_catalog :
    entry if entry.type == "identity"
  ][0]
  identity_endpoint = [
    for endpoint in local.identity_service.endpoints :
    endpoint if(endpoint.interface == "public")
  ][0]
  identity_internal_url = local.identity_endpoint.url
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "initSecret" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "openstack_images_image_v2" "constellation_os_image" {
  name             = local.name
  image_source_url = var.image_url
  web_download     = var.direct_download
  container_format = "bare"
  disk_format      = "raw"
  visibility       = "private"
  properties = {
    hw_firmware_type = "uefi"
    os_type          = "linux"
  }
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

resource "openstack_networking_router_v2" "vpc_router" {
  name                = local.name
  external_network_id = data.openstack_networking_network_v2.floating_ip_pool.network_id
}

resource "openstack_networking_router_interface_v2" "vpc_router_interface" {
  router_id = openstack_networking_router_v2.vpc_router.id
  subnet_id = openstack_networking_subnet_v2.vpc_subnetwork.id
}

resource "openstack_compute_secgroup_v2" "vpc_secgroup" {
  name        = local.name
  description = "Constellation VPC security group"

  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    self        = true
  }

  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "udp"
    cidr        = local.cidr_vpc_subnet_nodes
  }

  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "tcp"
    cidr        = local.cidr_vpc_subnet_nodes
  }

  rule {
    from_port   = local.ports_node_range_start
    to_port     = local.ports_node_range_end
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = local.ports_node_range_start
    to_port     = local.ports_node_range_end
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }

  dynamic "rule" {
    for_each = flatten([
      local.ports_kubernetes,
      local.ports_bootstrapper,
      local.ports_konnectivity,
      local.ports_verify,
      local.ports_recovery,
      var.debug ? [local.ports_debugd] : [],
    ])
    content {
      from_port   = rule.value
      to_port     = rule.value
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    }
  }
}

module "instance_group" {
  source                     = "./modules/instance_group"
  for_each                   = var.node_groups
  base_name                  = local.name
  node_group_name            = each.key
  role                       = each.value.role
  initial_count              = each.value.initial_count
  disk_size                  = each.value.state_disk_size
  state_disk_type            = each.value.state_disk_type
  availability_zone          = each.value.zone
  image_id                   = openstack_images_image_v2.constellation_os_image.image_id
  flavor_id                  = each.value.flavor_id
  security_groups            = [openstack_compute_secgroup_v2.vpc_secgroup.id]
  tags                       = local.tags
  uid                        = local.uid
  network_id                 = openstack_networking_network_v2.vpc_network.id
  init_secret_hash           = local.initSecretHash
  identity_internal_url      = local.identity_internal_url
  openstack_username         = var.openstack_username
  openstack_password         = var.openstack_password
  openstack_user_domain_name = var.openstack_user_domain_name
}

resource "openstack_networking_floatingip_v2" "public_ip" {
  pool        = data.openstack_networking_network_v2.floating_ip_pool.name
  description = "Public ip for first control plane node"
  tags        = local.tags
}


resource "openstack_compute_floatingip_associate_v2" "public_ip_associate" {
  floating_ip = openstack_networking_floatingip_v2.public_ip.address
  instance_id = module.instance_group["control_plane_default"].instance_ids.0
  depends_on = [
    openstack_networking_router_v2.vpc_router,
    openstack_networking_router_interface_v2.vpc_router_interface,
  ]
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
