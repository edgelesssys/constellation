output "ip" {
  value = aws_eip.lb.public_ip
}

output "uid" {
  value = local.uid
}

output "initSecret" {
  value     = random_password.initSecret.result
  sensitive = true
}
