output "ip" {
  value = openstack_networking_floatingip_v2.public_ip.address
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
