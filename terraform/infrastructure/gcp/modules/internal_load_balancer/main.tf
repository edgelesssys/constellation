terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.33.0"
    }
  }
}

locals {
  name = "${var.name}-${var.backend_port_name}"
}

resource "google_compute_region_health_check" "health" {
  name               = local.name
  region             = var.region
  check_interval_sec = 1
  timeout_sec        = 1

  dynamic "tcp_health_check" {
    for_each = var.health_check == "TCP" ? [1] : []
    content {
      port = var.port
    }
  }

  dynamic "https_health_check" {
    for_each = var.health_check == "HTTPS" ? [1] : []
    content {
      host         = ""
      port         = var.port
      request_path = "/readyz"
    }
  }
}

resource "google_compute_region_backend_service" "backend" {
  name                  = local.name
  protocol              = "TCP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  health_checks         = [google_compute_region_health_check.health.id]
  port_name             = var.backend_port_name
  timeout_sec           = 240
  region                = var.region

  backend {
    group           = var.backend_instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }
}

resource "google_compute_region_target_tcp_proxy" "proxy" {
  name            = local.name
  region          = var.region
  backend_service = google_compute_region_backend_service.backend.id
}

# forwarding rule
resource "google_compute_forwarding_rule" "forwarding" {
  name                  = local.name
  network               = var.network
  subnetwork            = var.backend_subnet
  region                = var.region
  ip_address            = var.ip_address
  ip_protocol           = "TCP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = var.port
  allow_global_access   = true
  target                = google_compute_region_target_tcp_proxy.proxy.id
  labels                = var.frontend_labels
}
