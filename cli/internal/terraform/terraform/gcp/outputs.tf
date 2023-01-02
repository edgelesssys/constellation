output "ip" {
  value = google_compute_global_address.loadbalancer_ip.address
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
