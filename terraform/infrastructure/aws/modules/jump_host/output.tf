output "ip" {
  value = aws_instance.jump_host.public_ip
  description = "Public IP of the jump host."
}
