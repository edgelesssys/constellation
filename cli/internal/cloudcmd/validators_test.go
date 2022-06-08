package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/attestation/qemu"
	"github.com/edgelesssys/constellation/coordinator/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")
	one := []byte("11111111111111111111111111111111")
	testPCRs := map[uint32][]byte{
		0: zero,
		1: one,
		2: zero,
		3: one,
		4: zero,
		5: zero,
	}

	testCases := map[string]struct {
		provider cloudprovider.Provider
		config   *config.Config
		pcrs     map[uint32][]byte
		wantErr  bool
	}{
		"gcp": {
			provider: cloudprovider.GCP,
			pcrs:     testPCRs,
		},
		"azure": {
			provider: cloudprovider.Azure,
			pcrs:     testPCRs,
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			pcrs:     testPCRs,
		},
		"no pcrs provided": {
			provider: cloudprovider.Azure,
			pcrs:     map[uint32][]byte{},
			wantErr:  true,
		},
		"invalid pcr length": {
			provider: cloudprovider.GCP,
			pcrs:     map[uint32][]byte{0: []byte("0000000000000000000000000000000")},
			wantErr:  true,
		},
		"unknown provider": {
			provider: cloudprovider.Unknown,
			pcrs:     testPCRs,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			conf := &config.Config{Provider: config.ProviderConfig{}}
			if tc.provider == cloudprovider.GCP {
				measurements := config.Measurements(tc.pcrs)
				conf.Provider.GCP = &config.GCPConfig{Measurements: measurements}
			}
			if tc.provider == cloudprovider.Azure {
				measurements := config.Measurements(tc.pcrs)
				conf.Provider.Azure = &config.AzureConfig{Measurements: measurements}
			}
			if tc.provider == cloudprovider.QEMU {
				measurements := config.Measurements(tc.pcrs)
				conf.Provider.QEMU = &config.QEMUConfig{Measurements: measurements}
			}

			validators, err := NewValidators(tc.provider, conf)

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

func TestValidatorsWarnings(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")

	testCases := map[string]struct {
		pcrs         map[uint32][]byte
		wantWarnings []string
		wantWInclude []string
	}{
		"no warnings": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				6:  zero,
				7:  zero,
				8:  zero,
				9:  zero,
				10: zero,
				11: zero,
				12: zero,
			},
		},
		"no warnings for missing non critical values": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
		},
		"warn for BIOS": {
			pcrs: map[uint32][]byte{
				0:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"BIOS"},
		},
		"warn for OPROM": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"OPROM"},
		},
		"warn for MBR": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"MBR"},
		},
		"warn for kernel": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				9:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"kernel"},
		},
		"warn for initrd": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				11: zero,
				12: zero,
			},
			wantWarnings: []string{"initrd"},
		},
		"warn for initialization": {
			pcrs: map[uint32][]byte{
				0:  zero,
				1:  zero,
				2:  zero,
				3:  zero,
				4:  zero,
				5:  zero,
				8:  zero,
				9:  zero,
				11: zero,
			},
			wantWInclude: []string{"initialization"},
		},
		"multi warning": {
			pcrs: map[uint32][]byte{},
			wantWarnings: []string{
				"BIOS",
				"OPROM",
				"MBR",
				"initrd",
				"kernel",
			},
			wantWInclude: []string{"initialization"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := Validators{pcrs: tc.pcrs}

			warnings := validators.Warnings()
			warningsInclueInit := validators.WarningsIncludeInit()

			if len(tc.wantWarnings) == 0 {
				assert.Empty(warnings)
			}
			for _, w := range tc.wantWarnings {
				assert.Contains(warnings, w)
			}
			for _, w := range tc.wantWarnings {
				assert.Contains(warningsInclueInit, w)
			}
			if len(tc.wantWInclude) == 0 {
				assert.Equal(len(warnings), len(warningsInclueInit))
			} else {
				assert.Greater(len(warningsInclueInit), len(warnings))
			}
			for _, w := range tc.wantWInclude {
				assert.Contains(warningsInclueInit, w)
			}
		})
	}
}

func TestValidatorsV(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")
	newTestPCRs := func() map[uint32][]byte {
		return map[uint32][]byte{
			0:  zero,
			1:  zero,
			2:  zero,
			3:  zero,
			4:  zero,
			5:  zero,
			6:  zero,
			7:  zero,
			8:  zero,
			9:  zero,
			10: zero,
			11: zero,
			12: zero,
		}
	}

	testCases := map[string]struct {
		provider cloudprovider.Provider
		pcrs     map[uint32][]byte
		wantVs   []atls.Validator
	}{
		"gcp": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
			wantVs: []atls.Validator{
				gcp.NewValidator(newTestPCRs()),
			},
		},
		"azure": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			wantVs: []atls.Validator{
				azure.NewValidator(newTestPCRs()),
			},
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			pcrs:     newTestPCRs(),
			wantVs: []atls.Validator{
				qemu.NewValidator(newTestPCRs()),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validators{provider: tc.provider, pcrs: tc.pcrs}

			resultValidators := validators.V()

			assert.Equal(len(tc.wantVs), len(resultValidators))
			for i, resValidator := range resultValidators {
				assert.Equal(tc.wantVs[i].OID(), resValidator.OID())
			}
		})
	}
}

func TestValidatorsUpdateInitPCRs(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")
	one := []byte("11111111111111111111111111111111")
	one64 := base64.StdEncoding.EncodeToString(one)
	oneHash := sha256.Sum256(one)
	pcrZeroUpdatedOne := sha256.Sum256(append(zero, oneHash[:]...))
	newTestPCRs := func() map[uint32][]byte {
		return map[uint32][]byte{
			0:  zero,
			1:  zero,
			2:  zero,
			3:  zero,
			4:  zero,
			5:  zero,
			6:  zero,
			7:  zero,
			8:  zero,
			9:  zero,
			10: zero,
			11: zero,
			12: zero,
		}
	}

	testCases := map[string]struct {
		provider  cloudprovider.Provider
		pcrs      map[uint32][]byte
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

			validators := &Validators{provider: tc.provider, pcrs: tc.pcrs}

			err := validators.UpdateInitPCRs(tc.ownerID, tc.clusterID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			for i := 0; i < len(tc.pcrs); i++ {
				switch {
				case i == int(vtpm.PCRIndexClusterID) && tc.clusterID == "":
					// should be deleted
					_, ok := validators.pcrs[uint32(i)]
					assert.False(ok)

				case i == int(vtpm.PCRIndexClusterID):
					pcr, ok := validators.pcrs[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr)

				case i == int(vtpm.PCRIndexOwnerID) && tc.ownerID == "":
					// should be deleted
					_, ok := validators.pcrs[uint32(i)]
					assert.False(ok)

				case i == int(vtpm.PCRIndexOwnerID):
					pcr, ok := validators.pcrs[uint32(i)]
					assert.True(ok)
					assert.Equal(pcrZeroUpdatedOne[:], pcr)

				default:
					assert.Equal(zero, validators.pcrs[uint32(i)])
				}
			}
		})
	}
}

func TestUpdatePCR(t *testing.T) {
	emptyMap := map[uint32][]byte{}
	defaultMap := map[uint32][]byte{
		0: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		1: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
	}

	testCases := map[string]struct {
		pcrMap      map[uint32][]byte
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

			pcrs := make(map[uint32][]byte)
			for k, v := range tc.pcrMap {
				pcrs[k] = v
			}

			validators := &Validators{
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
			for _, v := range pcrs {
				assert.Len(v, 32)
			}
		})
	}
}
