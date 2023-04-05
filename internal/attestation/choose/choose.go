/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package choose

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/aws"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/variant"
)

// Issuer returns the issuer for the given variant.
func Issuer(attestationVariant variant.Variant, log vtpm.AttestationLogger) (atls.Issuer, error) {
	switch attestationVariant {
	case variant.AWSNitroTPM{}:
		return aws.NewIssuer(log), nil
	case variant.AzureTrustedLaunch{}:
		return trustedlaunch.NewIssuer(log), nil
	case variant.AzureSEVSNP{}:
		return snp.NewIssuer(log), nil
	case variant.GCPSEVES{}:
		return gcp.NewIssuer(log), nil
	case variant.QEMUVTPM{}:
		return qemu.NewIssuer(log), nil
	case variant.Dummy{}:
		return atls.NewFakeIssuer(variant.Dummy{}), nil
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", attestationVariant)
	}
}

// Validator returns the validator for the given variant.
func Validator(
	attestationVariant variant.Variant, measurements measurements.M, idKeyCfg config.SNPFirmwareSignerConfig, log vtpm.AttestationLogger,
) (atls.Validator, error) {
	switch attestationVariant {
	case variant.AWSNitroTPM{}:
		return aws.NewValidator(config.AWSNitroTPM{Measurements: measurements}, log), nil
	case variant.AzureTrustedLaunch{}:
		return trustedlaunch.NewValidator(config.AzureTrustedLaunch{Measurements: measurements}, log), nil
	case variant.AzureSEVSNP{}:
		cfg := config.DefaultForAzureSEVSNP()
		cfg.Measurements = measurements
		cfg.FirmwareSignerConfig = idKeyCfg
		return snp.NewValidator(cfg, log), nil
	case variant.GCPSEVES{}:
		return gcp.NewValidator(config.GCPSEVES{Measurements: measurements}, log), nil
	case variant.QEMUVTPM{}:
		return qemu.NewValidator(config.QEMUVTPM{Measurements: measurements}, log), nil
	case variant.Dummy{}:
		return atls.NewFakeValidator(variant.Dummy{}), nil
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", attestationVariant)
	}
}
