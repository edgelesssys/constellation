terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.23.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

resource "google_service_account" "service_account" {
  account_id   = var.service_account_id
  display_name = "Constellation service account"
  description  = "Service account used inside Constellation"
}

// service_account creation is eventually consistent so add a delay to ensure it is created before the next step: https://registry.terraform.io/providers/hashicorp/google/4.69.1/docs/resources/google_service_account.html
resource "null_resource" "delay" {
  provisioner "local-exec" {
    command = "sleep 15"
  }
  triggers = {
    "service_account" = "${google_service_account.service_account.id}"
  }
}


resource "google_project_iam_member" "instance_admin_role" {
  project    = var.project_id
  role       = "roles/compute.instanceAdmin.v1"
  member     = "serviceAccount:${google_service_account.service_account.email}"
  depends_on = [null_resource.delay]
}

resource "google_project_iam_member" "network_admin_role" {
  project    = var.project_id
  role       = "roles/compute.networkAdmin"
  member     = "serviceAccount:${google_service_account.service_account.email}"
  depends_on = [null_resource.delay]
}

resource "google_project_iam_member" "security_admin_role" {
  project    = var.project_id
  role       = "roles/compute.securityAdmin"
  member     = "serviceAccount:${google_service_account.service_account.email}"
  depends_on = [null_resource.delay]
}

resource "google_project_iam_member" "storage_admin_role" {
  project    = var.project_id
  role       = "roles/compute.storageAdmin"
  member     = "serviceAccount:${google_service_account.service_account.email}"
  depends_on = [null_resource.delay]
}

resource "google_project_iam_member" "iam_service_account_user_role" {
  project    = var.project_id
  role       = "roles/iam.serviceAccountUser"
  member     = "serviceAccount:${google_service_account.service_account.email}"
  depends_on = [null_resource.delay]
}

resource "google_service_account_key" "service_account_key" {
  service_account_id = google_service_account.service_account.name
  depends_on         = [null_resource.delay]
}
