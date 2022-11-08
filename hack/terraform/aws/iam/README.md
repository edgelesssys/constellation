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

## Development

### iamlive

[iamlive](https://github.com/iann0036/iamlive) dynamically determines the minimal
permissions to call a set of AWS API calls.

It uses a local proxy to intercept API calls and incrementally generate the AWS
policy.

In one session start `iamlive`:

```sh
iamlive -mode proxy -bind-addr 0.0.0.0:10080 -force-wildcard-resource -output-file iamlive.policy.json
```

In another session execute terraform:

```sh
PREFIX="record-iam"
terraform init
HTTP_PROXY=http://127.0.0.1:10080 HTTPS_PROXY=http://127.0.0.1:10080 AWS_CA_BUNDLE="${HOME}/.iamlive/ca.pem" terraform apply -auto-approve -var name_prefix=${PREFIX}
HTTP_PROXY=http://127.0.0.1:10080 HTTPS_PROXY=http://127.0.0.1:10080 AWS_CA_BUNDLE="${HOME}/.iamlive/ca.pem" terraform destroy -auto-approve -var name_prefix=${PREFIX}
```

`iamlive` will present the generated policy, and after \<CTRL-C\> the `iamlive` process it will also write it to the specified file.
