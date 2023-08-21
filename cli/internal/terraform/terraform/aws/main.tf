terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.6.2"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
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
  ports_join         = "30090"
  target_group_arns = {
    control-plane : flatten([
      module.load_balancer_target_bootstrapper.target_group_arn,
      module.load_balancer_target_kubernetes.target_group_arn,
      module.load_balancer_target_verify.target_group_arn,
      module.load_balancer_target_recovery.target_group_arn,
      module.load_balancer_target_konnectivity.target_group_arn,
      module.load_balancer_target_join.target_group_arn,
      var.debug ? [module.load_balancer_target_debugd[0].target_group_arn] : [],
    ])
    worker : []
  }
  iam_instance_profile = {
    control-plane : var.iam_instance_profile_control_plane
    worker : var.iam_instance_profile_worker_nodes
  }
  # zones are all availability zones that are used by the node groups
  zones = distinct(sort([
    for node_group in var.node_groups : node_group.zone
  ]))
  // wildcard_lb_dns_name is the DNS name of the load balancer with a wildcard for the name.
  // example: given "name-1234567890.region.elb.amazonaws.com" it will return "*.region.elb.amazonaws.com"
  wildcard_lb_dns_name = replace(aws_lb.front_end.dns_name, "/^[^.]*\\./", "*.")

  tags = {
    constellation-uid = local.uid,
  }
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
  cidr_vpc_subnet_nodes    = "192.168.176.0/20"
  cidr_vpc_subnet_internet = "192.168.0.0/20"
  zone                     = var.zone
  zones                    = local.zones
  tags                     = local.tags
}

resource "aws_eip" "lb" {
  # TODO(malt3): use for_each = toset(module.public_private_subnet.all_zones)
  # in a future version to support all availability zones in the chosen region
  # This should only be done after we migrated to DNS-based addressing for the
  # control-plane.
  for_each = toset([var.zone])
  domain   = "vpc"
  tags     = merge(local.tags, { "constellation-ip-endpoint" = each.key == var.zone ? "legacy-primary-zone" : "additional-zone" })
}

resource "aws_lb" "front_end" {
  name               = "${local.name}-loadbalancer"
  internal           = false
  load_balancer_type = "network"
  tags               = local.tags
  security_groups    = [aws_security_group.security_group.id]

  dynamic "subnet_mapping" {
    # TODO(malt3): use for_each = toset(module.public_private_subnet.all_zones)
    # in a future version to support all availability zones in the chosen region
    # without needing to constantly replace the loadbalancer.
    # This has to wait until the bootstrapper that we upgrade from (source version) use
    # DNS-based addressing for the control-plane.
    # for_each = toset(module.public_private_subnet.all_zones)
    for_each = toset([var.zone])
    content {
      subnet_id     = module.public_private_subnet.public_subnet_id[subnet_mapping.key]
      allocation_id = aws_eip.lb[subnet_mapping.key].id
    }
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

module "load_balancer_target_join" {
  source               = "./modules/load_balancer_target"
  name                 = "${local.name}-join"
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  port                 = local.ports_join
  tags                 = local.tags
  healthcheck_protocol = "TCP"
}

module "instance_group" {
  source               = "./modules/instance_group"
  for_each             = var.node_groups
  base_name            = local.name
  node_group_name      = each.key
  role                 = each.value.role
  zone                 = each.value.zone
  uid                  = local.uid
  instance_type        = each.value.instance_type
  initial_count        = each.value.initial_count
  image_id             = var.ami
  state_disk_type      = each.value.disk_type
  state_disk_size      = each.value.disk_size
  target_group_arns    = local.target_group_arns[each.value.role]
  security_groups      = [aws_security_group.security_group.id]
  subnetwork           = module.public_private_subnet.private_subnet_id[each.value.zone]
  iam_instance_profile = local.iam_instance_profile[each.value.role]
  enable_snp           = var.enable_snp
  tags = merge(
    local.tags,
    { Name = "${local.name}-${each.value.role}" },
    { constellation-role = each.value.role },
    { constellation-node-group = each.key },
    { constellation-uid = local.uid },
    { constellation-init-secret-hash = local.initSecretHash },
    { "kubernetes.io/cluster/${local.name}" = "owned" }
  )
}
