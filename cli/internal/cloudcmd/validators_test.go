package cloudcmd

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/attestation/azure/snp"
	"github.com/edgelesssys/constellation/internal/attestation/azure/trustedlaunch"
	"github.com/edgelesssys/constellation/internal/attestation/gcp"
	"github.com/edgelesssys/constellation/internal/attestation/qemu"
	"github.com/edgelesssys/constellation/internal/attestation/vtpm"
	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
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
		provider           cloudprovider.Provider
		config             *config.Config
		pcrs               map[uint32][]byte
		enforceIdKeyDigest bool
		idkeydigest        string
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
		"set idkeydigest": {
			provider:           cloudprovider.Azure,
			pcrs:               testPCRs,
			idkeydigest:        "414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141",
			enforceIdKeyDigest: true,
		},
		"invalid idkeydigest": {
			provider:           cloudprovider.Azure,
			pcrs:               testPCRs,
			idkeydigest:        "41414141414141414141414141414141414141414141414141414141414141414141414141414141414141414141414",
			enforceIdKeyDigest: true,
			azureCVM:           true,
			wantErr:            true,
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
				conf.Provider.Azure = &config.AzureConfig{Measurements: measurements, EnforceIdKeyDigest: &tc.enforceIdKeyDigest, IdKeyDigest: tc.idkeydigest, ConfidentialVM: &tc.azureCVM}
			}
			if tc.provider == cloudprovider.QEMU {
				measurements := config.Measurements(tc.pcrs)
				conf.Provider.QEMU = &config.QEMUConfig{Measurements: measurements}
			}

			validators, err := NewValidator(tc.provider, conf)

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
		wantVs   atls.Validator
	}{
		"gcp": {
			provider: cloudprovider.GCP,
			pcrs:     newTestPCRs(),
			wantVs:   gcp.NewValidator(newTestPCRs(), nil, nil),
		},
		"azure cvm": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			wantVs:   snp.NewValidator(newTestPCRs(), nil, nil, false, nil),
		},
		"azure trusted launch": {
			provider: cloudprovider.Azure,
			pcrs:     newTestPCRs(),
			wantVs:   trustedlaunch.NewValidator(newTestPCRs(), nil, nil),
		},
		"qemu": {
			provider: cloudprovider.QEMU,
			pcrs:     newTestPCRs(),
			wantVs:   qemu.NewValidator(newTestPCRs(), nil, nil),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			validators := &Validator{provider: tc.provider, pcrs: tc.pcrs}

			resultValidator := validators.V(&cobra.Command{})

			assert.Equal(tc.wantVs.OID(), resultValidator.OID())
		})
	}
}

func TestValidatorUpdateInitPCRs(t *testing.T) {
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

			validators := &Validator{provider: tc.provider, pcrs: tc.pcrs}

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
			for _, v := range pcrs {
				assert.Len(v, 32)
			}
		})
	}
}
