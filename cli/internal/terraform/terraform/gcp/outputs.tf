output "ip" {
  value = google_compute_global_address.loadbalancer_ip.address
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
