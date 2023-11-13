# Create your cluster

:::info
This recording presents the essence of this page. It's recommended to read it in full for the motivation and all details.
:::

<asciinemaWidget src="/constellation/assets/create-cluster.cast" rows="20" cols="112" idleTimeLimit="3" preload="true" theme="edgeless" />

---

Creating your cluster requires two steps:

1. Creating the necessary resources in your cloud environment
2. Bootstrapping the Constellation cluster and setting up a connection

See the [architecture](../architecture/orchestration.md) section for details on the inner workings of this process.

:::tip
If you don't have a cloud subscription, you can also set up a [local Constellation cluster using virtualization](../getting-started/first-steps-local.md) for testing.
:::

## The *create* step

This step creates the necessary resources for your cluster in your cloud environment.
Before you create the cluster, make sure to have a [valid configuration file](./config.md).

### Create

<tabs groupId="usage">
<tabItem value="cli" label="CLI">

```bash
constellation create
```

*create* stores your cluster's state in a [`constellation-terraform`](../architecture/orchestration.md#cluster-creation-process) directory in your workspace.

</tabItem>
<tabItem value="terraform" label="Terraform">

Terraform allows for an easier GitOps integration as well as meeting regulatory requirements.
Since the Constellation CLI also uses Terraform under the hood, you can reuse the same Terraform files.

:::info
Familiarize with the [Terraform usage policy](../reference/terraform.md) before manually interacting with Terraform to create a cluster.
Please also refrain from changing the Terraform resource definitions, as Constellation is tightly coupled to them.
:::

