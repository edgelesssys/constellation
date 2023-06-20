terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

resource "random_id" "uid" {
  byte_length = 4
}


locals {
  # migration: allow the old node group names to work since they were created without the uid
  # and without multiple node groups in mind
  # node_group: worker_default => name == "<base>-1-worker"
  # node_group: control_plane_default => name:  "<base>-control-plane"
  # new names:
  # node_group: foo, role: Worker => name == "<base>-worker-<uid>"
  # node_group: bar, role: ControlPlane => name == "<base>-control-plane-<uid>"
  role_dashed     = var.role == "ControlPlane" ? "control-plane" : "worker"
  group_uid       = random_id.uid.hex
  maybe_uid       = (var.node_group_name == "control_plane_default" || var.node_group_name == "worker_default") ? "" : "-${local.group_uid}"
  maybe_one       = var.node_group_name == "worker_default" ? "-1" : ""
}
