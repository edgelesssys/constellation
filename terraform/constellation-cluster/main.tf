locals {
  yq_node_groups = join("\n", flatten([
    for name, group in var.node_groups : [
      "./yq eval '.nodeGroups.${name}.role = \"${group.role}\"' -i constellation-conf.yaml",
      "./yq eval '.nodeGroups.${name}.zone = \"${group.zone}\"' -i constellation-conf.yaml",
      "./yq eval '.nodeGroups.${name}.instanceType = \"${group.instance_type}\"' -i constellation-conf.yaml",
      "./yq eval '.nodeGroups.${name}.stateDiskSizeGB = ${group.disk_size}' -i constellation-conf.yaml",
      "./yq eval '.nodeGroups.${name}.stateDiskType = \"${group.disk_type}\"' -i constellation-conf.yaml",
      "./yq eval '.nodeGroups.${name}.initialCount = ${group.initial_count}' -i constellation-conf.yaml"
    ]
  ]))
}

resource "null_resource" "ensure_cli" {
  provisioner "local-exec" {
    command = <<EOT
         ${path.module}/install-constellation.sh
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
}

// terraform_data resource so that it is run only once
resource "terraform_data" "config_generate" {
  provisioner "local-exec" {
    command = <<EOT
         ./constellation config generate ${var.csp}
    EOT
  }
  depends_on = [
    null_resource.ensure_cli
  ]
}

resource "null_resource" "aws_config" {
  count = var.aws_config != null ? 1 : 0
  provisioner "local-exec" {
    command = <<EOT
      ./yq eval '.provider.aws.region = "${var.aws_config.region}"' -i constellation-conf.yaml
      ./yq eval '.provider.aws.zone = "${var.aws_config.zone}"' -i constellation-conf.yaml
      ./yq eval '.provider.aws.iamProfileControlPlane = "${var.aws_config.iam_instance_profile_control_plane}"' -i constellation-conf.yaml
      ./yq eval '.provider.aws.iamProfileWorkerNodes = "${var.aws_config.iam_instance_profile_worker_nodes}"' -i constellation-conf.yaml
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
  depends_on = [
    terraform_data.config_generate
  ]
}

resource "null_resource" "gcp_config" {
  count = var.gcp_config != null ? 1 : 0
  provisioner "local-exec" {
    command = <<EOT
      ./yq eval '.provider.gcp.project = "${var.gcp_config.project}"' -i constellation-conf.yaml
      ./yq eval '.provider.gcp.region = "${var.gcp_config.region}"' -i constellation-conf.yaml
      ./yq eval '.provider.gcp.zone = "${var.gcp_config.zone}"' -i constellation-conf.yaml
      ./yq eval '.provider.gcp.serviceAccountKeyPath = "${var.gcp_config.serviceAccountKeyPath}"' -i constellation-conf.yaml

      ./yq eval '.infrastructure.gcp.projectID = "${var.gcp_config.project}"' -i constellation-state.yaml
      ./yq eval '.infrastructure.gcp.ipCidrPod = "${var.gcp_config.ipCidrPod}"' -i constellation-state.yaml
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
  depends_on = [
    terraform_data.config_generate
  ]
}

resource "null_resource" "config" {
  provisioner "local-exec" {
    command = <<EOT
      ./yq eval '.name = "${var.name}"' -i constellation-conf.yaml
      if [ "${var.image}" != "" ]; then
      ./yq eval '.image = "${var.image}"' -i constellation-conf.yaml
      fi
      if [ "${var.kubernetes_version}" != "" ]; then
      ./yq eval '.kubernetesVersion = "${var.kubernetes_version}"' -i constellation-conf.yaml
      fi
      if [ "${var.microservice_version}" != "" ]; then
      ./yq eval '.microserviceVersion = "${var.microservice_version}"' -i constellation-conf.yaml
      fi
      ${local.yq_node_groups}
      ./constellation config fetch-measurements
    EOT
  }

  depends_on = [
    null_resource.aws_config, null_resource.gcp_config
  ]

  triggers = {
    always_run = timestamp()
  }
}


resource "null_resource" "infra_state" {
  provisioner "local-exec" {
    command = <<EOT
      ./yq eval '.infrastructure.uid = "${var.uid}"' -i constellation-state.yaml
      ./yq eval '.infrastructure.inClusterEndpoint = "${var.inClusterEndpoint}"' -i constellation-state.yaml
      ./yq eval '.infrastructure.clusterEndpoint = "${var.clusterEndpoint}"' -i constellation-state.yaml
      ./yq eval '.infrastructure.initSecret = "'"$(echo "${var.initSecretHash}" | tr -d '\n' | hexdump -ve '/1 "%02x"')"'"' -i constellation-state.yaml
      ./yq eval '.infrastructure.apiServerCertSANs = ${jsonencode(var.apiServerCertSANs)}' -i constellation-state.yaml
      ./yq eval '.infrastructure.name = "${var.name}"' -i constellation-state.yaml
      ./yq eval '.infrastructure.ipCidrNode = "${var.ipCidrNode}"' -i constellation-state.yaml
    EOT
  }
  depends_on = [
    terraform_data.config_generate
  ]
  triggers = {
    always_run = timestamp()
  }
}


resource "null_resource" "apply" {
  provisioner "local-exec" {
    command = "./constellation apply --debug --yes --skip-phases infrastructure"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "./constellation terminate --yes && rm constellation-conf.yaml constellation-mastersecret.json && rm -r constellation-upgrade"
  }

  depends_on = [
    null_resource.infra_state, null_resource.config, null_resource.ensure_cli
  ]
  triggers = {
    always_run = timestamp()
  }
}
