output "ip" {
  value = local.output_ip
}

output "api_server_cert_sans" {
  value = sort(concat([
    local.output_ip,
    ],
  var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
}

output "fallback_endpoint" {
  value = local.output_ip
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
