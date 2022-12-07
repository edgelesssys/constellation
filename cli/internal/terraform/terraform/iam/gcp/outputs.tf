output "sa_key" {
  value     = google_service_account_key.service_account_key.private_key
  sensitive = true
}
