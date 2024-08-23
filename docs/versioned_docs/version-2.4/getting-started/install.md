# Installation and setup

Constellation runs entirely in your cloud environment and can be controlled via a dedicated command-line interface (CLI).

The following guides you through the steps of installing the CLI on your machine, verifying it, and connecting it to your cloud service provider (CSP).

## Prerequisites

Make sure the following requirements are met:

- Your machine is running Linux or macOS
- You have admin rights on your machine
- [kubectl](https://kubernetes.io/docs/tasks/tools/) is installed
- Your CSP is Amazon Web Services (AWS), Microsoft Azure, or Google Cloud Platform (GCP)

## Install the Constellation CLI

The CLI executable is available at [GitHub](https://github.com/edgelesssys/constellation/releases).
Install it with the following commands:

<Tabs>
<TabItem value="linux-amd64" label="Linux (amd64)">

1. Download the CLI:

```bash
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-amd64
```

2. [Verify the signature](../workflows/verify-cli.md) (optional)

3. Install the CLI to your PATH:

```bash
sudo install constellation-linux-amd64 /usr/local/bin/constellation
```

</TabItem>
<TabItem value="linux-arm64" label="Linux (arm64)">

1. Download the CLI:

```bash
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-linux-arm64
```

2. [Verify the signature](../workflows/verify-cli.md) (optional)

3. Install the CLI to your PATH:

```bash
sudo install constellation-linux-arm64 /usr/local/bin/constellation
```

</TabItem>

<TabItem value="darwin-arm64" label="macOS (Apple Silicon)">

1. Download the CLI:

```bash
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-darwin-arm64
```

2. [Verify the signature](../workflows/verify-cli.md) (optional)

3. Install the CLI to your PATH:

```bash
sudo install constellation-darwin-arm64 /usr/local/bin/constellation
```

</TabItem>

<TabItem value="darwin-amd64" label="macOS (Intel)">

1. Download the CLI:

```bash
curl -LO https://github.com/edgelesssys/constellation/releases/latest/download/constellation-darwin-amd64
```

2. [Verify the signature](../workflows/verify-cli.md) (optional)

3. Install the CLI to your PATH:

```bash
sudo install constellation-darwin-amd64 /usr/local/bin/constellation
```

</TabItem>
</Tabs>

:::tip
The CLI supports autocompletion for various shells. To set it up, run `constellation completion` and follow the given steps.
:::

## Set up cloud credentials

The CLI makes authenticated calls to the CSP API. Therefore, you need to set up Constellation with the credentials for your CSP.

:::tip
If you don't have a cloud subscription, you can try [MiniConstellation](first-steps-local.md), which lets you set up a local Constellation cluster using virtualization.
:::

### Required permissions

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

The following [resource providers need to be registered](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/resource-providers-and-types#register-resource-provider) in your subscription:

- `Microsoft.Compute`
- `Microsoft.ManagedIdentity`
- `Microsoft.Network`
- `microsoft.insights`

By default, Constellation tries to register these automatically if they haven't been registered before.

You need the following permissions for your user account:

- `Contributor` (to create cloud resources)
- `User Access Administrator` (to create a service account)

If you don't have these permissions with scope *subscription*, ask your administrator to [create the service account and a resource group for your Constellation cluster](first-steps.md).
Your user account needs the `Contributor` permission scoped to this resource group.

</TabItem>
<TabItem value="gcp" label="GCP">

Create a new project for Constellation or use an existing one.
Enable the [Compute Engine API](https://console.cloud.google.com/apis/library/compute.googleapis.com) on it.

You need the following permissions on this project:

- `compute.*` (or the subset defined by `roles/compute.instanceAdmin.v1`)
- `iam.serviceAccountUser`

Follow Google's guide on [understanding](https://cloud.google.com/iam/docs/understanding-roles) and [assigning roles](https://cloud.google.com/iam/docs/granting-changing-revoking-access).

</TabItem>
<TabItem value="aws" label="AWS">

To set up a Constellation cluster, you need to perform two tasks that require permissions: create the infrastructure and create roles for cluster nodes. Both of these actions can be performed by different users, e.g., an administrator to create roles and a DevOps engineer to create the infrastructure.

To create the AWS IAM policies, your user requires the following minimal set of permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sts:GetCallerIdentity",
                "ec2:DescribeAccountAttributes",
                "iam:CreateRole",
                "iam:CreatePolicy",
                "iam:GetPolicy",
                "iam:GetRole",
                "iam:GetPolicyVersion",
                "iam:ListRolePolicies",
                "iam:ListAttachedRolePolicies",
                "iam:CreateInstanceProfile",
                "iam:AttachRolePolicy",
                "iam:GetInstanceProfile",
                "iam:AddRoleToInstanceProfile",
                "iam:PassRole",
                "iam:RemoveRoleFromInstanceProfile",
                "iam:DetachRolePolicy",
                "iam:DeleteInstanceProfile",
                "iam:ListPolicyVersions",
                "iam:ListInstanceProfilesForRole",
                "iam:DeletePolicy",
                "iam:DeletePolicyVersion",
                "iam:DeleteRole"
            ],
            "Resource": "*"
        }
    ]
}
```

To create the infrastructure, you can either use a predefined role from Amazon,
such as `PowerUserAccess`, or use the following minimal set of permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sts:GetCallerIdentity",
                "ec2:DescribeAccountAttributes",
                "ec2:AllocateAddress",
                "ec2:CreateVpc",
                "ec2:CreateTags",
                "logs:CreateLogGroup",
                "ec2:CreateLaunchTemplate",
                "ec2:DescribeAddresses",
                "ec2:DescribeLaunchTemplates",
                "logs:PutRetentionPolicy",
                "logs:DescribeLogGroups",
                "ec2:DescribeVpcs",
                "ec2:DescribeLaunchTemplateVersions",
                "logs:ListTagsLogGroup",
                "ec2:DescribeVpcClassicLink",
                "ec2:DescribeVpcClassicLinkDnsSupport",
                "ec2:DescribeVpcAttribute",
                "ec2:DescribeNetworkAcls",
                "ec2:DescribeRouteTables",
                "ec2:DescribeSecurityGroups",
                "ec2:CreateSubnet",
                "ec2:CreateSecurityGroup",
                "elasticloadbalancing:CreateTargetGroup",
                "ec2:CreateInternetGateway",
                "ec2:DescribeSubnets",
                "elasticloadbalancing:DescribeTargetGroups",
                "ec2:AttachInternetGateway",
                "elasticloadbalancing:ModifyTargetGroupAttributes",
                "ec2:DescribeInternetGateways",
                "autoscaling:CreateAutoScalingGroup",
                "iam:PassRole",
                "ec2:CreateNatGateway",
                "ec2:RevokeSecurityGroupEgress",
                "elasticloadbalancing:DescribeTargetGroupAttributes",
                "elasticloadbalancing:CreateLoadBalancer",
                "ec2:DescribeNatGateways",
                "elasticloadbalancing:DescribeTags",
                "autoscaling:DescribeScalingActivities",
                "ec2:CreateRouteTable",
                "autoscaling:DescribeAutoScalingGroups",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:AuthorizeSecurityGroupEgress",
                "ec2:CreateRoute",
                "ec2:AssociateRouteTable",
                "elasticloadbalancing:DescribeTargetHealth",
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:AddTags",
                "elasticloadbalancing:DescribeLoadBalancerAttributes",
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:DescribeListeners",
                "logs:DeleteLogGroup",
                "elasticloadbalancing:DeleteListener",
                "ec2:DisassociateRouteTable",
                "autoscaling:UpdateAutoScalingGroup",
                "elasticloadbalancing:DeleteLoadBalancer",
                "autoscaling:SetInstanceProtection",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DeleteRouteTable",
                "ec2:DeleteNatGateway",
                "ec2:DetachInternetGateway",
                "ec2:DisassociateAddress",
                "ec2:ReleaseAddress",
                "ec2:DeleteInternetGateway",
                "ec2:DeleteSubnet",
                "autoscaling:DeleteAutoScalingGroup",
                "ec2:DeleteLaunchTemplate",
                "elasticloadbalancing:DeleteTargetGroup",
                "ec2:DeleteSecurityGroup",
                "ec2:DeleteVpc"
            ],
            "Resource": "*"
        }
    ]
}
```

