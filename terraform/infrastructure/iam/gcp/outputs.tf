output "sa_key" {
  value     = google_service_account_key.service_account_key.private_key
  description = "Service account key for the service account being created."
  sensitive = true
}
