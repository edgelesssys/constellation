package cloudcmd

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestWarnAboutPCRs(t *testing.T) {
	zero := []byte("00000000000000000000000000000000")

	testCases := map[string]struct {
		pcrs         map[uint32][]byte
		wantWarnings []string
		wantWInclude []string
		wantErr      bool
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
		"bad config": {
			pcrs: map[uint32][]byte{
				0: []byte("000"),
			},
			wantErr: true,
		},
	}

	for _, provider := range []string{"gcp", "azure", "unknown"} {
		t.Run(provider, func(t *testing.T) {
			for name, tc := range testCases {
				t.Run(name, func(t *testing.T) {
					assert := assert.New(t)

					config := &config.Config{
						Provider: &config.ProviderConfig{
							Azure: &config.AzureConfig{PCRs: &tc.pcrs},
							GCP:   &config.GCPConfig{PCRs: &tc.pcrs},
						},
					}

					validators, err := NewValidators(cloudprovider.FromString(provider), config)

					v := validators.V()
					warnings := validators.Warnings()
					warningsInclueInit := validators.WarningsIncludeInit()

					if tc.wantErr || provider == "unknown" {
						assert.Error(err)
					} else {
						assert.NoError(err)
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
						assert.NotEmpty(v)
					}
				})
			}
		})
	}
}
