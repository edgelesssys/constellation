/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package choose

import (
	"encoding/asn1"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuer(t *testing.T) {
	testCases := map[string]struct {
		variant oid.Getter
		wantErr bool
	}{
		"aws-nitro-tpm": {
			variant: oid.AWSNitroTPM{},
		},
		"azure-sev-snp": {
			variant: oid.AzureSEVSNP{},
		},
		"azure-trusted-launch": {
			variant: oid.AzureTrustedLaunch{},
		},
		"gcp-sev-es": {
			variant: oid.GCPSEVES{},
		},
		"qemu-vtpm": {
			variant: oid.QEMUVTPM{},
		},
		"dummy": {
			variant: oid.Dummy{},
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
		variant oid.Getter
		wantErr bool
	}{
		"aws-nitro-tpm": {
			variant: oid.AWSNitroTPM{},
		},
		"azure-sev-snp": {
			variant: oid.AzureSEVSNP{},
		},
		"azure-trusted-launch": {
			variant: oid.AzureTrustedLaunch{},
		},
		"gcp-sev-es": {
			variant: oid.GCPSEVES{},
		},
		"qemu-vtpm": {
			variant: oid.QEMUVTPM{},
		},
		"dummy": {
			variant: oid.Dummy{},
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

			validator, err := Validator(tc.variant, nil, idkeydigest.Config{}, nil)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.True(validator.OID().Equal(tc.variant.OID()))
		})
	}
}

type unknownVariant struct{}

func (unknownVariant) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 9999, 9999}
}
