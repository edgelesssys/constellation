terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "1.48.0"
    }
  }
}

locals {
  role_dashed = var.role == "ControlPlane" ? "control-plane" : "worker"
  name        = "${var.name}-${local.role_dashed}"
  tags        = distinct(sort(concat(var.tags, ["constellation-role-${local.role_dashed}"])))
}

# TODO: get this API enabled in the test environment
# resource "openstack_compute_servergroup_v2" "instance_group" {
#   name = local.name
#   policies = ["soft-anti-affinity"]
# }

resource "openstack_compute_instance_v2" "instance_group_member" {
  name            = "${local.name}-${count.index}"
  count           = var.instance_count
  image_id        = var.image_id
  flavor_id       = var.flavor_id
  security_groups = var.security_groups
  tags            = local.tags
  # TODO: get this API enabled in the test environment
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
    boot_index            = 1
    delete_on_termination = true
  }
  metadata = {
    constellation-role             = local.role_dashed
    constellation-uid              = var.uid
    constellation-init-secret-hash = var.init_secret_hash
    openstack-auth-url             = var.identity_internal_url
  }
  user_data               = var.openstack_service_account_token
  availability_zone_hints = var.availability_zone
}
