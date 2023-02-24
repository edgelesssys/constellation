terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.55.0"
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
    ignore_changes = [
      default_version, # required. update procedure creates new versions of the launch template
      image_id,        # required. update procedure modifies the image id externally
    ]
  }
}

resource "aws_autoscaling_group" "autoscaling_group" {
  name = local.name
  launch_template {
    id = aws_launch_template.launch_template.id
  }
  min_size            = 1
  max_size            = 10
  desired_capacity    = var.instance_count
  vpc_zone_identifier = [var.subnetwork]
  target_group_arns   = var.target_group_arns

  dynamic "tag" {
    for_each = var.tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      launch_template.0.version, # required. update procedure creates new versions of the launch template
      min_size,                  # required. autoscaling modifies the instance count externally
      max_size,                  # required. autoscaling modifies the instance count externally
      desired_capacity,          # required. autoscaling modifies the instance count externally
    ]
  }
}
