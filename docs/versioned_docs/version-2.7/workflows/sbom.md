# Consume software bill of materials (SBOMs)

<AsciinemaWidget src="/constellation/assets/check-sbom.cast" rows="20" cols="112" idleTimeLimit="3" preload="true" theme="edgeless" />

---

Constellation builds produce a [software bill of materials (SBOM)](https://www.ntia.gov/SBOM) for each generated [artifact](../architecture/microservices.md).
You can use SBOMs to make informed decisions about dependencies and vulnerabilities in a given application. Enterprises rely on SBOMs to maintain an inventory of used applications, which allows them to take data-driven approaches to managing risks related to vulnerabilities.

SBOMs for Constellation are generated using [Syft](https://github.com/anchore/syft), signed using [Cosign](https://github.com/sigstore/cosign), and stored with the produced artifact.

:::note
The public key for Edgeless Systems' long-term code-signing key is:

```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf8F1hpmwE+YCFXzjGtaQcrL6XZVT
JmEe5iSLvG1SyQSAew7WdMKF6o9t8e2TFuCkzlOhhlws2OHWbiFZnFWCFw==
-----END PUBLIC KEY-----
```

The public key is also available for download at [https://edgeless.systems/es.pub](https://edgeless.systems/es.pub) and in the Twitter profile [@EdgelessSystems](https://twitter.com/EdgelessSystems).

Make sure the key is available in a file named `cosign.pub` to execute the following examples.
:::

## Verify and download SBOMs

The following sections detail how to work with each type of artifact to verify and extract the SBOM.

### Constellation CLI

The SBOM for Constellation CLI is made available on the [GitHub release page](https://github.com/edgelesssys/constellation/releases). The SBOM (`constellation.spdx.sbom`) and corresponding signature (`constellation.spdx.sbom.sig`) are valid for each Constellation CLI for a given version, regardless of architecture and operating system.

```bash
curl -LO https://github.com/edgelesssys/constellation/releases/download/v2.2.0/constellation.spdx.sbom
curl -LO https://github.com/edgelesssys/constellation/releases/download/v2.2.0/constellation.spdx.sbom.sig
cosign verify-blob --key cosign.pub --signature constellation.spdx.sbom.sig constellation.spdx.sbom
```

### Container Images

SBOMs for container images are [attached to the image using Cosign](https://docs.sigstore.dev/cosign/signing/other_types/#sboms-software-bill-of-materials) and uploaded to the same registry.

As a consumer, use cosign to download and verify the SBOM:

```bash
# Verify and download the attestation statement
cosign verify-attestation ghcr.io/edgelesssys/constellation/verification-service@v2.2.0 --type 'https://cyclonedx.org/bom' --key cosign.pub --output-file verification-service.att.json
# Extract SBOM from attestation statement
jq -r .payload verification-service.att.json | base64 -d > verification-service.cyclonedx.sbom
```

A successful verification should result in similar output:

```shell-session
$ cosign verify-attestation ghcr.io/edgelesssys/constellation/verification-service@v2.2.0 --type 'https://cyclonedx.org/bom' --key cosign.pub --output-file verification-service.sbom

Verification for ghcr.io/edgelesssys/constellation/verification-service@v2.2.0 --
The following checks were performed on each of these signatures:
  - The cosign claims were validated
  - The signatures were verified against the specified public key
$ jq -r .payload verification-service.sbom | base64 -d > verification-service.cyclonedx.sbom
```

:::note

This example considers only the `verification-service`. The same approach works for all containers in the [Constellation container registry](https://github.com/orgs/edgelesssys/packages?repo_name=constellation).

:::

<!--
TODO: Once mkosi is implemented
## Operating System
-->

## Vulnerability scanning

You can use a plethora of tools to consume SBOMs. This section provides suggestions for tools that are popular and known to produce reliable results, but any tool that consumes [SPDX](https://spdx.dev/) or [CycloneDX](https://cyclonedx.org/) files should work.

Syft is able to [convert between the two formats](https://github.com/anchore/syft#format-conversion-experimental) in case you require a specific type.

### Grype

[Grype](https://github.com/anchore/grype) is a CLI tool that lends itself well for integration into CI/CD systems or local developer machines. It's also able to consume the signed attestation statement directly and does the verification in one go.

```bash
grype att:verification-service.sbom --key cosign.pub --add-cpes-if-none -q
```

### Dependency Track

[Dependency Track](https://dependencytrack.org/) is one of the oldest and most mature solutions when it comes to managing software inventory and vulnerabilities. Once imported, it continuously scans SBOMs for new vulnerabilities. It supports the CycloneDX format and provides direct guidance on how to comply with [U.S. Executive Order 14028](https://docs.dependencytrack.org/usage/executive-order-14028/).
