output "service_account_key" {
  value       = google_service_account_key.service_account_key.private_key
  description = "Private key of the service account."
  sensitive   = true
}

output "service_account_mail_vm" {
  value       = google_service_account.vm.email
  description = "Mail address of the service account to be attached to the VMs"
  sensitive   = false
}
