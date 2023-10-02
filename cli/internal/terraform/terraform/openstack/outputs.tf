output "ip" {
  value = openstack_networking_floatingip_v2.public_ip.address
}

output "api_server_cert_sans" {
  value = sort(concat([openstack_networking_floatingip_v2.public_ip.address], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}

output "name" {
  value = local.name
}
