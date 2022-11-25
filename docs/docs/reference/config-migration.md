# Configuration Migrations

This document describes breaking changes in the configuration file format between Constellation releases.

## Migrating from CLI versions < 2.3

- The `sshUsers` was deprecated in v2.2 and now has been eventually removed from the configuration in v2.3.
  As an alternative for SSH, check the workflow section [Connect to nodes](https://constellation-docs.edgeless.systems/constellation/workflows/troubleshooting#connect-to-nodes).
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
