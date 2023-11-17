# IAM configuration for GCP

This terraform script creates the necessary GCP IAM configuration to be attached to Constellation nodes.

You can create the configuration with the following commands:

```sh
mkdir constellation_gcp_iam
cd constellation_gcp_iam
curl --remote-name-all https://raw.githubusercontent.com/edgelesssys/constellation/main/terraform/infrastructure/iam/gcp/{main.tf,outputs.tf,variables.tf,.terraform.lock.hcl}
terraform init
terraform apply
```

The following terraform output values are available (with their corresponding keys in the Constellation configuration file):

- `sa_key` - **Sensitive Value**
- `region` (region)
- `zone` (zone)
- `project_id` (project)

You can either get the values from the Terraform output and manually add them to your Constellation configuration file according to our [Documentation](https://docs.edgeless.systems/constellation/getting-started/first-steps). (If you add the values manually, you need to base64-decode the `sa_key` value and place it in a JSON file, then specify the path to this file in the Constellation configuration file for the `serviceAccountKeyPath` key.)

Or you can setup the constellation configuration file automaticcaly with the following commands:

```sh
terraform output sa_key | sed "s/\"//g" | base64 --decode | tee gcpServiceAccountKey.json
yq -i "
  .provider.gcp.serviceAccountKeyPath = \"$(realpath gcpServiceAccountKey.json)\" |
  .provider.gcp.project = $(terraform output project_id) |
  .provider.gcp.region = $(terraform output region) |
  .provider.gcp.zone = $(terraform output zone)
  " path/to/constellation-conf.yaml
```

Where `path/to/constellation-conf.yaml` is the path to your Constellation configuration file.
