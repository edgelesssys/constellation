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

The configuration file then contains a list of `measurements` similar to the following:

```yaml
# ...
measurements:
    0:
        expected: "0f35c214608d93c7a6e68ae7359b4a8be5a0e99eea9107ece427c4dea4e439cf"
        warnOnly: false
    4:
        expected: "02c7a67c01ec70ffaf23d73a12f749ab150a8ac6dc529bda2fe1096a98bf42ea"
        warnOnly: false
    5:
        expected: "e6949026b72e5045706cd1318889b3874480f7a3f7c5c590912391a2d15e6975"
        warnOnly: true
    8:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
    9:
        expected: "f0a6e8601b00e2fdc57195686cd4ef45eb43a556ac1209b8e25d993213d68384"
        warnOnly: false
    11:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
    12:
        expected: "da99eb6cf7c7fbb692067c87fd5ca0b7117dc293578e4fea41f95d3d3d6af5e2"
        warnOnly: false
    13:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
    14:
        expected: "d7c4cc7ff7933022f013e03bdee875b91720b5b86cf1753cad830f95e791926f"
        warnOnly: true
    15:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
# ...
```

Each entry specifies the expected value of the Constellation node, and whether the measurement should be enforced (`warnOnly: false`), or only a warning should be logged (`warnOnly: true`).
By default, the subset of the [available measurements](../architecture/attestation.md#runtime-measurements) that can be locally reproduced and verified is enforced.

During attestation, the validating side (CLI or [join service](../architecture/microservices.md#joinservice)) compares each measurement reported by the issuing side (first node or joining node) individually.
For mismatching measurements that have set `warnOnly` to `true` only a warning is emitted.
For mismatching measurements that have set `warnOnly` to `false` an error is emitted and attestation fails.
If attestation fails for a new node, it isn't permitted to join the cluster.

## The *verify* command

:::note
The steps below are purely optional. They're automatically executed by `constellation apply` when you initialize your cluster. The `constellation verify` command mostly has an illustrative purpose.
:::

The `verify` command obtains and verifies an attestation statement from a running Constellation cluster.

```bash
constellation verify [--cluster-id ...]
```

From the attestation statement, the command verifies the following properties:

* The cluster is using the correct Confidential VM (CVM) type.
* Inside the CVMs, the correct node images are running. The node images are identified through the measurements obtained in the previous step.
* The unique ID of the cluster matches the one from your `constellation-state.yaml` file or passed in via `--cluster-id`.

Once the above properties are verified, you know that you are talking to the right Constellation cluster and it's in a good and trustworthy shape.

### Custom arguments

The `verify` command also allows you to verify any Constellation deployment that you have network access to. For this you need the following:

* The IP address of a running Constellation cluster's [VerificationService](../architecture/microservices.md#verificationservice). The `VerificationService` is exposed via a `NodePort` service using the external IP address of your cluster. Run `kubectl get nodes -o wide` and look for `EXTERNAL-IP`.
* The cluster's *clusterID*. See [cluster identity](../architecture/keys.md#cluster-identity) for more details.
* A `constellation-conf.yaml` file with the expected measurements of the cluster in your working directory.

For example:

```shell-session
constellation verify -e 192.0.2.1 --cluster-id Q29uc3RlbGxhdGlvbkRvY3VtZW50YXRpb25TZWNyZXQ=
```
