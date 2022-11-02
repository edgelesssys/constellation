terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.36.1"
    }
  }
}

resource "aws_lb_target_group" "front_end" {
  name     = var.name
  port     = var.port
  protocol = "TCP"
  vpc_id   = var.vpc_id
  tags     = var.tags

  health_check {
    port                = var.port
    protocol            = var.healthcheck_protocol
    path                = var.healthcheck_protocol == "HTTPS" ? var.healthcheck_path : null
    interval            = 10
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = var.lb_arn
  port              = var.port
  protocol          = "TCP"
  tags              = var.tags

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.front_end.arn
  }
}
