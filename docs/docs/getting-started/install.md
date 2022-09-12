# Installation and Setup

Constellation runs entirely in your cloud environment and can be easily controlled via a dedicated Command Line Interface (CLI).

The following guides you through the steps of installing the CLI on your machine, verifying it, and connecting it to your Cloud Service Provider (CSP).

### Prerequisites

Make sure the following requirements are met:

- Your machine is running Linux or macOS
- You have admin rights on your machine
- [kubectl](https://kubernetes.io/docs/tasks/tools/) is installed
- Your cloud provider is Microsoft Azure or Google Cloud Platform (GCP)

## Install the Constellation CLI

Download the CLI executable from the [release page](https://github.com/edgelesssys/constellation/releases). Move the downloaded file to a directory in your `PATH` (default: `/usr/local/bin`) and make it executable by entering `chmod u+x constellation` in your terminal. 

:::note
Edgeless Systems uses [sigstore](https://www.sigstore.dev/) to sign each release of the CLI. You may want to [verify the signature](../workflows/verify-cli.md) before you use the CLI.
:::

:::tip
The CLI supports autocompletion for various shells. To set it up, run `constellation completion` and follow the given steps.
:::

## Set up cloud credentials

The CLI makes authenticated calls to the CSP API. Therefore, you need to set up Constellation with the credentials for your CSP. Currently, Microsoft Azure and Google Cloud Platform (GCP) are the only supported CSPs.

### Step 1: Authenticate

First, you need to authenticate with your CSP. The following lists the required steps for *testing* and *production* environments.

:::danger
The steps for a *testing* environment are simpler. However, they may expose secrets to the CSP. If in doubt, follow the *production* steps.
:::

<tabs groupId="csp">
<tabItem value="azure" label="Azure" default>

**Testing**

Simply open the [Azure Cloud Shell](https://docs.microsoft.com/en-us/azure/cloud-shell/overview).

**Production**

Use the latest version of the [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/) on a trusted machine:

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

If you are running from within VM on GCP, and the VM is allowed to access the necessary APIs, no further configuration is needed.

If you are using the [Google Cloud Shell](https://cloud.google.com/shell), make sure your [session is authorized](https://cloud.google.com/shell/docs/auth). For example, execute `gsutil` and accept the authorization prompt.

**Production**

For production, use one of the following options on a trusted machine:

- Use the [`gcloud` CLI](https://cloud.google.com/sdk/gcloud)

    ```bash
    gcloud auth application-default login
    ```

    This will ask you to log-in to your Google account and create your credentials.
    The Constellation CLI will automatically load these credentials when needed.

- Set up a service account and pass the credentials manually

    Follow [Google's guide](https://cloud.google.com/docs/authentication/production#manually) for setting up your credentials.

</tabItem>
</tabs>

### Step 2: Set permissions

Finally, set the required permissions for your user account.

<tabs groupId="csp">
<tabItem value="azure" label="Azure" default>

Set the following permissions:

- `Contributor`
- `User Access Administrator`

</tabItem>
<tabItem value="gcp" label="GCP" default>

Set the following permissions:

- `compute.*` (or the subset defined by `roles/compute.instanceAdmin.v1`)
- `iam.serviceAccountUser`

Follow Google's guide on [understanding](https://cloud.google.com/iam/docs/understanding-roles) and [assigning roles](https://cloud.google.com/iam/docs/granting-changing-revoking-access).

</tabItem>
</tabs>

### Next Steps

You are now ready to [deploy your first confidential Kubernetes cluster and application](first-steps.md).
