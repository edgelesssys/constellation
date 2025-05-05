terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.97.0"
    }
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

resource "aws_instance" "jump_host" {
  ami                         = data.aws_ami.ubuntu.id
  instance_type               = "c5a.large"
  associate_public_ip_address = true

  iam_instance_profile   = var.iam_instance_profile
  subnet_id              = var.subnet_id
  vpc_security_group_ids = var.security_groups

  tags = merge(var.additional_tags, {
    "Name" = "${var.base_name}-jump-host",
  })

  user_data = <<EOF
#!/bin/bash
set -x

# Uncomment to create user with password
# useradd -m user
# usermod -aG sudo user
# usermod --shell /bin/bash user
# sh -c "echo \"user:pass\" | chpasswd"

sysctl -w net.ipv4.ip_forward=1
sysctl -p

internal_ip=$(ip route get 8.8.8.8 | grep -oP 'src \K[^ ]+')

lb_ip=${var.lb_internal_ip}
if [[ ! $${lb_ip} =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  lb_ip=$(dig +short ${var.lb_internal_ip})
fi
%{for port in var.ports~}
iptables -t nat -A PREROUTING -p tcp --dport ${port} -j DNAT --to-destination $${lb_ip}:${port}
iptables -t nat -A POSTROUTING -p tcp -d $${lb_ip} --dport ${port} -j SNAT --to-source $${internal_ip}
%{endfor~}
EOF
}
