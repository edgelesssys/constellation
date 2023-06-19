terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.69.1"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
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
  name            = "${var.base_name}${local.maybe_one}-${local.role_dashed}${local.maybe_uid}"
  state_disk_name = "state-disk"
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "google_compute_instance_template" "template" {
  name         = local.name
  machine_type = var.instance_type
  tags         = ["constellation-${var.uid}"] // Note that this is also applied as a label
  labels = merge(var.labels, {
    constellation-role       = local.role_dashed,
    constellation-node-group = var.node_group_name,
  })

  confidential_instance_config {
    enable_confidential_compute = true
  }

  disk {
    disk_size_gb = 10
    source_image = var.image_id
    auto_delete  = true
    boot         = true
    mode         = "READ_WRITE"
  }

  disk {
    disk_size_gb = var.disk_size
    disk_type    = var.disk_type
    auto_delete  = true
    device_name  = local.state_disk_name // This name is used by disk mapper to find the disk
    boot         = false
    mode         = "READ_WRITE"
    type         = "PERSISTENT"
  }

  metadata = {
    kube-env                       = var.kube_env
    constellation-init-secret-hash = var.init_secret_hash
    serial-port-enable             = var.debug ? "TRUE" : "FALSE"
  }

  network_interface {
    network    = var.network
    subnetwork = var.subnetwork
    alias_ip_range {
      ip_cidr_range         = "/24"
      subnetwork_range_name = var.alias_ip_range_name
    }
  }

  scheduling {
    on_host_maintenance = "TERMINATE"
  }

  service_account {
    scopes = [
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/trace.append",
      "https://www.googleapis.com/auth/cloud-platform",
    ]
  }

  shielded_instance_config {
    enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  lifecycle {
    ignore_changes = [
      tags,
      labels,
      disk, # required. update procedure modifies the instance template externally
      metadata,
      network_interface,
      scheduling,
      service_account,
      shielded_instance_config,
    ]
  }
}

resource "google_compute_instance_group_manager" "instance_group_manager" {
  provider           = google-beta
  name               = local.name
  description        = "Instance group manager for Constellation"
  base_instance_name = local.name
  zone               = var.zone
  target_size        = var.instance_count

  dynamic "stateful_disk" {
    for_each = var.role == "ControlPlane" ? [1] : []
    content {
      device_name = local.state_disk_name
      delete_rule = "ON_PERMANENT_INSTANCE_DELETION"
    }
  }

  dynamic "stateful_internal_ip" {
    for_each = var.role == "ControlPlane" ? [1] : []
    content {
      interface_name = "nic0"
      delete_rule    = "ON_PERMANENT_INSTANCE_DELETION"
    }
  }

  version {
    instance_template = google_compute_instance_template.template.id
  }

  dynamic "named_port" {
    for_each = toset(var.named_ports)
    content {
      name = named_port.value.name
      port = named_port.value.port
    }
  }

  lifecycle {
    ignore_changes = [
      target_size, # required. autoscaling modifies the instance count externally
      version,     # required. update procedure modifies the instance template externally
    ]
  }
}
