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

locals {
  group_uid       = random_id.uid.hex
  name            = "${var.base_name}-${var.role}-${local.group_uid}"
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
    constellation-role       = var.role,
    constellation-node-group = var.node_group_name,
  })

  confidential_instance_config {
    enable_confidential_compute = true
    confidential_instance_type  = var.cc_technology == "SEV_SNP" ? "SEV_SNP" : null
  }

  # If SEV-SNP is used, we have to explicitly select a Milan processor, as per
  # https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance_template#confidential_instance_type
  min_cpu_platform = var.cc_technology == "SEV_SNP" ? "AMD Milan" : null

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
    serial-port-enable             = "TRUE"
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

  # Define all IAM access via the service account and not via scopes:
  # See: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance_template#nested_service_account
  service_account {
    email  = var.iam_service_account_vm
    scopes = ["cloud-platform"]
  }

  shielded_instance_config {
    enable_secure_boot          = false
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  lifecycle {
    ignore_changes = [
      name, # required. legacy instance templates used different naming scheme
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
  name               = local.name
  description        = "Instance group manager for Constellation"
  base_instance_name = local.name
  zone               = var.zone
  target_size        = var.initial_count

  dynamic "stateful_disk" {
    for_each = var.role == "control-plane" ? [1] : []
    content {
      device_name = local.state_disk_name
      delete_rule = "ON_PERMANENT_INSTANCE_DELETION"
    }
  }

  dynamic "stateful_internal_ip" {
    for_each = var.role == "control-plane" ? [1] : []
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
      name,               # required. legacy instance templates used different naming scheme
      base_instance_name, # required. legacy instance templates used different naming scheme
      target_size,        # required. autoscaling modifies the instance count externally
      version,            # required. update procedure modifies the instance template externally
    ]
  }
}
