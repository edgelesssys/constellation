terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.57.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "4.53.1"
    }
  }
}

locals {
  role_dashed     = var.role == "ControlPlane" ? "control-plane" : "worker"
  name            = "${var.name}-${local.role_dashed}"
  state_disk_name = "state-disk"
}

resource "google_compute_instance_template" "template" {
  name         = local.name
  machine_type = var.instance_type
  tags         = ["constellation-${var.uid}"] // Note that this is also applied as a label 
  labels       = merge(var.labels, { constellation-role = local.role_dashed })

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
      subnetwork_range_name = var.name
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
}

resource "google_compute_instance_group_manager" "instance_group_manager" {
  provider           = google-beta
  name               = local.name
  description        = "Instance group manager for Constellation"
  base_instance_name = local.name
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
