# Attestation configuration options

To allow users more in-depth control over validating attestation statements, a separate, attestation-specific entry in the user's Constellation config file needs to be added.

This option should allow for more in-depth configuration of how and what parts of an attestation statement are validated.
For example, one should be able to set the minimum microcode version of an SEV-SNP attestation statement, that is deemed acceptable,
or configure acceptable TCB statuses for Intel TDX attestation.

## In the user's config

A new `attestation` entry in the config file will be added, holding the options used for attestation validation.
Existing attestation options, like `measurements` and `idKeyDigests` will be moved to this entry.

This entry will replace the `attestationVariant` entry.

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
  # See our docs for information on about configurable values [Link to our docs explaining SEV-SNP on Azure]
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
      #     - 'maaFallback' : Use 'equal' checking for validation, but fallback to using MAA for validation if the reported digest does not match any of the values in 'acceptedKeyDigests'
      #     - 'warnOnly'    : Same as 'equal', but only prints a warning instead of returning an error if no match is found
      enforcementPolicy: maaFallback
    # Lowest acceptable microcode version
    microcodeVersion: 115
    # Lowest acceptable SNP version
    snpVersion: 8
    # Lowest acceptable bootloader version
    bootloaderVersion: 3
```

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
since these are currently used by all attestation variants
