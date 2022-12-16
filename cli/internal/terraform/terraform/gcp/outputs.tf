output "ip" {
  value = google_compute_address.loadbalancer_ip.address
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
