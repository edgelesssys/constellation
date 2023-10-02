output "ip" {
  value = google_compute_global_address.loadbalancer_ip.address
}

output "api_server_cert_sans" {
  value = sort(concat([google_compute_global_address.loadbalancer_ip.address], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "fallback_endpoint" {
  value = google_compute_global_address.loadbalancer_ip.address
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}

output "project" {
  value = var.project
}

output "ip_cidr_nodes" {
  value = local.cidr_vpc_subnet_nodes
}

output "ip_cidr_pods" {
  value = local.cidr_vpc_subnet_pods
}

output "name" {
  value = local.name
}
