output "service_account_key" {
  value       = google_service_account_key.service_account_key.private_key
  description = "Private key of the service account."
  sensitive   = true
}
