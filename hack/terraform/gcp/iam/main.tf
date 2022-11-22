terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.44.0"
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

resource "google_project_iam_binding" "instance_admin_role" {
  project = var.project_id
  role    = "roles/compute.instanceAdmin.v1"

  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_project_iam_binding" "network_admin_role" {
  project = var.project_id
  role    = "roles/compute.networkAdmin"

  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_project_iam_binding" "security_admin_role" {
  project = var.project_id
  role    = "roles/compute.securityAdmin"

  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_project_iam_binding" "storage_admin_role" {
  project = var.project_id
  role    = "roles/compute.storageAdmin"

  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_project_iam_binding" "iam_service_account_user_role" {
  project = var.project_id
  role    = "roles/iam.serviceAccountUser"

  members = [
    "serviceAccount:${google_service_account.service_account.email}",
  ]
}

resource "google_service_account_key" "service_account_key" {
  service_account_id = google_service_account.service_account.name
}
