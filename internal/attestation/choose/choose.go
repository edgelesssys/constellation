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
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/v2/internal/oid"
)

// Issuer returns the issuer for the given variant.
func Issuer(variant oid.Getter, log vtpm.AttestationLogger) (atls.Issuer, error) {
	switch variant {
	case oid.AWSNitroTPM{}:
		return aws.NewIssuer(log), nil
	case oid.AzureTrustedLaunch{}:
		return trustedlaunch.NewIssuer(log), nil
	case oid.AzureSEVSNP{}:
		return snp.NewIssuer(log), nil
	case oid.GCPSEVES{}:
		return gcp.NewIssuer(log), nil
	case oid.QEMUVTPM{}:
		return qemu.NewIssuer(log), nil
	case oid.Dummy{}:
		return atls.NewFakeIssuer(oid.Dummy{}), nil
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", variant)
	}
}

// Validator returns the validator for the given variant.
func Validator(
	variant oid.Getter, measurements measurements.M, idKeyCfg idkeydigest.Config, log vtpm.AttestationLogger,
) (atls.Validator, error) {
	switch variant {
	case oid.AWSNitroTPM{}:
		return aws.NewValidator(measurements, log), nil
	case oid.AzureTrustedLaunch{}:
		return trustedlaunch.NewValidator(measurements, log), nil
	case oid.AzureSEVSNP{}:
		return snp.NewValidator(measurements, idKeyCfg, log), nil
	case oid.GCPSEVES{}:
		return gcp.NewValidator(measurements, log), nil
	case oid.QEMUVTPM{}:
		return qemu.NewValidator(measurements, log), nil
	case oid.Dummy{}:
		return atls.NewFakeValidator(oid.Dummy{}), nil
	default:
		return nil, fmt.Errorf("unknown attestation variant: %s", variant)
	}
}