Download the Terraform files for the selected CSP from the [GitHub repository](https://github.com/edgelesssys/constellation/tree/main/terraform/infrastructure).

Find the image reference for your CSP and region, execute:

```bash
CONSTELL_VER=vX.Y.Z
curl -s https://cdn.confidential.cloud/constellation/v2/ref/-/stream/stable/$CONSTELL_VER/image/info.json | jq
```

From the list, select the `reference` for your CSP / Attestation combination and save it in the `IMAGE_REF` environment variable.

Create a `terraform.tfvars` file.
There, define all needed variables found in `variables.tf` using the values from the `constellation-config.yaml`.

<tabs groupId="csp">
<tabItem value="aws" label="AWS">

```bash
echo "name = \"$(yq '.name' constellation-conf.yaml)\"" >> terraform.tfvars
echo "debug = $(yq '.debugCluster' constellation-conf.yaml)" >> terraform.tfvars
echo "custom_endpoint = \"$(yq '.customEndpoint' constellation-conf.yaml)\"" >> terraform.tfvars
echo "node_groups = {
    control_plane_default = {
    role = \"$(yq '.nodeGroups.control_plane_default.role' constellation-conf.yaml)\"
    zone = \"$(yq '.nodeGroups.control_plane_default.zone' constellation-conf.yaml)\"
    instance_type = \"$(yq '.nodeGroups.control_plane_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.control_plane_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.control_plane_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.control_plane_default.initialCount' constellation-conf.yaml)\"
    }
    worker_default = {
    role = \"$(yq '.nodeGroups.worker_default.role' constellation-conf.yaml)\"
    zone = \"$(yq '.nodeGroups.worker_default.zone' constellation-conf.yaml)\"
    instance_type = \"$(yq '.nodeGroups.worker_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.worker_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.worker_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.worker_default.initialCount' constellation-conf.yaml)\"
    }
}" >> terraform.tfvars
echo "iam_instance_profile_control_plane = \"$(yq '.provider.aws.iamProfileControlPlane' constellation-conf.yaml)\"" >> terraform.tfvars
echo "iam_instance_profile_worker_nodes = \"$(yq '.provider.aws.iamProfileWorkerNodes' constellation-conf.yaml)\"" >> terraform.tfvars
echo "region = \"$(yq '.provider.aws.region' constellation-conf.yaml)\"" >> terraform.tfvars
echo "zone = \"$(yq '.provider.aws.zone' constellation-conf.yaml)\"" >> terraform.tfvars
echo "ami = \"$(yq '.provider.aws.zone' constellation-conf.yaml)\"" >> terraform.tfvars
echo "enable_snp = $(yq '.attestation | has("awsSEVSNP")' constellation-conf.yaml)" >> terraform.tfvars
terraform fmt terraform.tfvars
```

</tabItem>
<tabItem value="azure" label="Azure">

```bash
echo "name = \"$(yq '.name' constellation-conf.yaml)\"" >> terraform.tfvars
echo "debug = $(yq '.debugCluster' constellation-conf.yaml)" >> terraform.tfvars
echo "custom_endpoint = \"$(yq '.customEndpoint' constellation-conf.yaml)\"" >> terraform.tfvars
echo "image_id = \"$IMAGE_REF\"" >> terraform.tfvars
echo "node_groups = {
    control_plane_default = {
    role = \"$(yq '.nodeGroups.control_plane_default.role' constellation-conf.yaml)\"
    zones = [ \"$(yq '.nodeGroups.worker_default.zone' constellation-conf.yaml)\" ]
    instance_type = \"$(yq '.nodeGroups.control_plane_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.control_plane_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.control_plane_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.control_plane_default.initialCount' constellation-conf.yaml)\"
    }
    worker_default = {
    role = \"$(yq '.nodeGroups.worker_default.role' constellation-conf.yaml)\"
    zones = [ \"$(yq '.nodeGroups.worker_default.zone' constellation-conf.yaml)\" ]
    instance_type = \"$(yq '.nodeGroups.worker_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.worker_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.worker_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.worker_default.initialCount' constellation-conf.yaml)\"
    }
}" >> terraform.tfvars
echo "location = \"$(yq '.provider.azure.location' constellation-conf.yaml)\"" >> terraform.tfvars
echo "create_maa = $(yq '.attestation | has("azureSEVSNP")' constellation-conf.yaml)" >> terraform.tfvars
echo "confidential_vm = $(yq '.attestation | has("azureSEVSNP")' constellation-conf.yaml)" >> terraform.tfvars
echo "secure_boot = $(yq '.provider.azure.secureBoot' constellation-conf.yaml)" >> terraform.tfvars
echo "resource_group = \"$(yq '.provider.azure.resourceGroup' constellation-conf.yaml)\"" >> terraform.tfvars
echo "user_assigned_identity = \"$(yq '.provider.azure.userAssignedIdentity' constellation-conf.yaml)\"" >> terraform.tfvars
terraform fmt terraform.tfvars
```

</tabItem>
<tabItem value="gcp" label="GCP">

```bash
echo "name = \"$(yq '.name' constellation-conf.yaml)\"" >> terraform.tfvars
echo "debug = $(yq '.debugCluster' constellation-conf.yaml)" >> terraform.tfvars
echo "custom_endpoint = \"$(yq '.customEndpoint' constellation-conf.yaml)\"" >> terraform.tfvars
echo "image_id = \"$IMAGE_REF\"" >> terraform.tfvars
echo "node_groups = {
    control_plane_default = {
    role = \"$(yq '.nodeGroups.control_plane_default.role' constellation-conf.yaml)\"
    zone = \"$(yq '.nodeGroups.control_plane_default.zone' constellation-conf.yaml)\"
    instance_type = \"$(yq '.nodeGroups.control_plane_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.control_plane_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.control_plane_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.control_plane_default.initialCount' constellation-conf.yaml)\"
    }
    worker_default = {
    role = \"$(yq '.nodeGroups.worker_default.role' constellation-conf.yaml)\"
    zone = \"$(yq '.nodeGroups.worker_default.zone' constellation-conf.yaml)\"
    instance_type = \"$(yq '.nodeGroups.worker_default.instanceType' constellation-conf.yaml)\"
    disk_size = \"$(yq '.nodeGroups.worker_default.stateDiskSizeGB' constellation-conf.yaml)\"
    disk_type = \"$(yq '.nodeGroups.worker_default.stateDiskType' constellation-conf.yaml)\"
    initial_count = \"$(yq '.nodeGroups.worker_default.initialCount' constellation-conf.yaml)\"
    }
}" >> terraform.tfvars
echo "project = \"$(yq '.provider.gcp.project' constellation-conf.yaml)\"" >> terraform.tfvars
echo "region = \"$(yq '.provider.gcp.region' constellation-conf.yaml)\"" >> terraform.tfvars
echo "zone = \"$(yq '.provider.gcp.zone' constellation-conf.yaml)\"" >> terraform.tfvars
terraform fmt terraform.tfvars
```

</tabItem>
</tabs>

Initialize and apply Terraform to create the configured infrastructure:

```bash
terraform init
terraform apply
```

The Constellation [apply step](#the-apply-step) requires the already created `constellation-config.yaml` and the `constellation-state.yaml`.
Create the `constellation-state.yaml` using the output from the Terraform state and the `constellation-conf.yaml`:

<tabs groupId="csp">
<tabItem value="aws" label="AWS">

```bash
yq eval '.version ="v1"' --inplace constellation-state.yaml
yq eval ".infrastructure.initSecret =\"$(terraform output initSecret | jq -r | tr -d '\n' | hexdump -ve '/1 "%02x"' && echo '')\"" --inplace constellation-state.yaml
yq eval ".infrastructure.clusterEndpoint =\"$(terraform output out_of_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.inClusterEndpoint =\"$(terraform output in_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.ipCidrNode =\"$(terraform output ip_cidr_nodes | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.uid =\"$(terraform output uid | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.name =\"$(terraform output name | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.apiServerCertSANs =$(terraform output -json api_server_cert_sans)" --inplace constellation-state.yaml
```

</tabItem>
<tabItem value="azure" label="Azure">

:::info

  If the enforcement policy is set to `MAAFallback` in `constellation-config.yaml`, a manual update to the MAA provider's policy is necessary.
  You can apply the update with the following commands, where `<VERSION>` is the version of Constellation that should be set up. (e.g. `v2.12.0`)

  ```bash
  git clone --branch <VERSION> https://github.com/edgelesssys/constellation
  cd constellation/hack/maa-patch
  go run . $(terraform output attestationURL | jq -r)
  ```

:::

```bash
yq eval '.version ="v1"' --inplace constellation-state.yaml
yq eval ".infrastructure.initSecret =\"$(terraform output initSecret | jq -r | tr -d '\n' | hexdump -ve '/1 "%02x"' && echo '')\"" --inplace constellation-state.yaml
yq eval ".infrastructure.clusterEndpoint =\"$(terraform output out_of_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.inClusterEndpoint =\"$(terraform output in_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.ipCidrNode =\"$(terraform output ip_cidr_nodes | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.uid =\"$(terraform output uid | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.name =\"$(terraform output name | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.apiServerCertSANs =$(terraform output -json api_server_cert_sans)" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.resourceGroup =\"$(terraform output resource_group | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.subscriptionID =\"$(terraform output subscription_id | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.networkSecurityGroupName =\"$(terraform output network_security_group_name | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.loadBalancerName =\"$(terraform output loadbalancer_name | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.userAssignedIdentity =\"$(terraform output user_assigned_identity_client_id | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.azure.attestationURL =\"$(terraform output attestationURL | jq -r)\"" --inplace constellation-state.yaml
```

</tabItem>
<tabItem value="gcp" label="GCP">

```bash
yq eval '.version ="v1"' --inplace constellation-state.yaml
yq eval ".infrastructure.initSecret =\"$(terraform output initSecret | jq -r | tr -d '\n' | hexdump -ve '/1 "%02x"' && echo '')\"" --inplace constellation-state.yaml
yq eval ".infrastructure.clusterEndpoint =\"$(terraform output out_of_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.inClusterEndpoint =\"$(terraform output in_cluster_endpoint | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.ipCidrNode =\"$(terraform output ip_cidr_nodes | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.uid =\"$(terraform output uid | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.name =\"$(terraform output name | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.apiServerCertSANs =$(terraform output -json api_server_cert_sans)" --inplace constellation-state.yaml
yq eval ".infrastructure.gcp.projectID =\"$(terraform output project | jq -r)\"" --inplace constellation-state.yaml
yq eval ".infrastructure.gcp.ipCidrPod =\"$(terraform output ip_cidr_pods | jq -r)\"" --inplace constellation-state.yaml
```

</tabItem>
</tabs>
</tabItem>
<tabItem value="self-managed" label="Self-managed">

Self-managed infrastructure allows for managing the cloud resources necessary for a Constellation cluster separate from the Constellation CLI.
This provides flexibility in DevOps and can meet potential regulatory requirements.

To self-manage the infrastructure of your cluster, download the Terraform files for the selected CSP from the [Constellation GitHub repository](https://github.com/edgelesssys/constellation/tree/main/terraform/infrastructure).
They contain a minimum configuration for the resources necessary to run a Constellation cluster on the corresponding CSP. From this base, you can now add, edit, or substitute resources per your own requirements with the infrastructure
management tooling of your choice. You need to keep the essential functionality of the base configuration in order for your cluster to function correctly.

:::info

  On Azure, if the enforcement policy is set to `MAAFallback` in `constellation-config.yaml`, a manual update to the MAA provider's policy is necessary.
  <!-- vale off -->
  You can apply the update with the following command after creating the infrastructure, with `<URL>` being the URL of the MAA provider (i.e., `$(terraform output attestationURL | jq -r)`, when using the minimal Terraform configuration).

  ```bash
  constellation maa-patch <URL>
  ```
   <!-- vale on -->
:::

Make sure all necessary resources are created, e.g., through checking your CSP's portal and retrieve the necessary values, aligned with the outputs (specified in `outputs.tf`) of the base configuration.

Fill these outputs into the corresponding fields of the `constellation-state.yaml` file. For example, fill the IP or DNS name your cluster can be reached at into the `.Infrastructure.ClusterEndpoint` field.

Continue with [initializing your cluster](#the-apply-step).

</tabItem>
</tabs>

## The *apply* step

The following command initializes and bootstraps your cluster:

```bash
constellation apply
```

Next, configure `kubectl` for your cluster:

```bash
export KUBECONFIG="$PWD/constellation-admin.conf"
```

🏁 That's it. You've successfully created a Constellation cluster.

### Troubleshooting

In case `apply` fails, the CLI collects logs from the bootstrapping instance and stores them inside `constellation-cluster.log`.
