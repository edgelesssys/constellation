/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorUpdateInitPCRs(t *testing.T) {
	zero := measurements.WithAllBytes(0x00, measurements.WarnOnly)
	one := measurements.WithAllBytes(0x11, measurements.WarnOnly)
	one64 := base64.StdEncoding.EncodeToString(one.Expected[:])
	oneHash := sha256.Sum256(one.Expected[:])
	pcrZeroUpdatedOne := sha256.Sum256(append(zero.Expected[:], oneHash[:]...))
	newTestPCRs := func() measurements.M {
		return measurements.M{
			0:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			1:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			2:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			3:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			4:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			5:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			6:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			7:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			8:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			9:  measurements.WithAllBytes(0x00, measurements.WarnOnly),
			10: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			11: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			12: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			13: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			14: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			15: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			16: measurements.WithAllBytes(0x00, measurements.WarnOnly),
			17: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			18: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			19: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			20: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			21: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			22: measurements.WithAllBytes(0x11, measurements.WarnOnly),
			23: measurements.WithAllBytes(0x00, measurements.WarnOnly),
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

			err := UpdateInitPCRs(tc.config, tc.ownerID, tc.clusterID)

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
