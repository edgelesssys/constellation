# RFC 011: Attestation configuration options

To allow users more in-depth control over validating attestation statements, a separate, attestation-specific entry in the user's Constellation config file needs to be added.

This option should allow for more in-depth configuration of how and what parts of an attestation statement are validated.
For example, one should be able to set the minimum microcode version of an SEV-SNP attestation statement, that is deemed acceptable,
or configure acceptable TCB statuses for Intel TDX attestation.

## In the user's config

A new `attestation` entry in the config file will be added, holding the options used for attestation validation.
Existing attestation options, like `measurements` and `idKeyDigests` will be moved to this entry.

This entry will replace the `attestationVariant` entry.

Configuration entries that specify a minimum acceptable version, for example `microcodeVersion` for `azure-sev-snp`,
may instead use the meta value `latest`.
This will configure the CLI to use the latest available settings for the given environment.
The `latest` value will not be static for a given release, but is dynamically updated for all releases,
as it is independent of out images, and may need changing if a CSP makes changes to their infrastructure.

Example configuration for Azure SEV-SNP:

```yaml
version: v3
image: vX.Y.Z
name: constell
stateDiskSizeGB: 30
kubernetesVersion: "v1.25.6"
debugCluster: false
provider:
  azure: {} # Configuration for creating Azure cloud resources
# Configuration for attestation validation. This configuration provides sensible defaults for the Constellation version it was created for.
# See our docs for an overview on attestation: https://docs.edgeless.systems/constellation/architecture/attestation
attestation:
  # Azure SEV-SNP attestation configuration
  # See our docs for information about configurable values [Link to our docs explaining SEV-SNP on Azure]
  azureSEVSNP:
    # Expected confidential VM image measurements.
    measurements:
      15:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
    # Configuration for validating the firmware signature.
    firmwareSignerConfig:
      # List of accepted values for the firmware signing key digest.
      acceptedKeyDigests:
        - 7486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740cc96
      # Key digest enforcement policy. One of {'equal', 'maaFallback', 'warnOnly'}
      #     - 'equal'       : Error if the reported signing key digest does not match any of the values in 'acceptedKeyDigests'
      #     - 'maaFallback' : Use 'equal' checking for validation, but fallback to using Microsoft Azure Attestation (MAA) for validation if the reported digest does not match any of the values in 'acceptedKeyDigests'. See the Azure docs for more details: https://learn.microsoft.com/en-us/azure/attestation/overview#amd-sev-snp-attestation
      #     - 'warnOnly'    : Same as 'equal', but only prints a warning instead of returning an error if no match is found
      enforcementPolicy: maaFallback
    # Lowest acceptable microcode version
    microcodeVersion: 115
    # Lowest acceptable SNP version
    snpVersion: 8
    # Lowest acceptable bootloader version
    bootloaderVersion: latest
```

### AWS SEV-SNP

```yaml
attestation:
  # AWS SEV-SNP attestation.
  awsSEVSNP:
    # Expected TPM measurements.
    measurements:
      15:
        expected: "0000000000000000000000000000000000000000000000000000000000000000"
        warnOnly: false
    # Expected launch measurement in SNP report.
    launchMeasurement:
      # LaunchMeasurement enforcement policy. One of {'equal', 'warnOnly'}
      enforcementPolicy: equal
      validValues:
        - "c2c84b9364fc9f0f54b04534768c860c6e0e386ad98b96e8b98eca46ac8971d05c531ba48373f054c880cfd1f4a0a84e"
```

We want to allow users to disable enforcement.
This is because AWS may roll out unanounced/unreleased firmwares.
Such rollouts should not compromise cluster stability.

Multiple valid values are required since a cluster may have nodes with different launch-measurements during a firmware rollout.

Both values should be appliable through `upgrade apply` to easily react to changing measurements during cluster operation.
## In our code

`/internal/config/` holds default values that will be written to the config file.

`atls.Valdidator` initializers should accept a configuration struct which holds information about what values should be verified and how.
Each Validator should provide default values that can imported by the config.

```Go
attestCfg := sevcfg.New()
cfg.Measurements = measurements
validator := sev.NewValidator(attestCfg, log)
```

A new method `*Config.AttestationConfig()` returns an interface of the attestation configuration.
The interface should provide a `*Cfg.Variant()` method to retrieve the attestation variant used by the cluster.
This should work similar to how we retrieve the CSP using `*Config.GetProvider()`.

Creation of a Validator should require using a concrete type, so the attestation config interface needs to be cast to the configuration required by the validator.

Additional methods to set/retrieve common values of the config interface may be defined, for example to set expected measurements,
since these are currently used by all attestation variants.

