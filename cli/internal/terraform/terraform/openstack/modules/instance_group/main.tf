terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "1.52.1"
    }
  }
}

locals {
  tags      = distinct(sort(concat(var.tags, ["constellation-role-${var.role}"], ["constellation-node-group-${var.node_group_name}"])))
  group_uid = random_id.uid.hex
  name      = "${var.base_name}-${var.role}-${local.group_uid}"
}

resource "random_id" "uid" {
  byte_length = 4
}

# TODO(malt3): get this API enabled in the test environment
# resource "openstack_compute_servergroup_v2" "instance_group" {
#   name = local.name
#   policies = ["soft-anti-affinity"]
# }

resource "openstack_compute_instance_v2" "instance_group_member" {
  name            = "${local.name}-${count.index}"
  count           = var.initial_count
  image_id        = var.image_id
  flavor_id       = var.flavor_id
  security_groups = var.security_groups
  tags            = local.tags
  # TODO(malt3): get this API enabled in the test environment
  # scheduler_hints {
  #   group = openstack_compute_servergroup_v2.instance_group.id
  # }
  network {
    uuid = var.network_id
  }
  block_device {
    uuid                  = var.image_id
    source_type           = "image"
    destination_type      = "local"
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
    openstack-auth-url             = var.identity_internal_url
    openstack-username             = var.openstack_username
    openstack-password             = var.openstack_password
    openstack-user-domain-name     = var.openstack_user_domain_name
  }
  availability_zone_hints = var.availability_zone
}
