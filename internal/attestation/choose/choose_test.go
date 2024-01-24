/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package choose

import (
	"encoding/asn1"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuer(t *testing.T) {
	testCases := map[string]struct {
		variant variant.Variant
		wantErr bool
	}{
		"aws-sev-snp": {
			variant: variant.AWSSEVSNP{},
		},
		"aws-nitro-tpm": {
			variant: variant.AWSNitroTPM{},
		},
		"azure-sev-snp": {
			variant: variant.AzureSEVSNP{},
		},
		"azure-tdx": {
			variant: variant.AzureTDX{},
		},
		"azure-trusted-launch": {
			variant: variant.AzureTrustedLaunch{},
		},
		"gcp-sev-es": {
			variant: variant.GCPSEVES{},
		},
		"qemu-vtpm": {
			variant: variant.QEMUVTPM{},
		},
		"dummy": {
			variant: variant.Dummy{},
		},
		"unknown": {
			variant: unknownVariant{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			issuer, err := Issuer(tc.variant, nil)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.True(issuer.OID().Equal(tc.variant.OID()))
		})
	}
}

func TestValidator(t *testing.T) {
	testCases := map[string]struct {
		cfg     config.AttestationCfg
		wantErr bool
	}{
		"aws-nitro-tpm": {
			cfg: &config.AWSNitroTPM{},
		},
		"azure-sev-snp": {
			cfg: &config.AzureSEVSNP{},
		},
		"azure-tdx": {
			cfg: &config.AzureTDX{},
		},
		"azure-trusted-launch": {
			cfg: &config.AzureTrustedLaunch{},
		},
		"gcp-sev-es": {
			cfg: &config.GCPSEVES{},
		},
		"qemu-vtpm": {
			cfg: &config.QEMUVTPM{},
		},
		"dummy": {
			cfg: &config.DummyCfg{},
		},
		"unknown": {
			cfg:     unknownConfig{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			validator, err := Validator(tc.cfg, nil)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.True(validator.OID().Equal(tc.cfg.GetVariant().OID()))
		})
	}
}

type unknownVariant struct{}

func (unknownVariant) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 9999, 9999}
}

func (unknownVariant) String() string {
	return "unknown"
}

func (unknownVariant) Equal(other variant.Getter) bool {
	return other.OID().Equal(unknownVariant{}.OID())
}

type unknownConfig struct{}

func (unknownConfig) GetVariant() variant.Variant {
	return unknownVariant{}
}

func (unknownConfig) GetMeasurements() measurements.M {
	return nil
}

func (unknownConfig) SetMeasurements(measurements.M) {}

func (unknownConfig) EqualTo(config.AttestationCfg) (bool, error) { return false, nil }
