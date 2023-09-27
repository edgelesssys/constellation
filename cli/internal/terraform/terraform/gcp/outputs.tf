output "ip" {
  value = var.internal_load_balancer ? google_compute_address.loadbalancer_ip_internal[0].address : google_compute_global_address.loadbalancer_ip[0].address
}

output "api_server_cert_sans" {
  value = sort(concat([var.internal_load_balancer ? google_compute_address.loadbalancer_ip_internal[0].address : google_compute_global_address.loadbalancer_ip[0].address], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "fallback_endpoint" {
  value = var.internal_load_balancer ? google_compute_address.loadbalancer_ip_internal[0].address : google_compute_global_address.loadbalancer_ip[0].address
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
