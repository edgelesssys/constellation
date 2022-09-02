# Installation and Setup

Constellation runs entirely in your cloud environment and can be easily controlled via a dedicated Command Line Interface (CLI).

The installation process will guide you through the steps of installing the CLI on your machine, verifying it, and connecting it to your Cloud Service Provider (CSP).

### Prerequisites

Before we start, make sure the following requirements are fulfilled:

- Your machine is running Ubuntu or macOS
- You have admin rights on your machine
- [kubectl](https://kubernetes.io/docs/tasks/tools/) is installed
- Your cloud provider is Microsoft Azure or Google Cloud

## Install the Constellation CLI

The Constellation CLI can be downloaded from our [release page](https://github.com/edgelesssys/constellation/releases). Therefore, navigate to a release and download the file `constellation`. Move the downloaded file to a directory in your `PATH` (default: `/usr/local/bin`) and make it executable by entering `chmod s+x constellation` in your terminal.

Running `constellation` should then give you:

```shell-session
$ constellation
Manage your Constellation cluster.

Usage:
  constellation [command]

...
```

### Optional: Enable shell autocompletion

The Constellation CLI supports autocompletion for various shells. To set it up, run `constellation completion` and follow the steps.

## Verify the CLI

For extra security, make sure to verify your CLI. Therefore, install [cosign](https://github.com/sigstore/cosign). Then, head to our [release page](https://github.com/edgelesssys/constellation/releases) again and, from the same release as before, download the following files:

- `cosign.pub` (Edgeless System's cosign public key)
- `constellation.sig` (the CLI's signature)

You can then verify your CLI before launching a cluster using the paths to the public key, signature, and CLI executable:

```bash
cosign verify-blob --key cosign.pub --signature constellation.sig constellation
```

For more detailed information read our [installation, update and verification documentation](../architecture/orchestration.md).

## Set up cloud credentials

The CLI makes authenticated calls to the CSP API. Therefore, you need to set up Constellation with the credentials for your CSP.

### Authentication

In the following, we provide you with different ways to authenticate with your CSP.

:::danger

Don't use the testing methods for setting up a production-grade Constellation cluster. In those methods, secrets are stored on disk during installation which would be exposed to the CSP.

:::

<tabs>
<tabItem value="azure" label="Azure" default>

**Testing**

To try out Constellation, using a cloud environment such as [Azure Cloud Shell](https://docs.microsoft.com/en-us/azure/cloud-shell/overview) is the quickest way to get started.

**Production**

For production clusters, use the [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/) on a trusted machine:

```bash
az login
```

Other options are described in Azure's [authentication guide](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli).

</tabItem>
<tabItem value="gcp" label="GCP" default>

Enable the following cloud APIs first:

- [Compute Engine API](https://console.cloud.google.com/marketplace/product/google/compute.googleapis.com)
- [Cloud Resource Manager API](https://console.cloud.google.com/apis/library/cloudresourcemanager.googleapis.com)
- [Identity and Access Management (IAM) API](https://console.developers.google.com/apis/api/iam.googleapis.com)

**Testing**

- If you are running from within a Google VM, and the VM is allowed to access the necessary APIs, no further configuration is needed.

- If you are using the [Google Cloud Shell](https://cloud.google.com/shell), make sure your [session is authorized](https://cloud.google.com/shell/docs/auth). For example, execute `gsutil` and accept the authorization prompt.

**Production**

For production clusters, use one of the following options on a trusted machine:

- Use the [`gcloud` CLI](https://cloud.google.com/sdk/gcloud)

    ```bash
    gcloud auth application-default login
    ```

    This will ask you to log in to your Google account, and then create your credentials.
    The Constellation CLI will automatically load these credentials when needed.

- Set up a service account and pass the credentials manually

    Follow [Google's guide](https://cloud.google.com/docs/authentication/production#manually) for setting up your credentials.

</tabItem>
</tabs>

### Authorization

<tabs>
<tabItem value="azure" label="Azure" default>

Your user account needs the following permissions to set up a Constellation cluster:

- `Contributor`
- `User Access Administrator`

Additionally, you need to [create a user-assigned managed identity](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-manage-user-assigned-managed-identities) with the following roles:

- `Virtual Machine Contributor`
- `Application Insights Component Contributor`

The user-assigned identity is used by the instances of the cluster to access other cloud resources.

You also need an empty resource group per cluster. Notice that the user-assigned identity has to be created in a
different resource group as all resources within the cluster resource group will be deleted on cluster termination.

Last, you need to [create an Active Directory app registration](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app#register-an-application) (you don't need to add a redirect URI).
As supported account types choose 'Accounts in this organizational directory only'. Then [create a client secret](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal#option-2-create-a-new-application-secret), which will be used by Kubernetes.
On the cluster resource group, [add the app registration](https://docs.microsoft.com/en-us/azure/role-based-access-control/role-assignments-portal?tabs=current#step-2-open-the-add-role-assignment-page)
with role `Owner`.

User-assigned identity, cluster resource group, app registration client ID and client secret value need to be set in the `constellation-conf.yaml` configuration file.

</tabItem>
<tabItem value="gcp" label="GCP" default>

Your user account needs the following permissions to set up a Constellation:

- `compute.*` (or the subset defined by `roles/compute.instanceAdmin.v1`)

Follow Google's guide on [understanding](https://cloud.google.com/iam/docs/understanding-roles) and [assigning roles](https://cloud.google.com/iam/docs/granting-changing-revoking-access).

Additionally, you need a service account with the following permissions:

- `Compute Instance Admin (v1) (roles/compute.instanceAdmin.v1)`
- `Compute Network Admin (roles/compute.networkAdmin)`
- `Compute Security Admin (roles/compute.securityAdmin)`
- `Compute Storage Admin (roles/compute.storageAdmin)`
- `Service Account User (roles/iam.serviceAccountUser)`

The key for this service account is passed to the CLI and used by Kubernetes to authenticate with the cloud.
You can configure the path to the key in the `constellation-conf.yaml` configuration file.

GCP instances come with a [default service account](https://cloud.google.com/iam/docs/service-accounts#default) attached
that's used by the instances to access the cloud resources of the cluster. You don't need to configure it.

</tabItem>
</tabs>

### Troubleshooting

If you receive an error during any of the outlined steps, please verify that you have executed all previous steps in this documentation. Also, feel free to ask our active community on [Discord](https://discord.com/invite/rH8QTH56JN) for help.

### Next Steps

Once you have followed all previous steps, you can proceed [to deploy your first confidential Kubernetes cluster and application](first-steps.md).
