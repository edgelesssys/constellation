terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.1.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

locals {
  group_uid = random_id.uid.hex
  name      = "${var.base_name}-${lower(var.role)}-${local.group_uid}"
}

resource "random_id" "uid" {
  byte_length = 4
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
    instance_metadata_tags      = "disabled"
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

  # See: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/launch_template#cpu-options
  cpu_options {
    # use "enabled" to enable SEV-SNP
    # use "disabled" to disable SEV-SNP (but still require SNP-capable hardware)
    # use null to leave the setting unset (allows non-SNP-capable hardware to be used)
    amd_sev_snp = var.enable_snp ? "enabled" : null
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      cpu_options,     # required. we cannot change the CPU options of a launch template
      name_prefix,     # required. Allow legacy scale sets to keep their old names
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
  desired_capacity    = var.initial_count
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
      name,                      # required. Allow legacy scale sets to keep their old names
      launch_template.0.version, # required. update procedure creates new versions of the launch template
      min_size,                  # required. autoscaling modifies the instance count externally
      max_size,                  # required. autoscaling modifies the instance count externally
      desired_capacity,          # required. autoscaling modifies the instance count externally
    ]
  }
}
