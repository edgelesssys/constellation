terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.58.0"
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
  uid                = random_id.uid.hex
  name               = "${var.name}-${local.uid}"
  initSecretHash     = random_password.initSecret.bcrypt_hash
  ports_node_range   = "30000-32767"
  ports_kubernetes   = "6443"
  ports_bootstrapper = "9000"
  ports_konnectivity = "8132"
  ports_verify       = "30081"
  ports_recovery     = "9999"
  ports_debugd       = "4000"

  tags = { constellation-uid = local.uid }
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "initSecret" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "aws_vpc" "vpc" {
  cidr_block = "192.168.0.0/16"
  tags       = merge(local.tags, { Name = "${local.name}-vpc" })
}

module "public_private_subnet" {
  source                   = "./modules/public_private_subnet"
  name                     = local.name
  vpc_id                   = aws_vpc.vpc.id
  cidr_vpc_subnet_nodes    = "192.168.178.0/24"
  cidr_vpc_subnet_internet = "192.168.0.0/24"
  zone                     = var.zone
  tags                     = local.tags
}

resource "aws_eip" "lb" {
  vpc  = true
  tags = local.tags
}

resource "aws_lb" "front_end" {
  name               = "${local.name}-loadbalancer"
  internal           = false
  load_balancer_type = "network"
  tags               = local.tags

  subnet_mapping {
    subnet_id     = module.public_private_subnet.public_subnet_id
    allocation_id = aws_eip.lb.id
  }
  enable_cross_zone_load_balancing = true
}

resource "aws_security_group" "security_group" {
  name        = local.name
  vpc_id      = aws_vpc.vpc.id
  description = "Security group for ${local.name}"
  tags        = local.tags

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
    from_port   = local.ports_recovery
    to_port     = local.ports_recovery
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "recovery"
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.vpc.cidr_block]
    description = "allow all internal"
  }

  dynamic "ingress" {
    for_each = var.debug ? [1] : []
    content {
      from_port   = local.ports_debugd
      to_port     = local.ports_debugd
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
      description = "debugd"
    }
  }
}

resource "aws_cloudwatch_log_group" "log_group" {
  name              = local.name
  retention_in_days = 30
  tags              = local.tags
}

module "load_balancer_target_bootstrapper" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-bootstrapper"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_bootstrapper
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "load_balancer_target_kubernetes" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-kubernetes"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_kubernetes
  tags                 = local.tags
  healthcheck_protocol = "HTTPS"
  healthcheck_path     = "/readyz"
}

module "load_balancer_target_verify" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-verify"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_verify
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "load_balancer_target_recovery" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-recovery"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_recovery
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "load_balancer_target_debugd" {
  count                = var.debug ? 1 : 0 // only deploy debugd in debug mode
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-debugd"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_debugd
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "load_balancer_target_konnectivity" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-konnectivity"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_konnectivity
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "instance_group_control_plane" {
  source          = "./modules/instance_group"
  name            = local.name
  role            = "control-plane"
  uid             = local.uid
  instance_type   = var.instance_type
  instance_count  = var.control_plane_count
  image_id        = var.ami
  state_disk_type = var.state_disk_type
  state_disk_size = var.state_disk_size
  target_group_arns = flatten([
    module.load_balancer_target_bootstrapper.target_group_arn,
    module.load_balancer_target_kubernetes.target_group_arn,
    module.load_balancer_target_verify.target_group_arn,
    module.load_balancer_target_recovery.target_group_arn,
    module.load_balancer_target_konnectivity.target_group_arn,
    var.debug ? [module.load_balancer_target_debugd[0].target_group_arn] : [],
  ])
  security_groups      = [aws_security_group.security_group.id]
  subnetwork           = module.public_private_subnet.private_subnet_id
  iam_instance_profile = var.iam_instance_profile_control_plane
  tags = merge(
    local.tags,
    { Name = local.name },
    { constellation-role = "control-plane" },
    { constellation-uid = local.uid },
    { KubernetesCluster = "Constellation-${local.uid}" },
    { constellation-init-secret-hash = local.initSecretHash }
  )
}

module "instance_group_worker_nodes" {
  source               = "./modules/instance_group"
  name                 = local.name
  role                 = "worker"
  uid                  = local.uid
  instance_type        = var.instance_type
  instance_count       = var.worker_count
  image_id             = var.ami
  state_disk_type      = var.state_disk_type
  state_disk_size      = var.state_disk_size
  subnetwork           = module.public_private_subnet.private_subnet_id
  target_group_arns    = []
  security_groups      = [aws_security_group.security_group.id]
  iam_instance_profile = var.iam_instance_profile_worker_nodes
  tags = merge(
    local.tags,
    { Name = local.name },
    { constellation-role = "worker" },
    { constellation-uid = local.uid },
    { KubernetesCluster = "Constellation-${local.uid}" },
    { constellation-init-secret-hash = local.initSecretHash }
  )
}
