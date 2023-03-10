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
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	testPCRs := measurements.M{
		0: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength),
		1: measurements.WithAllBytes(0xFF, false, measurements.PCRMeasurementLength),
		2: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength),
		3: measurements.WithAllBytes(0xFF, false, measurements.PCRMeasurementLength),
		4: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength),
		5: measurements.WithAllBytes(0x00, false, measurements.PCRMeasurementLength),
	}

	testCases := map[string]struct {
		provider           cloudprovider.Provider
		config             *config.Config
		pcrs               measurements.M
		enforceIDKeyDigest bool
		digest             idkeydigest.IDKeyDigests
		azureCVM           bool
		wantErr            bool
	}{
		"gcp": {
			provider: cloudprovider.GCP,
			pcrs:     testPCRs,
		},
		"azure cvm": {
			provider: cloudprovider.Azure,
			pcrs:     testPCRs,
			azureCVM: true,
		},
		"azure trusted launch": {
			provider: cloudprovider.Azure,
			pcrs:     testPCRs,
			azureCVM: false,
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			pcrs:     testPCRs,
		},
		"no pcrs provided": {
			provider: cloudprovider.Azure,
			pcrs:     measurements.M{},
			wantErr:  true,
		},
		"unknown provider": {
			provider: cloudprovider.Unknown,
			pcrs:     testPCRs,
			wantErr:  true,
		},
		"set idkeydigest": {
			provider:           cloudprovider.Azure,
			pcrs:               testPCRs,
			digest:             idkeydigest.IDKeyDigests{[]byte("414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141")},
			enforceIDKeyDigest: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			conf := &config.Config{Provider: config.ProviderConfig{}}
			if tc.provider == cloudprovider.GCP {
				conf.Provider.GCP = &config.GCPConfig{Measurements: tc.pcrs}
			}
			if tc.provider == cloudprovider.Azure {
				conf.Provider.Azure = &config.AzureConfig{Measurements: tc.pcrs, EnforceIDKeyDigest: &tc.enforceIDKeyDigest, IDKeyDigest: tc.digest, ConfidentialVM: &tc.azureCVM}
			}
			if tc.provider == cloudprovider.QEMU {
				conf.Provider.QEMU = &config.QEMUConfig{Measurements: tc.pcrs}
			}

			validators, err := NewValidator(tc.provider, conf, logger.NewTest(t))

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.pcrs, validators.pcrs)
				assert.Equal(tc.provider, validators.provider)
			}
		})
	}
}

func TestValidatorV(t *testing.T) {
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			1:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			2:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			3:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			4:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			5:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			6:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			7:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			8:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			9:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			10: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			11: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			12: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
		}
	}

	testCases := map[string]struct {
		provider cloudprovider.Provider
		pcrs     measurements.M
		wantVs   atls.Validator
		azureCVM bool
	}{
		"gcp": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
			wantVs:   gcp.NewValidator(newTestPCRs(), nil),
		},
		"azure cvm": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			wantVs:   snp.NewValidator(newTestPCRs(), idkeydigest.IDKeyDigests{}, false, nil),
			azureCVM: true,
		},
		"azure trusted launch": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			wantVs:   trustedlaunch.NewValidator(newTestPCRs(), nil),
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			pcrs:     newTestPCRs(),
			wantVs:   qemu.NewValidator(newTestPCRs(), nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{provider: tc.provider, pcrs: tc.pcrs, azureCVM: tc.azureCVM}

			resultValidator := validators.V(&cobra.Command{})

			assert.Equal(tc.wantVs.OID(), resultValidator.OID())
		})
	}
}

func TestValidatorUpdateInitPCRs(t *testing.T) {
	zero := measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength)
	one := measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength)
	one64 := base64.StdEncoding.EncodeToString(one.Expected[:])
	oneHash := sha256.Sum256(one.Expected[:])
	pcrZeroUpdatedOne := sha256.Sum256(append(zero.Expected[:], oneHash[:]...))
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			1:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			2:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			3:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			4:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			5:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			6:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			7:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			8:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			9:  measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			10: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			11: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			12: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			13: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			14: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			15: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			16: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
			17: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			18: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			19: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			20: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			21: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			22: measurements.WithAllBytes(0x11, true, measurements.PCRMeasurementLength),
			23: measurements.WithAllBytes(0x00, true, measurements.PCRMeasurementLength),
		}
	}

	testCases := map[string]struct {
		provider  cloudprovider.Provider
		pcrs      measurements.M
		ownerID   string
		clusterID string
		wantErr   bool
	}{
		"gcp update owner ID": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
			ownerID:  one64,
		},
		"gcp update cluster ID": {
			provider:  cloudprovider.GCP,
			pcrs:      newTestPCRs(),
			clusterID: one64,
		},
		"gcp update both": {
			provider:  cloudprovider.GCP,
			pcrs:      newTestPCRs(),
			ownerID:   one64,
			clusterID: one64,
		},
		"azure update owner ID": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			ownerID:  one64,
		},
		"azure update cluster ID": {
			provider:  cloudprovider.Azure,
			pcrs:      newTestPCRs(),
			clusterID: one64,
		},
		"azure update both": {
			provider:  cloudprovider.Azure,
			pcrs:      newTestPCRs(),
			ownerID:   one64,
			clusterID: one64,
		},
		"owner ID and cluster ID empty": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
		},
		"invalid encoding": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
			ownerID:  "invalid",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{provider: tc.provider, pcrs: tc.pcrs}

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
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

				case i == int(measurements.PCRIndexOwnerID) && tc.ownerID == "":
					// should be deleted
					_, ok := validators.pcrs[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexOwnerID):
					pcr, ok := validators.pcrs[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

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
		0: measurements.WithAllBytes(0xAA, false, measurements.PCRMeasurementLength),
		1: measurements.WithAllBytes(0xBB, false, measurements.PCRMeasurementLength),
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
				provider: cloudprovider.GCP,
				pcrs:     pcrs,
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