If an attestation config specifies the minimum version of a parameter as `latest`,
that value is substituted with the most recent version of that parameter for the given CSP from our API.
Substituting values should use a similar signature verification logic for the config's signature as is used by the measurements `*M.FetchAndVerify(...)` flow.
The value substitution is part of the unmarshalling logic.

## Attestation config API

Config values are uploaded to S3 and can be accessed via HTTP.

The attestation config API uses the same CSP names as [the image API](./009-image-api.md#image-api-endpoints).

The following HTTP endpoint is available:

- `GET /constellation/v1/attestation/<ATTESTATION_VARIANT>/`
  - `list` returns a sorted list of available configurations for a given attestation variant
  - `<YEAR>-<MONTH>-<DAY>-<HOUR>-<MINUTE>.json`, e.g. `2023-01-23-14-32.json` returns an attestation config for the given date, if it exists. A list of available configs can be queried using the `list` endpoint.
  - `<YEAR>-<MONTH>-<DAY>-<HOUR>-<MINUTE>.json.sig` returns the signature of the attestation config file.

While this API should stay compatible with old release, extensive changes to our code may require breaking changes to the format of the attestation config files.
In this case a new API version will be used to retrieve the config in the updated format, e.g. `/constellation/v2/attestation/<ATTESTATION_VARIANT>/`.
The old API will still receive updates for at least the next release cycle, during this time this API version will also return a deprecation warning when requesting `list`.

### Azure SEV-SNP
IMPORTANT: Since the current version fetches from the Azure SEV-SNP report are not guaranteed to be globally rolled out at the time of the report, we introduce a minimum age (2 weeks) of the version to consider it a valid latest version.
This validation is only enforced on the fetcher side! This means that the HTTP endpoints contain all versions, even those that do not yet have the minimum age.
### AWS

AWS provides a way to precalculate launch-measurements for their firmware in SEV-SNP CVMs.
Since the launch-measurement can change at any point in time we need to serve up-to-date measurements through the attestation config API.
This will enable users to (a) check which measurements are currently available and manually select any of them.
It will also enable users to (b) specify a value `latest` for the launch-measurement.

#### Object structure
Each object represents parts of a valid attestation config.
There are at most three versions of the launch-measurements available in one API object.
Two versions that represent measurements we have already seen on an EC2 instance.
One version that represents measurements of a newly released firmware that has not been seen on any machine.

To sort the versions within each object we use a key `addedOn`.
To determine if a measurement has been deployed yet we use a key `firstSeenOn`.
If `firstSeenOn` is set to a placeholder value the underlying firmware has not been seen yet, but it might be seen by a user (we can't know that).

**/list:**
```
[
  "2023-01-23-14-32",
  "2023-01-10-14-32"
]
```

**/2023-01-23-14-32.json:**
```
{
  "launchMeasurements": [
    {
      "firstSeenOn": "-"
      "value": "v1.3.1",
    },
    {
      "firstSeenOn": "2023-01-14T15:04:05Z07:00"
      "value": "v1.3",
    },
    {
      "firstSeenOn": "2023-01-01T15:04:05Z07:00"
      "value": "v1.2",
    }
  ]
}
```

#### API behavior

We are assuming that AWS will not release a new firmware, without completing the rollout of the currently released version (i.e. there are at most two distinct firmware versions at all times).

A pipeline is responsible for updating the API:

1. periodically scan [aws/uefi](https://github.com/aws/uefi) for new releases, build them and precalculate measurements `new`.
2. `firstSeenOn` is set to a placeholder value.
3. start new AWS EC2 instance and fetch a new SNP report from that machine, including the current measurement `current`.
4. get the `latest` object from the API and compare the included launch-measurements to `current`. Any modifications are done by creating a new object and adding it to the API.
  - `if !latest.launchMeasurements.Contains(new)`: add `new` with `firstSeenOn` set to `placeholder`. The oldest measurement is removed.
  - `if latest.launchMeasurements.Latest().firstSeenOn == placeholder && latest.launchMeasurements.Latest().value == current.value`: update `firstSeenOn` to current date, remove the oldest measurement and any older (than `Latest()`) measurements with `firstSeenOn==placeholder`.
  - `if latest.launchMeasurements.Latest().firstSeenOn != placeholder && latest.launchMeasurements.Latest().value == current.value`: do nothing.
  - `if !list.Contains(current)`: fail. Alerts us of unreleased/unplanned firmware updates.
5. If the new object from step 4 is different than `latest`, push it to the API.

The pipelines should run ~ daily.

Ideally AWS will announce firmware changes in the future as part of maintenance announcements.
This announcement should include a timeframe when the maintenance will start and when all machines will have the new firmware.
If AWS starts doing this, the update logic in step 1-4. will change.
