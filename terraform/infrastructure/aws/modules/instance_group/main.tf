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
    # Disable SMT. We are already disabling it inside the image.
    # Disabling SMT only in the image, not in the Hypervisor creates problems.
    # Thus, also disable it in the Hypervisor.
    # TODO(derpsteb): reenable once AWS confirms it's safe to do so.
    # threads_per_core = 1
    # When setting threads_per_core we also have to set core_count.
    # For the currently supported SNP instance families (C6a, M6a, R6a) default_cores
    # equals the maximum number of available cores.
    # core_count = data.aws_ec2_instance_type.instance_data.default_cores
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

  # TODO(msanft): Remove this (to have the 10m default) once AWS SEV-SNP boot problems are resolved.
  # Set a higher timeout for the ASG to fulfill the desired healthy capcity. Temporary workaround to
  # long boot times on SEV-SNP machines on AWS.
  wait_for_capacity_timeout = var.enable_snp ? "20m" : "10m"

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

data "aws_ec2_instance_type" "instance_data" {
  instance_type = var.instance_type
}
