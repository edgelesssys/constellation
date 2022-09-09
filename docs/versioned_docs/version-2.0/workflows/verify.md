# Verify your cluster

Constellation's [attestation feature](../architecture/attestation.md) allows you, or a third party, to verify the confidentiality and integrity of your Constellation.

## Fetch measurements

To verify the integrity of Constellation you need trusted measurements to verify against. For each of the released images there are signed measurements, which you can download using the CLI:

```bash
constellation config fetch-measurements
```

This command performs the following steps:
1. Look up the signed measurements for the configured image.
2. Download the measurements.
3. Verify the signature.
4. Write measurements into configuration file.

## The *verify* command

Once measurements are configured, this command verifies an attestation statement issued by a Constellation, thereby verifying the integrity and confidentiality of the whole cluster.

The following command performs attestation on the Constellation in your current workspace:

<tabs>
<tabItem value="azure" label="Azure" default>

```bash
constellation verify azure
```

</tabItem>
<tabItem value="gcp" label="GCP" default>

```bash
constellation verify gcp
```

</tabItem>
</tabs>

The command makes sure the value passed to `-cluster-id` matches the *clusterID* presented in the attestation statement.
This allows you to verify that you are connecting to a specific Constellation instance
Additionally, the confidential computing capabilities, as well as the VM image, are verified to match the expected configurations.

### Custom arguments

You can provide additional arguments for `verify` to verify any Constellation you have network access to. This requires you to provide:

* The IP address of a running Constellation's [VerificationService](../architecture/components.md#verification-service). The *VerificationService* is exposed via a NodePort service using the external IP address of your cluster. Run `kubectl get nodes -o wide` and look for `EXTERNAL-IP`.
* The Constellation's *clusterID*. See [cluster identity](../architecture/keys.md#cluster-identity) for more details.

<tabs>
<tabItem value="azure" label="Azure" default>

```bash
constellation verify azure -e 192.0.2.1 --cluster-id Q29uc3RlbGxhdGlvbkRvY3VtZW50YXRpb25TZWNyZXQ=
```

</tabItem>
<tabItem value="gcp" label="GCP" default>

```bash
constellation verify gcp -e 192.0.2.1 --cluster-id Q29uc3RlbGxhdGlvbkRvY3VtZW50YXRpb25TZWNyZXQ=
```

</tabItem>
</tabs>
