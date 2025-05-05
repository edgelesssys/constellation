terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.97.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

provider "aws" {
  region = var.region
}

locals {
  uid                   = random_id.uid.hex
  name                  = "${var.name}-${local.uid}"
  init_secret_hash      = random_password.init_secret.bcrypt_hash
  cidr_vpc_subnet_nodes = "192.168.176.0/20"
  ports_node_range      = "30000-32767"
  load_balancer_ports = flatten([
    { name = "kubernetes", port = "6443", health_check = "HTTPS" },
    { name = "bootstrapper", port = "9000", health_check = "TCP" },
    { name = "verify", port = "30081", health_check = "TCP" },
    { name = "konnectivity", port = "8132", health_check = "TCP" },
    { name = "recovery", port = "9999", health_check = "TCP" },
    { name = "join", port = "30090", health_check = "TCP" },
    var.debug ? [{ name = "debugd", port = "4000", health_check = "TCP" }] : [],
    var.emergency_ssh ? [{ name = "ssh", port = "22", health_check = "TCP" }] : [],
  ])
  target_group_arns = {
    control-plane : [
      for port in local.load_balancer_ports : module.load_balancer_targets[port.name].target_group_arn
    ]
    worker : []
  }
  iam_instance_profile = {
    control-plane : var.iam_instance_profile_name_control_plane
    worker : var.iam_instance_profile_name_worker_nodes
  }
  # zones are all availability zones that are used by the node groups
  zones = distinct(sort([
    for node_group in var.node_groups : node_group.zone
  ]))
  // wildcard_lb_dns_name is the DNS name of the load balancer with a wildcard for the name.
  // example: given "name-1234567890.region.elb.amazonaws.com" it will return "*.region.elb.amazonaws.com"
  wildcard_lb_dns_name = replace(aws_lb.front_end.dns_name, "/^[^.]*\\./", "*.")

  tags = merge(
    var.additional_tags,
    { constellation-uid = local.uid }
  )

  in_cluster_endpoint     = aws_lb.front_end.dns_name
  out_of_cluster_endpoint = var.internal_load_balancer && var.debug ? module.jump_host[0].ip : local.in_cluster_endpoint
  revision                = 1
}

# A way to force replacement of resources if the provider does not want to replace them
# see: https://developer.hashicorp.com/terraform/language/resources/terraform-data#example-usage-data-for-replace_triggered_by
resource "terraform_data" "replacement" {
  input = local.revision
}

resource "random_id" "uid" {
  byte_length = 4
}

resource "random_password" "init_secret" {
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
  cidr_vpc_subnet_nodes    = local.cidr_vpc_subnet_nodes
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
  for_each = var.internal_load_balancer ? [] : toset([var.zone])
  domain   = "vpc"
  tags     = merge(local.tags, { "constellation-ip-endpoint" = each.key == var.zone ? "legacy-primary-zone" : "additional-zone" })
}

resource "aws_lb" "front_end" {
  name               = "${local.name}-loadbalancer"
  internal           = var.internal_load_balancer
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
      allocation_id = var.internal_load_balancer ? "" : aws_eip.lb[subnet_mapping.key].id
    }
  }
  enable_cross_zone_load_balancing = true

  lifecycle {
    ignore_changes = [security_groups]
  }
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

  dynamic "ingress" {
    for_each = local.load_balancer_ports
    content {
      description = ingress.value.name
      from_port   = ingress.value.port
      to_port     = ingress.value.port
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.vpc.cidr_block]
    description = "allow all internal"
  }

}

module "load_balancer_targets" {
  for_each             = { for port in local.load_balancer_ports : port.name => port }
  source               = "./modules/load_balancer_target"
  base_name            = "${local.name}-${each.value.name}"
  port                 = each.value.port
  healthcheck_protocol = each.value.health_check
  healthcheck_path     = each.value.name == "kubernetes" ? "/readyz" : ""
  vpc_id               = aws_vpc.vpc.id
  lb_arn               = aws_lb.front_end.arn
  tags                 = local.tags
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
  image_id             = var.image_id
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
    { constellation-init-secret-hash = local.init_secret_hash },
    { "kubernetes.io/cluster/${local.name}" = "owned" }
  )
}

module "jump_host" {
  count                = var.internal_load_balancer && var.debug ? 1 : 0
  source               = "./modules/jump_host"
  base_name            = local.name
  subnet_id            = module.public_private_subnet.public_subnet_id[var.zone]
  lb_internal_ip       = aws_lb.front_end.dns_name
  ports                = [for port in local.load_balancer_ports : port.port]
  security_groups      = [aws_security_group.security_group.id]
  iam_instance_profile = var.iam_instance_profile_name_worker_nodes
  additional_tags      = var.additional_tags
}
