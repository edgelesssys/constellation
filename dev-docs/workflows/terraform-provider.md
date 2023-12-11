# Constellation Terraform Provider

This document explains the basic ways of working with the [Constellation Terraform Provider](../../terraform-provider-constellation/).

## Building the Terraform Provider

The Constellation Terraform provider can be built through Bazel, either via the [`devbuild` target](./build-develop-deploy.md) (recommended), which will create a `terraform` directory
with the provider binary and some utility files in the current working directory, or explicitly via this command:

```bash
bazel build //terraform-provider-constellation:tf_provider
```

Documentation for the provider can be generated with:

```bash
bazel run //:generate
# or
bazel run //bazel/ci:terraform_docgen
```

## Using the Terraform Provider

The Terraform provider binary can be used with the normal Terraform CLI, by setting a [development override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers),
so that the registry path to the provider is replaced with the path to the locally built provider. If using the [`devbuild` target](./build-develop-deploy.md), a `config.tfrc` file with the override set to the path
of the built binary is placed automatically in the `terraform` directory in the current working directory. Otherwise, the file can be also built and copied to the current working directory explicitly via this command:

```bash
bazel build //terraform-provider-constellation:terraform_rc
cp bazel-bin/terraform-provider-constellation/config.tfrc .
sed -i "s|@@TERRAFORM_PROVIDER_PATH@@|$(realpath bazel-bin/terraform-provider-constellation/tf_provider_/tf_provider)|g" config.tfrc
```

Afterwards, all Terraform commands that should use the local provider build should be prefixed with `TF_CLI_CONFIG_FILE=config.tfrc` like so:

```bash
TF_CLI_CONFIG_FILE=config.tfrc terraform apply
...
```

## Testing the Terraform Provider

Terraform acceptance tests can be run hermetically through Bazel (recommended):

```bash
bazel test --test_tag_filters=integration //terraform-provider-constellation/internal/provider:provider_test
```

The tests can also be run through Go, but the `TF_ACC` environment variable needs to be set to `1`, and the host's Terraform binary is used, which may produce inaccurate test results.

```bash
cd terraform-provider-constellation
TF_ACC=1 go test -v ./...
```
