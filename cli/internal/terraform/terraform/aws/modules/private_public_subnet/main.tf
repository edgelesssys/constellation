terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

resource "aws_eip" "nat" {
  vpc = true
}

resource "aws_subnet" "private" {
  vpc_id            = var.vpc_id
  cidr_block        = var.cidr_vpc_subnet_nodes
  availability_zone = var.zone
  tags = {
    Name = "${var.name}-subnet-nodes"
  }
}

resource "aws_subnet" "public" {
  vpc_id            = var.vpc_id
  cidr_block        = var.cidr_vpc_subnet_internet
  availability_zone = var.zone
  tags = {
    Name = "${var.name}-subnet-internet"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = var.vpc_id

  tags = {
    Name = "${var.name}-internet-gateway"
  }
}

resource "aws_nat_gateway" "gw" {
  subnet_id     = aws_subnet.public.id
  allocation_id = aws_eip.nat.id

  tags = {
    Name = "${var.name}-nat-gateway"
  }
}

resource "aws_route_table" "private_nat" {
  vpc_id = var.vpc_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_nat_gateway.gw.id
  }

  tags = {
    Name = "${var.name}-nat-route"
  }
}

resource "aws_route_table" "public_igw" {
  vpc_id = var.vpc_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = "${var.name}-igw-route"
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
