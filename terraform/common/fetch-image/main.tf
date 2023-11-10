locals {
  image_ref     = startswith(var.image, "v") ? "ref/-/stream/stable/${var.image}" : var.image
  region_filter = var.region != "" ? " and .region == \"${var.region}\"" : ""

  fetch_image_command = <<EOT
    curl -s https://cdn.confidential.cloud/constellation/v2/${local.image_ref}/image/info.json | \
    ./yq eval '.list[] | select(.csp == "${var.csp}" and .attestationVariant == "${var.attestation_variant}"${local.region_filter}) | .reference' - | tr -d '\n' > "image.txt"

    if [ '${var.csp}' = 'azure' ]; then
      sed -i 's/CommunityGalleries/communityGalleries/g' image.txt
      sed -i 's/Images/images/g' image.txt
      sed -i 's/Versions/versions/g' image.txt
    fi
  EOT
}


resource "null_resource" "fetch_image" {
  provisioner "local-exec" {
    command = local.fetch_image_command

    environment = {
      attestation_variant = var.attestation_variant
    }
  }
  provisioner "local-exec" {
    when    = destroy
    command = "rm image.txt"
  }
  triggers = {
    always_run = "${timestamp()}"
  }
}

data "local_file" "image" {
  filename   = "image.txt"
  depends_on = [null_resource.fetch_image]
}
