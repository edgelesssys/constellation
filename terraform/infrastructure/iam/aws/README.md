# IAM instance profiles for AWS

This terraform script creates the necessary profiles that need to be attached to Constellation nodes.

You can create the profiles with the following commands:

```sh
mkdir constellation_aws_iam
cd constellation_aws_iam
curl --remote-name-all https://raw.githubusercontent.com/edgelesssys/constellation/main/hack/terraform/aws/iam/{main,output,variables}.tf
terraform init
terraform apply -auto-approve -var name_prefix=my_constellation
```

You can either get the profile names from the Terraform output values `control_plane_instance_profile` and `worker_nodes_instance_profile` and manually add them to your Constellation configuration file.

Or you can do this with a `yq` command:

```sh
yq -i "
  .provider.aws.iamProfileControlPlane = $(terraform output control_plane_instance_profile) |
  .provider.aws.iamProfileWorkerNodes = $(terraform output worker_nodes_instance_profile)
  " path/to/constellation-conf.yaml
```
