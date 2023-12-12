output "out_of_cluster_endpoint" {
  value = openstack_networking_floatingip_v2.public_ip.address
}

output "in_cluster_endpoint" {
  value = openstack_networking_floatingip_v2.public_ip.address
}

output "extra_api_server_cert_sans" {
  value = sort(concat([openstack_networking_floatingip_v2.public_ip.address], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "uid" {
  value = local.uid
}

output "init_secret" {
  value     = random_password.init_secret.result
  sensitive = true
}

output "name" {
  value = local.name
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
}
