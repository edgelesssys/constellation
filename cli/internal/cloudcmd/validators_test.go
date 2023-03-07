/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/atls"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/v2/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	testPCRs := measurements.M{
		0: measurements.WithAllBytes(0x00, false),
		1: measurements.WithAllBytes(0xFF, false),
		2: measurements.WithAllBytes(0x00, false),
		3: measurements.WithAllBytes(0xFF, false),
		4: measurements.WithAllBytes(0x00, false),
		5: measurements.WithAllBytes(0x00, false),
	}

	testCases := map[string]struct {
		config  *config.Config
		wantErr bool
	}{
		"gcp": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.GCPSEVES{}.String(),
					GCP: &config.GCPConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"azure cvm": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.AzureSEVSNP{}.String(),
					Azure: &config.AzureConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"azure trusted launch": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.AzureTrustedLaunch{}.String(),
					Azure: &config.AzureConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"qemu": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.QEMUVTPM{}.String(),
					QEMU: &config.QEMUConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"no pcrs provided": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.AzureSEVSNP{}.String(),
					Azure: &config.AzureConfig{
						Measurements: measurements.M{},
					},
				},
			},
			wantErr: true,
		},
		"unknown variant": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: "unknown",
					QEMU: &config.QEMUConfig{
						Measurements: testPCRs,
					},
				},
			},
			wantErr: true,
		},
		"set idkeydigest": {
			config: &config.Config{
				Provider: config.ProviderConfig{
					AttestationVariant: oid.AzureSEVSNP{}.String(),
					Azure: &config.AzureConfig{
						Measurements:       testPCRs,
						IDKeyDigest:        idkeydigest.IDKeyDigests{[]byte("414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141")},
						EnforceIDKeyDigest: &[]bool{true}[0],
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators, err := NewValidator(tc.config, logger.NewTest(t))

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.config.GetMeasurements(), validators.pcrs)
				variant, err := oid.FromString(tc.config.Provider.AttestationVariant)
				require.NoError(t, err)
				assert.Equal(variant, validators.attestationVariant)
			}
		})
	}
}

func TestValidatorV(t *testing.T) {
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, true),
			1:  measurements.WithAllBytes(0x00, true),
			2:  measurements.WithAllBytes(0x00, true),
			3:  measurements.WithAllBytes(0x00, true),
			4:  measurements.WithAllBytes(0x00, true),
			5:  measurements.WithAllBytes(0x00, true),
			6:  measurements.WithAllBytes(0x00, true),
			7:  measurements.WithAllBytes(0x00, true),
			8:  measurements.WithAllBytes(0x00, true),
			9:  measurements.WithAllBytes(0x00, true),
			10: measurements.WithAllBytes(0x00, true),
			11: measurements.WithAllBytes(0x00, true),
			12: measurements.WithAllBytes(0x00, true),
		}
	}

	testCases := map[string]struct {
		variant oid.Getter
		pcrs    measurements.M
		wantVs  atls.Validator
	}{
		"gcp": {
			variant: oid.GCPSEVES{},
			pcrs:    newTestPCRs(),
			wantVs:  gcp.NewValidator(newTestPCRs(), nil),
		},
		"azure cvm": {
			variant: oid.AzureSEVSNP{},
			pcrs:    newTestPCRs(),
			wantVs:  snp.NewValidator(newTestPCRs(), idkeydigest.IDKeyDigests{}, false, nil),
		},
		"azure trusted launch": {
			variant: oid.AzureTrustedLaunch{},
			pcrs:    newTestPCRs(),
			wantVs:  trustedlaunch.NewValidator(newTestPCRs(), nil),
		},
		"qemu": {
			variant: oid.QEMUVTPM{},
			pcrs:    newTestPCRs(),
			wantVs:  qemu.NewValidator(newTestPCRs(), nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{attestationVariant: tc.variant, pcrs: tc.pcrs}

			resultValidator := validators.V(&cobra.Command{})

			assert.Equal(tc.wantVs.OID(), resultValidator.OID())
		})
	}
}

