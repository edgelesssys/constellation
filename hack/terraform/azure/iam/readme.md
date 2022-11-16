# Terraform Azure IAM creation

This terraform configuration creates the necessary Azure resources that need to be available to host a Constellation cluster.

You can create the resources with the following commands:
```sh
mkdir constellation_azure_iam
cd constellation_azure_iam
curl --remote-name-all https://raw.githubusercontent.com/edgelesssys/constellation/main/hack/terraform/azure/iam/{main.tf,output.tf,variables.tf,.terraform.lock.hcl}
terraform init
terraform apply
```

The following terraform output values are available (with their corresponding keys in the Constellation configuration file):
- `subscription_id` (subscription)
- `tenant_id` (tenant)
- `region` (location)
- `base_resource_group_name` (resourceGroup)
- `application_id` (appClientID)
- `uami_id` (userAssignedIdentity)
- `application_client_secret_value` (clientSecretValue) - **Sensitive Value**

You can either get the profile names from the Terraform output and manually add them to your Constellation configuration file according to our [Documentation](https://docs.edgeless.systems/constellation/getting-started/first-steps).
Or you can do this with a `yq` command:
```sh
yq -i "
  .provider.azure.subscription = $(terraform output subscription_id) |
  .provider.azure.tenant = $(terraform output tenant_id) |
  .provider.azure.location = $(terraform output region) |
  .provider.azure.resourceGroup = $(terraform output base_resource_group_name) |
  .provider.azure.appClientID = $(terraform output application_id) |
  .provider.azure.userAssignedIdentity = $(terraform output uami_id) |
  .provider.azure.clientSecretValue = $(terraform output application_client_secret_value)
  " path/to/constellation-conf.yaml
```

Where `path/to/constellation-conf.yaml` is the path to your Constellation configuration file.
