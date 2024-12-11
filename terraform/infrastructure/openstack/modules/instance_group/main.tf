terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "3.0.0"
    }
  }
}

locals {
  tags              = distinct(sort(concat(var.tags, ["constellation-role-${var.role}"], ["constellation-node-group-${var.node_group_name}"])))
  group_uid         = random_id.uid.hex
  name              = "${var.base_name}-${var.role}-${local.group_uid}"
  flavor_id_is_uuid = length(var.flavor_id) == 36 && length(regexall("^[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}$", var.flavor_id)) == 1
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "openstack_networking_port_v2" "port" {
  name           = "${local.name}-${count.index}"
  count          = var.initial_count
  admin_state_up = "true"

  network_id = var.network_id
  fixed_ip {
    subnet_id = var.subnet_id
  }

  security_group_ids = var.security_groups
}

# TODO(malt3): get this API enabled in the test environment
# resource "openstack_compute_servergroup_v2" "instance_group" {
#   name = local.name
#   policies = ["soft-anti-affinity"]
# }

data "openstack_compute_flavor_v2" "flavor" {
  flavor_id = local.flavor_id_is_uuid ? var.flavor_id : null
  name      = local.flavor_id_is_uuid ? null : var.flavor_id
}

resource "openstack_compute_instance_v2" "instance_group_member" {
  name      = "${local.name}-${count.index}"
  count     = var.initial_count
  flavor_id = data.openstack_compute_flavor_v2.flavor.id
  tags      = local.tags
  # TODO(malt3): get this API enabled in the test environment
  # scheduler_hints {
  #   group = openstack_compute_servergroup_v2.instance_group.id
  # }
  network {
    port = openstack_networking_port_v2.port[count.index].id
  }
  block_device {
    uuid                  = var.image_id
    source_type           = "image"
    destination_type      = "volume"
    volume_size           = "5"
    boot_index            = 0
    delete_on_termination = true
  }
  block_device {
    source_type           = "blank"
    destination_type      = "volume"
    volume_size           = var.disk_size
    volume_type           = var.state_disk_type
    boot_index            = 1
    delete_on_termination = true
  }
  metadata = {
    constellation-role             = var.role
    constellation-uid              = var.uid
    constellation-init-secret-hash = var.init_secret_hash
  }
  user_data = jsonencode({
    openstack-auth-url               = var.identity_internal_url
    openstack-username               = var.openstack_username
    openstack-password               = var.openstack_password
    openstack-user-domain-name       = var.openstack_user_domain_name
    openstack-region-name            = var.openstack_region_name
    openstack-load-balancer-endpoint = var.openstack_load_balancer_endpoint
  })
  availability_zone_hints = length(var.availability_zone) > 0 ? var.availability_zone : null
  lifecycle {
    ignore_changes = [block_device] # block device contains current image, which can be updated from inside the cluster
  }
}
