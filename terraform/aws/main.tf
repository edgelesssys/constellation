terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.4.1"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-2"
}

locals {
  uid              = random_id.uid.hex
  name             = "${var.name}-${local.uid}"
  tag              = "constellation-${local.uid}"
  ports_node_range = "30000-32767"
  ports_ssh        = "22"

  ports_kubernets    = "6443"
  ports_bootstrapper = "9000"
  ports_konnectivity = "8132"
  ports_verify       = "30081"
  ports_debugd       = "4000"

  cidr_vpc_subnet_nodes = "192.168.178.0/24"
  count_control_plane   = 1
  count_worker          = 1
  image_id              = "ami-02f3416038bdb17fb" // Ubuntu 22.04 LTS
  instance_type         = "t2.micro"
  disk_size             = 30
}

resource "random_id" "uid" {
  byte_length = 8
}

resource "aws_vpc" "vpc" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "${local.name}-vpc"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = local.cidr_vpc_subnet_nodes
  tags = {
    Name = "${local.name}-subnet"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.name}-gateway"
  }
}

resource "aws_security_group" "security_group" {
  name   = local.name
  vpc_id = aws_vpc.vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = split("-", local.ports_node_range)[0]
    to_port     = split("-", local.ports_node_range)[1]
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = local.ports_bootstrapper
    to_port     = local.ports_bootstrapper
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = local.ports_kubernets
    to_port     = local.ports_kubernets
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = local.ports_konnectivity
    to_port     = local.ports_konnectivity
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = local.ports_debugd
    to_port     = local.ports_debugd
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

}

module "load_balancer_bootstrapper" {
  source = "./modules/load_balancer"
  name   = "bootstrapper"
  vpc    = aws_vpc.vpc.id
  subnet = aws_subnet.main.id
  port   = local.ports_bootstrapper
}

module "load_balancer_kubernetes" {
  source = "./modules/load_balancer"
  name   = "kubernetes"
  vpc    = aws_vpc.vpc.id
  subnet = aws_subnet.main.id
  port   = local.ports_kubernets
}

module "load_balancer_verify" {
  source = "./modules/load_balancer"
  name   = "verify"
  vpc    = aws_vpc.vpc.id
  subnet = aws_subnet.main.id
  port   = local.ports_verify
}

module "load_balancer_debugd" {
  source = "./modules/load_balancer"
  name   = "debugd"
  vpc    = aws_vpc.vpc.id
  subnet = aws_subnet.main.id
  port   = local.ports_debugd
}

module "load_balancer_konnectivity" {
  source = "./modules/load_balancer"
  name   = "konnectivity"
  vpc    = aws_vpc.vpc.id
  subnet = aws_subnet.main.id
  port   = local.ports_konnectivity
}

module "instance_group_control_plane" {
  source = "./modules/instance_group"
  name   = local.name
  role   = "control-plane"

  uid            = local.uid
  instance_type  = local.instance_type
  instance_count = local.count_control_plane
  image_id       = local.image_id
  disk_size      = local.disk_size
  target_group_arns = [
    module.load_balancer_bootstrapper.target_group_arn,
    module.load_balancer_kubernetes.target_group_arn,
    module.load_balancer_verify.target_group_arn,
    module.load_balancer_debugd.target_group_arn
  ]
  subnetwork           = aws_subnet.main.id
  iam_instance_profile = var.control_plane_iam_instance_profile
}

module "instance_group_worker_nodes" {
  source               = "./modules/instance_group"
  name                 = local.name
  role                 = "worker"
  uid                  = local.uid
  instance_type        = local.instance_type
  instance_count       = local.count_worker
  image_id             = local.image_id
  disk_size            = local.disk_size
  subnetwork           = aws_subnet.main.id
  target_group_arns    = []
  iam_instance_profile = var.worker_nodes_iam_instance_profile
}
