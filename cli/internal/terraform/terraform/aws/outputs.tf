output "ip" {
  value = aws_eip.lb[var.zone].public_ip
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