func TestValidatorUpdateInitPCRs(t *testing.T) {
	zero := measurements.WithAllBytes(0x00, true)
	one := measurements.WithAllBytes(0x11, true)
	one64 := base64.StdEncoding.EncodeToString(one.Expected[:])
	oneHash := sha256.Sum256(one.Expected[:])
	pcrZeroUpdatedOne := sha256.Sum256(append(zero.Expected[:], oneHash[:]...))
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, true),
			1:  measurements.WithAllBytes(0x00, true),
			2:  measurements.WithAllBytes(0x00, true),
			3:  measurements.WithAllBytes(0x00, true),
			4:  measurements.WithAllBytes(0x00, true),
			5:  measurements.WithAllBytes(0x00, true),
			6:  measurements.WithAllBytes(0x00, true),
			7:  measurements.WithAllBytes(0x00, true),
			8:  measurements.WithAllBytes(0x00, true),
			9:  measurements.WithAllBytes(0x00, true),
			10: measurements.WithAllBytes(0x00, true),
			11: measurements.WithAllBytes(0x00, true),
			12: measurements.WithAllBytes(0x00, true),
			13: measurements.WithAllBytes(0x00, true),
			14: measurements.WithAllBytes(0x00, true),
			15: measurements.WithAllBytes(0x00, true),
			16: measurements.WithAllBytes(0x00, true),
			17: measurements.WithAllBytes(0x11, true),
			18: measurements.WithAllBytes(0x11, true),
			19: measurements.WithAllBytes(0x11, true),
			20: measurements.WithAllBytes(0x11, true),
			21: measurements.WithAllBytes(0x11, true),
			22: measurements.WithAllBytes(0x11, true),
			23: measurements.WithAllBytes(0x00, true),
		}
	}

	testCases := map[string]struct {
		variant   oid.Getter
		pcrs      measurements.M
		ownerID   string
		clusterID string
		wantErr   bool
	}{
		"gcp update owner ID": {
			variant: oid.GCPSEVES{},
			pcrs:    newTestPCRs(),
			ownerID: one64,
		},
		"gcp update cluster ID": {
			variant:   oid.GCPSEVES{},
			pcrs:      newTestPCRs(),
			clusterID: one64,
		},
		"gcp update both": {
			variant:   oid.GCPSEVES{},
			pcrs:      newTestPCRs(),
			ownerID:   one64,
			clusterID: one64,
		},
		"azure update owner ID": {
			variant: oid.AzureSEVSNP{},
			pcrs:    newTestPCRs(),
			ownerID: one64,
		},
		"azure update cluster ID": {
			variant:   oid.AzureSEVSNP{},
			pcrs:      newTestPCRs(),
			clusterID: one64,
		},
		"azure update both": {
			variant:   oid.AzureSEVSNP{},
			pcrs:      newTestPCRs(),
			ownerID:   one64,
			clusterID: one64,
		},
		"owner ID and cluster ID empty": {
			variant: oid.GCPSEVES{},
			pcrs:    newTestPCRs(),
		},
		"invalid encoding": {
			variant: oid.GCPSEVES{},
			pcrs:    newTestPCRs(),
			ownerID: "invalid",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{attestationVariant: tc.variant, pcrs: tc.pcrs}

			err := validators.UpdateInitPCRs(tc.ownerID, tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			for i := 0; i < len(tc.pcrs); i++ {
				switch {
				case i == int(measurements.PCRIndexClusterID) && tc.clusterID == "":
					// should be deleted
					_, ok := validators.pcrs[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexClusterID):
					pcr, ok := validators.pcrs[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne, pcr.Expected)

				case i == int(measurements.PCRIndexOwnerID) && tc.ownerID == "":
					// should be deleted
					_, ok := validators.pcrs[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexOwnerID):
					pcr, ok := validators.pcrs[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne, pcr.Expected)

				default:
					if i >= 17 && i <= 22 {
						assert.Equal(one, validators.pcrs[uint32(i)])
					} else {
						assert.Equal(zero, validators.pcrs[uint32(i)])
					}
				}
			}
		})
	}
}

func TestUpdatePCR(t *testing.T) {
	emptyMap := measurements.M{}
	defaultMap := measurements.M{
		0: measurements.WithAllBytes(0xAA, false),
		1: measurements.WithAllBytes(0xBB, false),
	}

	testCases := map[string]struct {
		pcrMap      measurements.M
		pcrIndex    uint32
		encoded     string
		wantEntries int
		wantErr     bool
	}{
		"empty input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     "",
			wantEntries: 0,
			wantErr:     false,
		},
		"empty input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     "",
			wantEntries: len(defaultMap),
			wantErr:     false,
		},
		"correct input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries: 1,
			wantErr:     false,
		},
		"correct input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries: len(defaultMap) + 1,
			wantErr:     false,
		},
		"hex input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     hex.EncodeToString([]byte("Constellation")),
			wantEntries: 1,
			wantErr:     false,
		},
		"hex input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     hex.EncodeToString([]byte("Constellation")),
			wantEntries: len(defaultMap) + 1,
			wantErr:     false,
		},
		"unencoded input, empty map": {
			pcrMap:      emptyMap,
			pcrIndex:    10,
			encoded:     "Constellation",
			wantEntries: 0,
			wantErr:     true,
		},
		"unencoded input, default map": {
			pcrMap:      defaultMap,
			pcrIndex:    10,
			encoded:     "Constellation",
			wantEntries: len(defaultMap),
			wantErr:     true,
		},
		"empty input at occupied index": {
			pcrMap:      defaultMap,
			pcrIndex:    0,
			encoded:     "",
			wantEntries: len(defaultMap) - 1,
			wantErr:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			pcrs := make(measurements.M)
			for k, v := range tc.pcrMap {
				pcrs[k] = v
			}

			validators := &Validator{
				attestationVariant: oid.GCPSEVES{},
				pcrs:               pcrs,
			}
			err := validators.updatePCR(tc.pcrIndex, tc.encoded)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Len(pcrs, tc.wantEntries)
		})
	}
}
