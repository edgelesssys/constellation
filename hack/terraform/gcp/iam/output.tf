output "sa_key" {
  value     = google_service_account_key.service_account_key.private_key
  sensitive = true
}

output "region" {
  value = var.region
}

output "zone" {
  value = var.zone
}

output "project_id" {
  value = var.project_id
}
