terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.97.0"
    }
  }
}

locals {
  # az_number is a stable mapping of az suffix to a number used for calculating the subnet cidr
  az_number = {
    # we start counting at 2 to have the legacy subnet before the first newly created networks
    # the legacy subnet did not start at a /20 boundary
    # 0 => 192.168.176.0/24 (unused private subnet cidr)
    # 1 => 192.168.177.0/24 (unused private subnet cidr)
    legacy = 2 # => 192.168.178.0/24 (legacy private subnet)
    a      = 3 # => 192.168.179.0/24 (first newly created zonal private subnet)
    b      = 4
    c      = 5
    d      = 6
    e      = 7
    f      = 8
    g      = 9
    h      = 10
    i      = 11
    j      = 12
    k      = 13
    l      = 14
    m      = 15 # => 192.168.191.0/24 (last reserved zonal private subnet cidr). In reality, AWS doesn't have that many zones in a region.
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_availability_zone" "all" {
  for_each = toset(data.aws_availability_zones.available.names)

  name = each.key
}

resource "aws_eip" "nat" {
  for_each = toset(var.zones)
  domain   = "vpc"
  tags     = var.tags
}

resource "aws_subnet" "private" {
  for_each          = data.aws_availability_zone.all
  vpc_id            = var.vpc_id
  cidr_block        = cidrsubnet(var.cidr_vpc_subnet_nodes, 4, local.az_number[each.value.name_suffix])
  availability_zone = each.key
  tags              = merge(var.tags, { Name = "${var.name}-subnet-nodes" }, { "kubernetes.io/role/internal-elb" = 1 }) # aws-load-balancer-controller needs role annotation
  lifecycle {
    ignore_changes = [
      cidr_block, # required. Legacy subnets used fixed cidr blocks for the single zone that don't match the new scheme.
    ]
  }
}

resource "aws_subnet" "public" {
  for_each          = data.aws_availability_zone.all
  vpc_id            = var.vpc_id
  cidr_block        = cidrsubnet(var.cidr_vpc_subnet_internet, 4, local.az_number[each.value.name_suffix])
  availability_zone = each.key
  tags              = merge(var.tags, { Name = "${var.name}-subnet-internet" }, { "kubernetes.io/role/elb" = 1 }) # aws-load-balancer-controller needs role annotation
  lifecycle {
    ignore_changes = [
      cidr_block, # required. Legacy subnets used fixed cidr blocks for the single zone that don't match the new scheme.
    ]
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = var.vpc_id
  tags   = merge(var.tags, { Name = "${var.name}-internet-gateway" })
}

resource "aws_nat_gateway" "gw" {
  for_each      = toset(var.zones)
  subnet_id     = aws_subnet.public[each.key].id
  allocation_id = aws_eip.nat[each.key].id
  tags          = merge(var.tags, { Name = "${var.name}-nat-gateway" })
}

resource "aws_route_table" "private_nat" {
  for_each = toset(var.zones)
  vpc_id   = var.vpc_id
  tags     = merge(var.tags, { Name = "${var.name}-private-nat" })

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.gw[each.key].id
  }
}

resource "aws_route_table" "public_igw" {
  for_each = toset(var.zones)
  vpc_id   = var.vpc_id
  tags     = merge(var.tags, { Name = "${var.name}-public-igw" })

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }
}

resource "aws_route_table_association" "private_nat" {
  for_each       = toset(var.zones)
  subnet_id      = aws_subnet.private[each.key].id
  route_table_id = aws_route_table.private_nat[each.key].id
}

resource "aws_route_table_association" "route_to_internet" {
  for_each       = toset(var.zones)
  subnet_id      = aws_subnet.public[each.key].id
  route_table_id = aws_route_table.public_igw[each.key].id
}
