/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"crypto/sha256"
	"crypto/sha512"
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
	"github.com/edgelesssys/constellation/v2/internal/attestation/tdx"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/oid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	testTDXMeasurements := measurements.M{
		0: measurements.WithAllBytes(0x00, false, measurements.TDXMeasurementLength),
		1: measurements.WithAllBytes(0xFF, false, measurements.TDXMeasurementLength),
		2: measurements.WithAllBytes(0x00, false, measurements.TDXMeasurementLength),
		3: measurements.WithAllBytes(0xFF, false, measurements.TDXMeasurementLength),
		4: measurements.WithAllBytes(0x00, false, measurements.TDXMeasurementLength),
		5: measurements.WithAllBytes(0x00, false, measurements.TDXMeasurementLength),
	}

	testCases := map[string]struct {
		config  *config.Config
		wantErr bool
	}{
		"gcp": {
			config: &config.Config{
				AttestationVariant: oid.GCPSEVES{}.String(),
				Provider: config.ProviderConfig{
					GCP: &config.GCPConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"azure cvm": {
			config: &config.Config{
				AttestationVariant: oid.AzureSEVSNP{}.String(),
				Provider: config.ProviderConfig{
					Azure: &config.AzureConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"azure trusted launch": {
			config: &config.Config{
				AttestationVariant: oid.AzureTrustedLaunch{}.String(),
				Provider: config.ProviderConfig{
					Azure: &config.AzureConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"qemu vTPM": {
			config: &config.Config{
				AttestationVariant: oid.QEMUVTPM{}.String(),
				Provider: config.ProviderConfig{
					QEMU: &config.QEMUConfig{
						Measurements: testPCRs,
					},
				},
			},
		},
		"qemu TDX": {
			config: &config.Config{
				AttestationVariant: oid.QEMUTDX{}.String(),
				Provider: config.ProviderConfig{
					QEMU: &config.QEMUConfig{
						Measurements: testTDXMeasurements,
					},
				},
			},
		},
		"no measurements provided": {
			config: &config.Config{
				AttestationVariant: oid.AzureSEVSNP{}.String(),
				Provider: config.ProviderConfig{
					Azure: &config.AzureConfig{
						Measurements: measurements.M{},
					},
				},
			},
			wantErr: true,
		},
		"unknown variant": {
			config: &config.Config{
				AttestationVariant: "unknown",
				Provider: config.ProviderConfig{
					QEMU: &config.QEMUConfig{
						Measurements: testPCRs,
					},
				},
			},
			wantErr: true,
		},
		"set idkeydigest": {
			config: &config.Config{
				AttestationVariant: oid.AzureSEVSNP{}.String(),
				Provider: config.ProviderConfig{
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
				assert.Equal(tc.config.GetMeasurements(), validators.measurements)
				variant, err := oid.FromString(tc.config.AttestationVariant)
				require.NoError(t, err)
				assert.Equal(variant, validators.attestationVariant)
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
	newTestRTMRs := func() measurements.M {
		return measurements.M{
			0: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			1: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			2: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			3: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			4: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
			5: measurements.WithAllBytes(0x00, true, measurements.TDXMeasurementLength),
		}
	}

	testCases := map[string]struct {
		variant      oid.Getter
		measurements measurements.M
		wantVs       atls.Validator
	}{
		"gcp": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
			wantVs:       gcp.NewValidator(newTestPCRs(), nil),
		},
		"azure cvm": {
			variant:      oid.AzureSEVSNP{},
			measurements: newTestPCRs(),
			wantVs:       snp.NewValidator(newTestPCRs(), idkeydigest.IDKeyDigests{}, false, nil),
		},
		"azure trusted launch": {
			variant:      oid.AzureTrustedLaunch{},
			measurements: newTestPCRs(),
			wantVs:       trustedlaunch.NewValidator(newTestPCRs(), nil),
		},
		"qemu vTPM": {
			variant:      oid.QEMUVTPM{},
			measurements: newTestPCRs(),
			wantVs:       qemu.NewValidator(newTestPCRs(), nil),
		},
		"qemu TDX": {
			variant:      oid.QEMUTDX{},
			measurements: newTestRTMRs(),
			wantVs:       tdx.NewValidator(newTestRTMRs(), nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{attestationVariant: tc.variant, measurements: tc.measurements}

			resultValidator := validators.V(&cobra.Command{})

			assert.Equal(tc.wantVs.OID(), resultValidator.OID())
		})
	}
}

func TestValidatorUpdateInitMeasurements(t *testing.T) {
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
		variant      oid.Getter
		measurements measurements.M
		ownerID      string
		clusterID    string
		wantErr      bool
	}{
		"gcp update owner ID": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
			ownerID:      one64,
		},
		"gcp update cluster ID": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
			clusterID:    one64,
		},
		"gcp update both": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
			ownerID:      one64,
			clusterID:    one64,
		},
		"azure update owner ID": {
			variant:      oid.AzureSEVSNP{},
			measurements: newTestPCRs(),
			ownerID:      one64,
		},
		"azure update cluster ID": {
			variant:      oid.AzureSEVSNP{},
			measurements: newTestPCRs(),
			clusterID:    one64,
		},
		"azure update both": {
			variant:      oid.AzureSEVSNP{},
			measurements: newTestPCRs(),
			ownerID:      one64,
			clusterID:    one64,
		},
		"owner ID and cluster ID empty": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
		},
		"invalid encoding": {
			variant:      oid.GCPSEVES{},
			measurements: newTestPCRs(),
			ownerID:      "invalid",
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{attestationVariant: tc.variant, measurements: tc.measurements}

			err := validators.UpdateInitMeasurements(tc.ownerID, tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			for i := 0; i < len(tc.measurements); i++ {
				switch {
				case i == int(measurements.PCRIndexClusterID) && tc.clusterID == "":
					// should be deleted
					_, ok := validators.measurements[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexClusterID):
					pcr, ok := validators.measurements[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

				case i == int(measurements.PCRIndexOwnerID) && tc.ownerID == "":
					// should be deleted
					_, ok := validators.measurements[uint32(i)]
					assert.False(ok)

				case i == int(measurements.PCRIndexOwnerID):
					pcr, ok := validators.measurements[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr.Expected)

				default:
					if i >= 17 && i <= 22 {
						assert.Equal(one, validators.measurements[uint32(i)])
					} else {
						assert.Equal(zero, validators.measurements[uint32(i)])
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
		variant      oid.Getter
		measurements measurements.M
		clusterID    string
		wantErr      bool
	}{
		"QEMUT TDX update update cluster ID": {
			variant:      oid.QEMUTDX{},
			measurements: newTestTDXMeasurements(),
			clusterID:    one64,
		},
		"cluster ID empty": {
			variant:      oid.QEMUTDX{},
			measurements: newTestTDXMeasurements(),
		},
		"invalid encoding": {
			variant:      oid.QEMUTDX{},
			measurements: newTestTDXMeasurements(),
			clusterID:    "invalid",
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{attestationVariant: tc.variant, measurements: tc.measurements}

			err := validators.UpdateInitMeasurements("", tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			for i := 0; i < len(tc.measurements); i++ {
				switch {
				case i == measurements.TDXIndexClusterID && tc.clusterID == "":
					// should be deleted
					_, ok := validators.measurements[uint32(i)]
					assert.False(ok)

				case i == measurements.TDXIndexClusterID:
					pcr, ok := validators.measurements[uint32(i)]
					assert.True(ok)
					assert.Equal(tdxZeroUpdatedOne[:], pcr.Expected)

				default:
					assert.Equal(zero, validators.measurements[uint32(i)])
				}
			}
		})
	}
}

func TestUpdateMeasurement(t *testing.T) {
	emptyMap := measurements.M{}
	defaultMapPCRs := measurements.M{
		0: measurements.WithAllBytes(0xAA, false, measurements.PCRMeasurementLength),
		1: measurements.WithAllBytes(0xBB, false, measurements.PCRMeasurementLength),
	}
	defaultMapTDXMeasurements := measurements.M{
		0: measurements.WithAllBytes(0xAA, false, measurements.TDXMeasurementLength),
		1: measurements.WithAllBytes(0xBB, false, measurements.TDXMeasurementLength),
	}

	testCases := map[string]struct {
		attestationVariant oid.Getter
		measurementMap     measurements.M
		measurementIndex   uint32
		encoded            string
		wantEntries        int
		wantErr            bool
	}{
		"TPM: empty input, empty map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     emptyMap,
			measurementIndex:   10,
			encoded:            "",
			wantEntries:        0,
			wantErr:            false,
		},
		"TPM: empty input, default map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     defaultMapPCRs,
			measurementIndex:   10,
			encoded:            "",
			wantEntries:        len(defaultMapPCRs),
			wantErr:            false,
		},
		"TPM: correct input, empty map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     emptyMap,
			measurementIndex:   10,
			encoded:            base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries:        1,
			wantErr:            false,
		},
		"TPM: correct input, default map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     defaultMapPCRs,
			measurementIndex:   10,
			encoded:            base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries:        len(defaultMapPCRs) + 1,
			wantErr:            false,
		},
		"TPM: hex input, empty map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     emptyMap,
			measurementIndex:   10,
			encoded:            hex.EncodeToString([]byte("Constellation")),
			wantEntries:        1,
			wantErr:            false,
		},
		"TPM: hex input, default map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     defaultMapPCRs,
			measurementIndex:   10,
			encoded:            hex.EncodeToString([]byte("Constellation")),
			wantEntries:        len(defaultMapPCRs) + 1,
			wantErr:            false,
		},
		"TPM: unencoded input, empty map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     emptyMap,
			measurementIndex:   10,
			encoded:            "Constellation",
			wantEntries:        0,
			wantErr:            true,
		},
		"TPM: unencoded input, default map": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     defaultMapPCRs,
			measurementIndex:   10,
			encoded:            "Constellation",
			wantEntries:        len(defaultMapPCRs),
			wantErr:            true,
		},
		"TPM: empty input at occupied index": {
			attestationVariant: oid.GCPSEVES{},
			measurementMap:     defaultMapPCRs,
			measurementIndex:   0,
			encoded:            "",
			wantEntries:        len(defaultMapPCRs) - 1,
			wantErr:            false,
		},
		"TDX: empty input, empty map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     emptyMap,
			measurementIndex:   3,
			encoded:            "",
			wantEntries:        0,
			wantErr:            false,
		},
		"TDX: empty input, default map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     defaultMapTDXMeasurements,
			measurementIndex:   3,
			encoded:            "",
			wantEntries:        len(defaultMapTDXMeasurements),
			wantErr:            false,
		},
		"TDX: correct input, empty map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     emptyMap,
			measurementIndex:   3,
			encoded:            base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries:        1,
			wantErr:            false,
		},
		"TDX: correct input, default map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     defaultMapTDXMeasurements,
			measurementIndex:   3,
			encoded:            base64.StdEncoding.EncodeToString([]byte("Constellation")),
			wantEntries:        len(defaultMapPCRs) + 1,
			wantErr:            false,
		},
		"TDX: hex input, empty map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     emptyMap,
			measurementIndex:   3,
			encoded:            hex.EncodeToString([]byte("Constellation")),
			wantEntries:        1,
			wantErr:            false,
		},
		"TDX: hex input, default map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     defaultMapTDXMeasurements,
			measurementIndex:   3,
			encoded:            hex.EncodeToString([]byte("Constellation")),
			wantEntries:        len(defaultMapTDXMeasurements) + 1,
			wantErr:            false,
		},
		"TDX: unencoded input, empty map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     emptyMap,
			measurementIndex:   3,
			encoded:            "Constellation",
			wantEntries:        0,
			wantErr:            true,
		},
		"TDX: unencoded input, default map": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     defaultMapTDXMeasurements,
			measurementIndex:   3,
			encoded:            "Constellation",
			wantEntries:        len(defaultMapTDXMeasurements),
			wantErr:            true,
		},
		"TDX: empty input at occupied index": {
			attestationVariant: oid.QEMUTDX{},
			measurementMap:     defaultMapTDXMeasurements,
			measurementIndex:   0,
			encoded:            "",
			wantEntries:        len(defaultMapTDXMeasurements) - 1,
			wantErr:            false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			measurements := make(measurements.M)
			for k, v := range tc.measurementMap {
				measurements[k] = v
			}

			validators := &Validator{
				attestationVariant: tc.attestationVariant,
				measurements:       measurements,
			}
			err := validators.updateMeasurement(tc.measurementIndex, tc.encoded)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.Len(measurements, tc.wantEntries)
		})
	}
}
