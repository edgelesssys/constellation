locals {

  image_ref         = startswith(var.image, "v") ? "ref/-/stream/stable/${var.image}" : var.image
  fetch_ami_command = <<EOT
    curl -s https://cdn.confidential.cloud/constellation/v2/${local.image_ref}/image/info.json | \
    ./yq eval '.list[] | select(.csp == "aws" and .attestationVariant == "${var.attestation_variant}" and .region == "${var.region}") | .reference' - | tr -d '\n' > "${path.module}/ami.txt"
    echo -n "AMI: "
    cat "${path.module}/ami.txt"
  EOT
}

resource "null_resource" "fetch_ami" {
  provisioner "local-exec" {
    command = local.fetch_ami_command

    environment = {
      attestation_variant = var.attestation_variant
    }
  }
  triggers = {
    always_run = "${timestamp()}"
  }
}

data "local_file" "ami" {
  filename   = "${path.module}/ami.txt"
  depends_on = [null_resource.fetch_ami]
}
