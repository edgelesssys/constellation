terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

resource "aws_lb" "front_end" {
  name               = var.name
  internal           = false
  load_balancer_type = "network"
  subnets            = [var.subnet]

  tags = {
    Name = "loadbalancer"
  }

  enable_cross_zone_load_balancing = true
}

resource "aws_lb_target_group" "front_end" {
  name     = var.name
  port     = var.port
  protocol = "TCP"
  vpc_id   = var.vpc

  health_check {
    port     = var.port
    protocol = "TCP"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = var.port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.front_end.arn
  }
}
