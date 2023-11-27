# Terraform development
## Lock file generation

Lock files are only checked in for modules where the provider is explicitly used.
For modules that only consume other modules no lock file is provided to avoid duplication.

## iamlive

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
