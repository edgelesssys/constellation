terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.36.1"
    }
  }
}

locals {
  name = "${var.name}-${lower(var.role)}"
}


resource "aws_launch_configuration" "control_plane_launch_config" {
  name_prefix          = local.name
  image_id             = var.image_id
  instance_type        = var.instance_type
  iam_instance_profile = var.iam_instance_profile
  security_groups      = var.security_groups
  metadata_options {
    http_tokens = "required"
  }

  root_block_device {
    encrypted = true
  }

  ebs_block_device {
    device_name           = "/dev/sdb" # Note: AWS may adjust this to /dev/xvdb, /dev/hdb or /dev/nvme1n1 depending on the disk type. See: https://docs.aws.amazon.com/en_us/AWSEC2/latest/UserGuide/device_naming.html
    volume_size           = var.state_disk_size
    volume_type           = var.state_disk_type
    encrypted             = true
    delete_on_termination = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "control_plane_autoscaling_group" {
  name                 = local.name
  launch_configuration = aws_launch_configuration.control_plane_launch_config.name
  min_size             = 1
  max_size             = 10
  desired_capacity     = var.instance_count
  vpc_zone_identifier  = [var.subnetwork]
  target_group_arns    = var.target_group_arns

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
}
