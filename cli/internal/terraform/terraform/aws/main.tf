terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.4.3"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = var.region
}

locals {
  uid              = random_id.uid.hex
  name             = "${var.name}-${local.uid}"
  tag              = "constellation-${local.uid}"
  ports_node_range = "30000-32767"
  ports_ssh        = "22"

  ports_kubernetes   = "6443"
  ports_bootstrapper = "9000"
  ports_konnectivity = "8132"
  ports_verify       = "30081"
  ports_debugd       = "4000"

  disk_size = 10

  cidr_vpc_subnet_nodes    = "192.168.178.0/24"
  cidr_vpc_subnet_internet = "192.168.0.0/24"
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "aws_vpc" "vpc" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "${local.name}-vpc"
  }
}

# TODO: This dual subnet setup can end up in two different zones, and the LB needs IPs in each subnet to work. Pin a zone, or get this working properly with multiple zones?
resource "aws_subnet" "private" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = local.cidr_vpc_subnet_nodes
  tags = {
    Name = "${local.name}-subnet-nodes"
  }
}

resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = local.cidr_vpc_subnet_internet
  tags = {
    Name = "${local.name}-subnet-internet"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.name}-internet-gateway"
  }
}

resource "aws_nat_gateway" "gw" {
  subnet_id     = aws_subnet.public.id
  allocation_id = aws_eip.nat.id

  tags = {
    Name = "${local.name}-nat-gateway"
  }
}

resource "aws_route_table" "private_nat" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_nat_gateway.gw.id
  }

  tags = {
    Name = "${local.name}-nat-route"
  }
}

resource "aws_route_table" "public_igw" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = "${local.name}-nat-route"
  }
}

resource "aws_route_table_association" "private-nat" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private_nat.id
}

resource "aws_route_table_association" "route_to_internet" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public_igw.id
}

resource "aws_eip" "lb" {
  vpc = true
}

resource "aws_eip" "nat" {
  vpc = true
}

resource "aws_lb" "front_end" {
  name               = "${local.name}-loadbalancer"
  internal           = false
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = aws_subnet.public.id
    allocation_id = aws_eip.lb.id
  }

  tags = {
    Name = "loadbalancer"
  }

  enable_cross_zone_load_balancing = true
}

resource "aws_security_group" "security_group" {
  name        = local.name
  vpc_id      = aws_vpc.vpc.id
  description = "Security group for ${local.name}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  ingress {
    from_port   = split("-", local.ports_node_range)[0]
    to_port     = split("-", local.ports_node_range)[1]
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "K8s node ports"
  }

  # TODO REMOVE PLS PLS PLS PLS PLS
  ingress {
    from_port   = local.ports_ssh
    to_port     = local.ports_ssh
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH"
  }

  ingress {
    from_port   = local.ports_bootstrapper
    to_port     = local.ports_bootstrapper
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "bootstrapper"
  }

  ingress {
    from_port   = local.ports_kubernetes
    to_port     = local.ports_kubernetes
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "kubernetes"
  }

  ingress {
    from_port   = local.ports_konnectivity
    to_port     = local.ports_konnectivity
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "konnectivity"
  }

  ingress {
    from_port   = local.ports_debugd
    to_port     = local.ports_debugd
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "debugd"
  }

}

module "load_balancer_target_bootstrapper" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-bootstrapper"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_bootstrapper
}

module "load_balancer_target_kubernetes" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-kubernetes"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_kubernetes
}

module "load_balancer_target_verify" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-verify"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_verify
}

module "load_balancer_target_debugd" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-debugd"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_debugd
}

module "load_balancer_target_konnectivity" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-konnectivity"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_konnectivity
}

module "load_balancer_target_ssh" {
  source = "./modules/load_balancer_target"
  name   = "${local.name}-ssh"
  vpc    = aws_vpc.vpc.id
  lb_arn = aws_lb.front_end.arn
  port   = local.ports_ssh
}

module "instance_group_control_plane" {
  source = "./modules/instance_group"
  name   = local.name
  role   = "control-plane"

  uid            = local.uid
  instance_type  = var.instance_type
  instance_count = var.control_plane_count
  image_id       = var.ami
  disk_size      = local.disk_size

  target_group_arns = [
    module.load_balancer_target_bootstrapper.target_group_arn,
    module.load_balancer_target_kubernetes.target_group_arn,
    module.load_balancer_target_verify.target_group_arn,
    module.load_balancer_target_debugd.target_group_arn,
    module.load_balancer_target_konnectivity.target_group_arn,
    module.load_balancer_target_ssh.target_group_arn,
  ]
  security_groups      = [aws_security_group.security_group.id]
  subnetwork           = aws_subnet.private.id
  iam_instance_profile = var.iam_instance_profile_control_plane
}

module "instance_group_worker_nodes" {
  source               = "./modules/instance_group"
  name                 = local.name
  role                 = "worker"
  uid                  = local.uid
  instance_type        = var.instance_type
  instance_count       = var.worker_count
  image_id             = var.ami
  disk_size            = local.disk_size
  subnetwork           = aws_subnet.private.id
  target_group_arns    = []
  security_groups      = []
  iam_instance_profile = var.iam_instance_profile_worker_nodes
}
