# Migrations

This document describes breaking changes and migrations between Constellation releases.
Use [`constellation config migrate`](./cli.md#constellation-config-migrate) to automatically update an old config file to a new format.


## Migrations to v2.19.0

### Azure

* To allow seamless upgrades on Azure when Kubernetes services of type `LoadBalancer` are deployed, the target  
 load balancer in which the `cloud-controller-manager` creates load balancing rules was changed. Instead of using the load balancer
 created and maintained by the CLI's Terraform code, the `cloud-controller-manager` now creates its own load balancer in Azure.
 If your Constellation has services of type `LoadBalancer`, please remove them before the upgrade and re-apply them
 afterward. 


## Migrating from Azure's service principal authentication to managed identity authentication (during the upgrade to Constellation v2.8.0)

- The `provider.azure.appClientID` and `provider.azure.appClientSecret` fields are no longer supported and should be removed.
- To keep using an existing UAMI, add the `Owner` permission with the scope of your `resourceGroup`.
- Otherwise, simply [create new Constellation IAM credentials](../workflows/config.md#creating-an-iam-configuration) and use the created UAMI.
- To migrate the authentication for an existing cluster on Azure to an UAMI with the necessary permissions:
  1. Remove the `aadClientId` and `aadClientSecret` from the azureconfig secret.
  2. Set `useManagedIdentityExtension` to `true`  and use the `userAssignedIdentity` from the Constellation config for the value of `userAssignedIdentityID`.
  3. Restart the CSI driver, cloud controller manager, cluster autoscaler, and Constellation operator pods.


## Migrating from CLI versions before 2.10

- AWS cluster upgrades require additional IAM permissions for the newly introduced `aws-load-balancer-controller`. Please upgrade your IAM roles using `iam upgrade apply`. This will show necessary changes and apply them, if desired.
- The global `nodeGroups` field was added.
- The fields `instanceType`, `stateDiskSizeGB`, and `stateDiskType` for each cloud provider are now part of the configuration of individual node groups.
- The `constellation create` command no longer uses the flags `--control-plane-count` and `--worker-count`. Instead, the initial node count is configured per node group in the `nodeGroups` field.

## Migrating from CLI versions before 2.9

- The `provider.azure.appClientID` and `provider.azure.clientSecretValue` fields were removed to enforce migration to managed identity authentication

## Migrating from CLI versions before 2.8

- The `measurements` field for each cloud service provider was replaced with a global `attestation` field.
- The `confidentialVM`, `idKeyDigest`, and `enforceIdKeyDigest` fields for the Azure cloud service provider were removed in favor of using the global `attestation` field.
- The optional global field `attestationVariant` was replaced by the now required `attestation` field.

## Migrating from CLI versions before 2.3

- The `sshUsers` field was deprecated in v2.2 and has been removed from the configuration in v2.3.
  As an alternative for SSH, check the workflow section [Connect to nodes](../workflows/troubleshooting.md#node-shell-access).
- The `image` field for each cloud service provider has been replaced with a global `image` field. Use the following mapping to migrate your configuration:
    <details>
    <summary>Show all</summary>

    | CSP   | old image                                                                                                                                                                             | new image |
    | ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
    | AWS   | `ami-06b8cbf4837a0a57c`                                                                                                                                                               | `v2.2.2`  |
    | AWS   | `ami-02e96dc04a9e438cd`                                                                                                                                                               | `v2.2.2`  |
    | AWS   | `ami-028ead928a9034b2f`                                                                                                                                                               | `v2.2.2`  |
    | AWS   | `ami-032ac10dd8d8266e3`                                                                                                                                                               | `v2.2.1`  |
    | AWS   | `ami-032e0d57cc4395088`                                                                                                                                                               | `v2.2.1`  |
    | AWS   | `ami-053c3e49e19b96bdd`                                                                                                                                                               | `v2.2.1`  |
    | AWS   | `ami-0e27ebcefc38f648b`                                                                                                                                                               | `v2.2.0`  |
    | AWS   | `ami-098cd37f66523b7c3`                                                                                                                                                               | `v2.2.0`  |
    | AWS   | `ami-04a87d302e2509aad`                                                                                                                                                               | `v2.2.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/2.2.2`     | `v2.2.2`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_CVM/images/constellation/versions/2.2.2` | `v2.2.2`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/2.2.1`     | `v2.2.1`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_CVM/images/constellation/versions/2.2.1` | `v2.2.1`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/2.2.0`     | `v2.2.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_CVM/images/constellation/versions/2.2.0` | `v2.2.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/2.1.0`     | `v2.1.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_CVM/images/constellation/versions/2.1.0` | `v2.1.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/2.0.0`     | `v2.0.0`  |
    | Azure | `/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation_CVM/images/constellation/versions/2.0.0` | `v2.0.0`  |
    | GCP   | `projects/constellation-images/global/images/constellation-v2-2-2`                                                                                                                    | `v2.2.2`  |
    | GCP   | `projects/constellation-images/global/images/constellation-v2-2-1`                                                                                                                    | `v2.2.1`  |
    | GCP   | `projects/constellation-images/global/images/constellation-v2-2-0`                                                                                                                    | `v2.2.0`  |
    | GCP   | `projects/constellation-images/global/images/constellation-v2-1-0`                                                                                                                    | `v2.1.0`  |
    | GCP   | `projects/constellation-images/global/images/constellation-v2-0-0`                                                                                                                    | `v2.0.0`  |
    </details>
- The `enforcedMeasurements` field has been removed and merged with the `measurements` field.
  - To migrate your config containing a new image (`v2.3` or greater), remove the old `measurements` and `enforcedMeasurements` entries from your config and run `constellation fetch-measurements`
  - To migrate your config containing an image older than `v2.3`, remove the `enforcedMeasurements` entry and replace the entries in `measurements` as shown in the example below:

    ```diff
    measurements:
    -    0: DzXCFGCNk8em5ornNZtKi+Wg6Z7qkQfs5CfE3qTkOc8=
    +    0:
    +        expected: DzXCFGCNk8em5ornNZtKi+Wg6Z7qkQfs5CfE3qTkOc8=
    +        warnOnly: true
    -    8: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
    +    8:
    +        expected: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
    +        warnOnly: false
    -enforcedMeasurements:
    -    - 8
    ```
