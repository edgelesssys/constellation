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

resource "random_id" "uid" {
  byte_length = 8
}

resource "aws_iam_instance_profile" "control_plane_instance_profile" {
  name = "${var.name_prefix}_control_plane_instance_profile"
  role = aws_iam_role.control_plane_role.name
}

resource "aws_iam_role" "control_plane_role" {
  name = "${var.name_prefix}_control_plane_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "control_plane_policy" {
  name   = "${var.name_prefix}_control_plane_policy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:DescribeTargetGroupAttributes",
        "elasticloadbalancing:DescribeRules",
        "shield:GetSubscriptionState",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:ModifyTargetGroupAttributes",
        "elasticloadbalancing:DescribeTags",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLaunchConfigurations",
        "autoscaling:DescribeTags",
        "ec2:AttachVolume",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CreateRoute",
        "ec2:CreateSecurityGroup",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:DeleteRoute",
        "ec2:DeleteSecurityGroup",
        "ec2:DeleteVolume",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeImages",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes",
        "ec2:DescribeVpcs",
        "ec2:DetachVolume",
        "ec2:ModifyInstanceAttribute",
        "ec2:ModifyVolume",
        "ec2:RevokeSecurityGroupIngress",
        "elasticloadbalancing:AddTags",
        "elasticloadbalancing:AddTags",
        "elasticloadbalancing:ApplySecurityGroupsToLoadBalancer",
        "elasticloadbalancing:AttachLoadBalancerToSubnets",
        "elasticloadbalancing:ConfigureHealthCheck",
        "elasticloadbalancing:CreateListener",
        "elasticloadbalancing:CreateLoadBalancer",
        "elasticloadbalancing:CreateLoadBalancerListeners",
        "elasticloadbalancing:CreateLoadBalancerPolicy",
        "elasticloadbalancing:CreateTargetGroup",
        "elasticloadbalancing:DeleteListener",
        "elasticloadbalancing:DeleteLoadBalancer",
        "elasticloadbalancing:DeleteLoadBalancerListeners",
        "elasticloadbalancing:DeleteTargetGroup",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:DescribeLoadBalancerAttributes",
        "elasticloadbalancing:DescribeLoadBalancerPolicies",
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:DescribeTargetHealth",
        "elasticloadbalancing:DetachLoadBalancerFromSubnets",
        "elasticloadbalancing:ModifyListener",
        "elasticloadbalancing:ModifyLoadBalancerAttributes",
        "elasticloadbalancing:ModifyTargetGroup",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:RegisterTargets",
        "elasticloadbalancing:SetLoadBalancerPoliciesForBackendServer",
        "elasticloadbalancing:SetLoadBalancerPoliciesOfListener",
        "iam:CreateServiceLinkedRole",
        "kms:DescribeKey",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:ListTagsLogGroup",
        "logs:PutLogEvents",
        "tag:GetResources",
        "ec2:DescribeLaunchTemplateVersions",
        "autoscaling:SetDesiredCapacity",
        "autoscaling:TerminateInstanceInAutoScalingGroup",
        "ec2:DescribeInstanceStatus",
        "ec2:CreateLaunchTemplateVersion",
        "ec2:ModifyLaunchTemplate"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attach_control_plane_policy" {
  role       = aws_iam_role.control_plane_role.name
  policy_arn = aws_iam_policy.control_plane_policy.arn
}

resource "aws_iam_instance_profile" "worker_node_instance_profile" {
  name = "${var.name_prefix}_worker_node_instance_profile"
  role = aws_iam_role.worker_node_role.name
}

resource "aws_iam_role" "worker_node_role" {
  name = "${var.name_prefix}_worker_node_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "worker_node_policy" {
  name   = "${var.name_prefix}_worker_node_policy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeImages",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ecr:BatchCheckLayerAvailability",
        "ecr:BatchGetImage",
        "ecr:DescribeRepositories",
        "ecr:GetAuthorizationToken",
        "ecr:GetDownloadUrlForLayer",
        "ecr:GetRepositoryPolicy",
        "ecr:ListImages",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:ListTagsLogGroup",
        "logs:PutLogEvents",
        "tag:GetResources"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attach_worker_node_policy" {
  role       = aws_iam_role.worker_node_role.name
  policy_arn = aws_iam_policy.worker_node_policy.arn
}

// Add all permissions here, which are needed by the bootstrapper
resource "aws_iam_policy" "constellation_bootstrapper_policy" {
  name   = "${var.name_prefix}_constellation_bootstrapper_policy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:DescribeLoadBalancers"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attach_bootstrapper_policy_worker" {
  role       = aws_iam_role.worker_node_role.name
  policy_arn = aws_iam_policy.constellation_bootstrapper_policy.arn
}

resource "aws_iam_role_policy_attachment" "attach_bootstrapper_policy_control_plane" {
  role       = aws_iam_role.control_plane_role.name
  policy_arn = aws_iam_policy.constellation_bootstrapper_policy.arn
}

// TODO(msanft): incorporate this into the custom worker node policy
resource "aws_iam_role_policy_attachment" "csi_driver_policy_worker" {
  role       = aws_iam_role.worker_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
}

// TODO(msanft): incorporate this into the custom control-plane node policy
resource "aws_iam_role_policy_attachment" "csi_driver_policy_control_plane" {
  role       = aws_iam_role.control_plane_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
}

// This policy is required by the AWS load balancer controller and can be found at
// https://github.com/kubernetes-sigs/aws-load-balancer-controller/blob/b44633a/docs/install/iam_policy.json.
resource "aws_iam_policy" "lb_policy" {
  name   = "${var.name_prefix}_lb_policy"
  policy = file("${path.module}/alb_policy.json")
}

resource "aws_iam_role_policy_attachment" "attach_lb_policy_worker" {
  role       = aws_iam_role.worker_node_role.name
  policy_arn = aws_iam_policy.lb_policy.arn
}

resource "aws_iam_role_policy_attachment" "attach_lb_policy_control_plane" {
  role       = aws_iam_role.control_plane_role.name
  policy_arn = aws_iam_policy.lb_policy.arn
}