Follow Amazon's guide on [understanding](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html) and [managing policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html).

</TabItem>
</Tabs>

### Authentication

You need to authenticate with your CSP. The following lists the required steps for *testing* and *production* environments.

:::note
The steps for a *testing* environment are simpler. However, they may expose secrets to the CSP. If in doubt, follow the *production* steps.
:::

<Tabs groupId="csp">
<TabItem value="azure" label="Azure">

**Testing**

Simply open the [Azure Cloud Shell](https://docs.microsoft.com/en-us/azure/cloud-shell/overview).

**Production**

Use the latest version of the [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/) on a trusted machine:

```bash
az login
```

Other options are described in Azure's [authentication guide](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli).

</TabItem>
<TabItem value="gcp" label="GCP">

**Testing**

You can use the [Google Cloud Shell](https://cloud.google.com/shell). Make sure your [session is authorized](https://cloud.google.com/shell/docs/auth). For example, execute `gsutil` and accept the authorization prompt.

**Production**

Use one of the following options on a trusted machine:

- Use the [`gcloud` CLI](https://cloud.google.com/sdk/gcloud)

    ```bash
    gcloud auth application-default login
    ```

    This will ask you to log-in to your Google account and create your credentials.
    The Constellation CLI will automatically load these credentials when needed.

- Set up a service account and pass the credentials manually

    Follow [Google's guide](https://cloud.google.com/docs/authentication/production#manually) for setting up your credentials.

</TabItem>
<TabItem value="aws" label="AWS">

**Testing**

You can use the [AWS CloudShell](https://console.aws.amazon.com/cloudshell/home). Make sure you are [authorized to use it](https://docs.aws.amazon.com/cloudshell/latest/userguide/sec-auth-with-identities.html).

**Production**

Use the latest version of the [AWS CLI](https://aws.amazon.com/cli/) on a trusted machine:

```bash
aws configure
```

Options and first steps are described in the [AWS CLI documentation](https://docs.aws.amazon.com/cli/index.html).

</TabItem>

</Tabs>

## Next steps

You are now ready to [deploy your first confidential Kubernetes cluster and application](first-steps.md).
