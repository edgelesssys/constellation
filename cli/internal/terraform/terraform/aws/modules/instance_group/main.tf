terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.38.0"
    }
  }
}

locals {
  name = "${var.name}-${lower(var.role)}"
}


resource "aws_launch_template" "launch_template" {
  name_prefix   = local.name
  image_id      = var.image_id
  instance_type = var.instance_type
  iam_instance_profile {
    name = var.iam_instance_profile
  }
  vpc_security_group_ids = var.security_groups
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    instance_metadata_tags      = "enabled"
    http_put_response_hop_limit = 2
  }

  block_device_mappings {
    device_name = "/dev/sdb"
    ebs {
      volume_size           = var.state_disk_size
      volume_type           = var.state_disk_type
      encrypted             = true
      delete_on_termination = true
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "control_plane_autoscaling_group" {
  name = local.name
  launch_template {
    id = aws_launch_template.launch_template.id
  }
  min_size            = 1
  max_size            = 10
  desired_capacity    = var.instance_count
  vpc_zone_identifier = [var.subnetwork]
  target_group_arns   = var.target_group_arns

  lifecycle {
    create_before_destroy = true
  }

  tag {
    key                 = "Name"
    value               = local.name
    propagate_at_launch = true
  }
  tag {
    key                 = "constellation-role"
    value               = var.role
    propagate_at_launch = true
  }
  tag {
    key                 = "constellation-uid"
    value               = var.uid
    propagate_at_launch = true
  }

  tag {
    key                 = "KubernetesCluster"
    value               = "Constellation-${var.uid}"
    propagate_at_launch = true
  }
}
