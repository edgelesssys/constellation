output "ip" {
  value = module.control_plane.instance_ips[0]
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
