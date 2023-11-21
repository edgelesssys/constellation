/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorUpdateInitPCRs(t *testing.T) {
	zero := measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength)
	one := measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength)
	one64 := base64.StdEncoding.EncodeToString(one.Expected[:])
	oneHash := sha256.Sum256(one.Expected[:])
	pcrZeroUpdatedOne := sha256.Sum256(append(zero.Expected[:], oneHash[:]...))
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			1:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			2:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			3:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			4:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			5:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			6:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			7:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			8:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			9:  measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			10: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			11: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			12: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			13: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			14: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			15: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			16: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
			17: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			18: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			19: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			20: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			21: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			22: measurements.WithAllBytes(0x11, measurements.WarnOnly, measurements.PCRMeasurementLength),
			23: measurements.WithAllBytes(0x00, measurements.WarnOnly, measurements.PCRMeasurementLength),
		}
	}

	testCases := map[string]struct {
		config    config.AttestationCfg
		ownerID   string
		clusterID string
		wantErr   bool
	}{
		"gcp update owner ID": {
			config: &config.GCPSEVES{
				Measurements: newTestPCRs(),
			},
			ownerID: one64,
		},
		"gcp update cluster ID": {
			config: &config.GCPSEVES{
				Measurements: newTestPCRs(),
			},
			clusterID: one64,
		},
		"gcp update both": {
			config: &config.GCPSEVES{
				Measurements: newTestPCRs(),
			},
			ownerID:   one64,
			clusterID: one64,
		},
		"azure update owner ID": {
			config: &config.AzureSEVSNP{
				Measurements: newTestPCRs(),
			},
			ownerID: one64,
		},
		"azure update cluster ID": {
			config: &config.AzureSEVSNP{
				Measurements: newTestPCRs(),
			},
			clusterID: one64,
		},
		"azure update both": {
			config: &config.AzureSEVSNP{
				Measurements: newTestPCRs(),
			},
			ownerID:   one64,
			clusterID: one64,
		},
		"owner ID and cluster ID empty": {
			config: &config.AzureSEVSNP{
				Measurements: newTestPCRs(),
			},
		},
		"invalid encoding": {
			config: &config.GCPSEVES{
				Measurements: newTestPCRs(),
			},
			ownerID: "invalid",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := UpdateInitMeasurements(tc.config, tc.ownerID, tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(t, err)
			m := tc.config.GetMeasurements()
			for i := 0; i < len(m); i++ {
				switch {
				case i == int(measurements.PCRIndexClusterID) && tc.clusterID == "":
					// should be deleted
					_, ok := m[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexClusterID):
					pcr, ok := m[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

				case i == int(measurements.PCRIndexOwnerID) && tc.ownerID == "":
					// should be deleted
					_, ok := m[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexOwnerID):
					pcr, ok := m[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

				default:
					if i >= 17 && i <= 22 {
						assert.Equal(one, m[uint32(i)])
					} else {
						assert.Equal(zero, m[uint32(i)])
					}
				}
			}
		})
	}
}

func TestValidatorUpdateInitMeasurementsTDX(t *testing.T) {
	zero := measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength)
	one := measurements.WithAllBytes(0x11, true, measurements.TDXMeasurementLength)
	one64 := base64.StdEncoding.EncodeToString(one.Expected[:])
	oneHash := sha512.Sum384(one.Expected[:])
	tdxZeroUpdatedOne := sha512.Sum384(append(zero.Expected[:], oneHash[:]...))
	newTestTDXMeasurements := func() measurements.M {
		return measurements.M{
			0: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			1: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			2: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			3: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			4: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
		}
	}

	testCases := map[string]struct {
		measurements measurements.M
		clusterID    string
		wantErr      bool
	}{
		"QEMUT TDX update update cluster ID": {
			measurements: newTestTDXMeasurements(),
			clusterID:    one64,
		},
		"cluster ID empty": {
			measurements: newTestTDXMeasurements(),
		},
		"invalid encoding": {
			measurements: newTestTDXMeasurements(),
			clusterID:    "invalid",
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cfg := &config.QEMUTDX{Measurements: tc.measurements}

			err := UpdateInitMeasurements(cfg, "", tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			for i := 0; i < len(tc.measurements); i++ {
				switch {
				case i == measurements.TDXIndexClusterID && tc.clusterID == "":
					// should be deleted
					_, ok := cfg.Measurements[uint32(i)]
					assert.False(ok)

				case i == measurements.TDXIndexClusterID:
					pcr, ok := cfg.Measurements[uint32(i)]
					assert.True(ok)
					assert.Equal(tdxZeroUpdatedOne[:], pcr.Expected)

				default:
					assert.Equal(zero, cfg.Measurements[uint32(i)])
				}
			}
		})
	}
}
