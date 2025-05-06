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

resource "google_compute_health_check" "health" {
  name               = local.name
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

resource "google_compute_backend_service" "backend" {
  name                  = local.name
  protocol              = "TCP"
  load_balancing_scheme = "EXTERNAL"
  health_checks         = [google_compute_health_check.health.self_link]
  port_name             = var.backend_port_name
  timeout_sec           = 240

  dynamic "backend" {
    for_each = var.backend_instance_groups
    content {
      group          = backend.value
      balancing_mode = "UTILIZATION"
    }
  }
}

resource "google_compute_target_tcp_proxy" "proxy" {
  name            = local.name
  backend_service = google_compute_backend_service.backend.self_link
}

resource "google_compute_global_forwarding_rule" "forwarding" {
  name                  = local.name
  ip_address            = var.ip_address
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = var.port
  target                = google_compute_target_tcp_proxy.proxy.self_link
  labels                = var.frontend_labels
}
