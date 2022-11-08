# Verify your cluster

Constellation's [attestation feature](../architecture/attestation.md) allows you, or a third party, to verify the integrity and confidentiality of your Constellation cluster.

## Fetch measurements

To verify the integrity of Constellation you need trusted measurements to verify against. For each node image released by Edgeless Systems, there are signed measurements, which you can download using the CLI:

```bash
constellation config fetch-measurements
```

This command performs the following steps:
1. Download the signed measurements for the configured image. By default, this will use Edgeless Systems' public measurement registry.
2. Verify the signature of the measurements. This will use Edgeless Systems' [public key](https://edgeless.systems/es.pub).
3. Write measurements into configuration file.

After the command succeeded the configuration file will contain a list of key value pairs (index, hash) under the `measurements` key and a list of indices under `enforcedMeasurements`.
Not all keys under `measurements` will have a matching index under `enforcedMeasurements`.
This is because only a subset of the available measurements can be locally reproduced and verified.
These non-reproducible values typically measure parts of the boot process that are controlled by the CSP.
You can find a list with descriptions in the [runtime measurements](../architecture/attestation.md#runtime-measurements) section.

During attestation, the validating side (CLI or [join service](../architecture/components.md#joinservice)) will compare each [measurement](../architecture/attestation.md##runtime-measurement) reported by the issuing side (first node or joining node) individually.
For mismatching measurements that are set under the `measurements` key a warning will be emitted.
For mismatching measurements that are additionally set under `enforcedMeasurements` an error will be emitted and attestation fails.
If attestation fails, the new node can't join the cluster.

## The *verify* command

:::note
The steps below are purely optional. They're automatically executed by `constellation init` when you initialize your cluster. The `constellation verify` command mostly has an illustrative purpose.
:::

The `verify` command obtains and verifies an attestation statement from a running Constellation cluster.

```bash
constellation verify [--cluster-id ...]
```

From the attestation statement, the command verifies the following properties:
* The cluster is using the correct Confidential VM (CVM) type.
* Inside the CVMs, the correct node images are running. The node images are identified through the measurements obtained in the previous step.
* The unique ID of the cluster matches the one from your `constellation-id.json` file or passed in via `--cluster-id`.

Once the above properties are verified, you know that you are talking to the right Constellation cluster and it's in a good and trustworthy shape.

### Custom arguments

The `verify` command also allows you to verify any Constellation deployment that you have network access to. For this you need the following:

* The IP address of a running Constellation cluster's [VerificationService](../architecture/components.md#verification-service). The `VerificationService` is exposed via a `NodePort` service using the external IP address of your cluster. Run `kubectl get nodes -o wide` and look for `EXTERNAL-IP`.
* The cluster's *clusterID*. See [cluster identity](../architecture/keys.md#cluster-identity) for more details.

For example:

```shell-session
constellation verify -e 192.0.2.1 --cluster-id Q29uc3RlbGxhdGlvbkRvY3VtZW50YXRpb25TZWNyZXQ=
```
