provider_installation {

  dev_overrides {
    # The substitution is made in terraform-provider-devbuild
    "registry.terraform.io/edgelesssys/constellation" = "@@TERRAFORM_PROVIDER_PATH@@"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
