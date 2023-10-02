output "ip" {
  value = aws_eip.lb[var.zone].public_ip
}

output "api_server_cert_sans" {
  value = sort(concat([aws_eip.lb[var.zone].public_ip, local.wildcard_lb_dns_name], var.custom_endpoint == "" ? [] : [var.custom_endpoint]))
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
