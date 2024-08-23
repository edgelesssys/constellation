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

- `Microsoft.Attestation`
- `Microsoft.Compute`
- `Microsoft.Insights`
- `Microsoft.ManagedIdentity`
- `Microsoft.Network`

By default, Constellation tries to register these automatically if they haven't been registered before.

To [create the IAM configuration](../workflows/config.md#creating-an-iam-configuration) for Constellation, you need the following permissions:

- `*/register/action` \[1]
- `Microsoft.Authorization/roleAssignments/*`
- `Microsoft.Authorization/roleDefinitions/*`
- `Microsoft.ManagedIdentity/userAssignedIdentities/*`
- `Microsoft.Resources/subscriptions/resourcegroups/*`

The built-in `Owner` role is a superset of these permissions.

To [create a Constellation cluster](../workflows/create.md#the-create-step), you need the following permissions:

- `Microsoft.Attestation/attestationProviders/*`
- `Microsoft.Compute/virtualMachineScaleSets/*`
- `Microsoft.Insights/components/*`
- `Microsoft.ManagedIdentity/userAssignedIdentities/*`
- `Microsoft.Network/loadBalancers/*`
- `Microsoft.Network/loadBalancers/backendAddressPools/*`
- `Microsoft.Network/networkSecurityGroups/*`
- `Microsoft.Network/publicIPAddresses/*`
- `Microsoft.Network/virtualNetworks/*`
- `Microsoft.Network/virtualNetworks/subnets/*`

The built-in `Contributor` role is a superset of these permissions.

Follow Microsoft's guide on [understanding](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-definitions) and [assigning roles](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments).

1: You can omit `*/register/Action` if the resource providers mentioned above are already registered and the `ARM_SKIP_PROVIDER_REGISTRATION` environment variable is set to `true` when creating the IAM configuration.

</TabItem>
<TabItem value="gcp" label="GCP">

Create a new project for Constellation or use an existing one.
Enable the [Compute Engine API](https://console.cloud.google.com/apis/library/compute.googleapis.com) on it.

To [create the IAM configuration](../workflows/config.md#creating-an-iam-configuration) for Constellation, you need the following permissions:

- `iam.serviceAccountKeys.create`
- `iam.serviceAccountKeys.delete`
- `iam.serviceAccountKeys.get`
- `iam.serviceAccounts.create`
- `iam.serviceAccounts.delete`
- `iam.serviceAccounts.get`
- `resourcemanager.projects.getIamPolicy`
- `resourcemanager.projects.setIamPolicy`

Together, the built-in roles `roles/editor` and `roles/resourcemanager.projectIamAdmin` form a superset of these permissions.

To [create a Constellation cluster](../workflows/create.md#the-create-step), you need the following permissions:

- `compute.addresses.createInternal`
- `compute.addresses.deleteInternal`
- `compute.addresses.get`
- `compute.addresses.useInternal`
- `compute.backendServices.create`
- `compute.backendServices.delete`
- `compute.backendServices.get`
- `compute.backendServices.use`
- `compute.disks.create`
- `compute.firewalls.create`
- `compute.firewalls.delete`
- `compute.firewalls.get`
- `compute.globalAddresses.create`
- `compute.globalAddresses.delete`
- `compute.globalAddresses.get`
- `compute.globalAddresses.use`
- `compute.globalForwardingRules.create`
- `compute.globalForwardingRules.delete`
- `compute.globalForwardingRules.get`
- `compute.globalForwardingRules.setLabels`
- `compute.globalOperations.get`
- `compute.healthChecks.create`
- `compute.healthChecks.delete`
- `compute.healthChecks.get`
- `compute.healthChecks.useReadOnly`
- `compute.instanceGroupManagers.create`
- `compute.instanceGroupManagers.delete`
- `compute.instanceGroupManagers.get`
- `compute.instanceGroups.create`
- `compute.instanceGroups.delete`
- `compute.instanceGroups.get`
- `compute.instanceGroups.use`
- `compute.instances.create`
- `compute.instances.setLabels`
- `compute.instances.setMetadata`
- `compute.instances.setTags`
- `compute.instanceTemplates.create`
- `compute.instanceTemplates.delete`
- `compute.instanceTemplates.get`
- `compute.instanceTemplates.useReadOnly`
- `compute.networks.create`
- `compute.networks.delete`
- `compute.networks.get`
- `compute.networks.updatePolicy`
- `compute.routers.create`
- `compute.routers.delete`
- `compute.routers.get`
- `compute.routers.update`
- `compute.subnetworks.create`
- `compute.subnetworks.delete`
- `compute.subnetworks.get`
- `compute.subnetworks.use`
- `compute.targetTcpProxies.create`
- `compute.targetTcpProxies.delete`
- `compute.targetTcpProxies.get`
- `compute.targetTcpProxies.use`
- `iam.serviceAccounts.actAs`

Together, the built-in roles `roles/editor`, `roles/compute.instanceAdmin` and `roles/resourcemanager.projectIamAdmin` form a superset of these permissions.

Follow Google's guide on [understanding](https://cloud.google.com/iam/docs/understanding-roles) and [assigning roles](https://cloud.google.com/iam/docs/granting-changing-revoking-access).

</TabItem>
<TabItem value="aws" label="AWS">

To set up a Constellation cluster, you need to perform two tasks that require permissions: create the infrastructure and create roles for cluster nodes. Both of these actions can be performed by different users, e.g., an administrator to create roles and a DevOps engineer to create the infrastructure.

To [create the IAM configuration](../workflows/config.md#creating-an-iam-configuration) for Constellation, you need the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeAccountAttributes",
                "iam:AddRoleToInstanceProfile",
                "iam:AttachRolePolicy",
                "iam:CreateInstanceProfile",
                "iam:CreatePolicy",
                "iam:CreateRole",
                "iam:DeleteInstanceProfile",
                "iam:DeletePolicy",
                "iam:DeletePolicyVersion",
                "iam:DeleteRole",
                "iam:DetachRolePolicy",
                "iam:GetInstanceProfile",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:GetRole",
                "iam:ListAttachedRolePolicies",
                "iam:ListInstanceProfilesForRole",
                "iam:ListPolicyVersions",
                "iam:ListRolePolicies",
                "iam:PassRole",
                "iam:RemoveRoleFromInstanceProfile",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

The built-in `AdministratorAccess` policy is a superset of these permissions.

To [create a Constellation cluster](../workflows/create.md#the-create-step), you need the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "autoscaling:CreateAutoScalingGroup",
                "autoscaling:DeleteAutoScalingGroup",
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeScalingActivities",
                "autoscaling:SetInstanceProtection",
                "autoscaling:UpdateAutoScalingGroup",
                "ec2:AllocateAddress",
                "ec2:AssociateRouteTable",
                "ec2:AttachInternetGateway",
                "ec2:AuthorizeSecurityGroupEgress",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:CreateInternetGateway",
                "ec2:CreateLaunchTemplate",
                "ec2:CreateNatGateway",
                "ec2:CreateRoute",
                "ec2:CreateRouteTable",
                "ec2:CreateSecurityGroup",
                "ec2:CreateSubnet",
                "ec2:CreateTags",
                "ec2:CreateVpc",
                "ec2:DeleteInternetGateway",
                "ec2:DeleteLaunchTemplate",
                "ec2:DeleteNatGateway",
                "ec2:DeleteRouteTable",
                "ec2:DeleteSecurityGroup",
                "ec2:DeleteSubnet",
                "ec2:DeleteVpc",
                "ec2:DescribeAccountAttributes",
                "ec2:DescribeAddresses",
                "ec2:DescribeAvailabilityZones",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeLaunchTemplates",
                "ec2:DescribeLaunchTemplateVersions",
                "ec2:DescribeNatGateways",
                "ec2:DescribeNetworkAcls",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeRouteTables",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcAttribute",
                "ec2:DescribeVpcClassicLink",
                "ec2:DescribeVpcClassicLinkDnsSupport",
                "ec2:DescribeVpcs",
                "ec2:DetachInternetGateway",
                "ec2:DisassociateAddress",
                "ec2:DisassociateRouteTable",
                "ec2:ReleaseAddress",
                "ec2:RevokeSecurityGroupEgress",
                "elasticloadbalancing:AddTags",
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:CreateTargetGroup",
                "elasticloadbalancing:DeleteListener",
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:DeleteTargetGroup",
                "elasticloadbalancing:DescribeListeners",
                "elasticloadbalancing:DescribeLoadBalancerAttributes",
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:DescribeTags",
                "elasticloadbalancing:DescribeTargetGroupAttributes",
                "elasticloadbalancing:DescribeTargetGroups",
                "elasticloadbalancing:DescribeTargetHealth",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:ModifyTargetGroupAttributes",
                "iam:PassRole",
                "logs:CreateLogGroup",
                "logs:DeleteLogGroup",
                "logs:DescribeLogGroups",
                "logs:ListTagsLogGroup",
                "logs:PutRetentionPolicy",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

The built-in `PowerUserAccess` policy is a superset of these permissions.

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
