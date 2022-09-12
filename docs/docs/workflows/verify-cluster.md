# Verify your cluster

Constellation's [attestation feature](../architecture/attestation.md) allows you, or a third party, to verify the integrity (and confidentiality) of your Constellation.

## Fetch measurements

To verify the integrity of Constellation you need trusted measurements to verify against. For each of the released images there are signed measurements, which you can download using the CLI:

```bash
constellation config fetch-measurements
```

This command performs the following steps:
1. Download the signed measurements for the configured image. By default, this will use Edgeless Systems' public measurement registry. 
3. Verify the signed images. By default, this will use Edgeless Systems' [public key](https://edgeless.systems/es.pub). 
4. Write measurements into configuration file.

## The *verify* command

Once measurements are configured, this command verifies an attestation statement issued by a Constellation, thereby verifying the integrity and confidentiality of the whole cluster.

The following command performs attestation on the Constellation deployment in your current workspace:

```bash
constellation verify --cluster-id [...]
```

The command ensures that the value passed as `--cluster-id` matches the unique *clusterID* presented in the attestation statement.
This allows you to verify that you are connecting to the right Constellation instance
Additionally, the command verifies that the Confidential VM type and node images used by your Constellation deployment match the expected configurations.

### Custom arguments

The `verify` command also allows you to verify any Constellation deployment that you have network access to. For this you need to following:

* The IP address of a running Constellation deployment's [VerificationService](../architecture/components.md#verification-service). The `VerificationService` is exposed via a `NodePort` service using the external IP address of your cluster. Run `kubectl get nodes -o wide` and look for `EXTERNAL-IP`.
* The deployment's *clusterID*. See [cluster identity](../architecture/keys.md#cluster-identity) for more details.

For example:

```shell-session
constellation verify -e 192.0.2.1 --cluster-id Q29uc3RlbGxhdGlvbkRvY3VtZW50YXRpb25TZWNyZXQ=
```
