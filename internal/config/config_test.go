package config

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	def := Default()
	assert.NotNil(def)
}

func TestFromFile(t *testing.T) {
	someProviderConfig := &ProviderConfig{
		GCP: &GCPConfig{
			FirewallInput: &client.FirewallInput{
				Ingress: cloudtypes.Firewall{
					{
						Name:        "firstFirewallRule",
						Description: "firstFirewallRule description",
						Protocol:    "tcp",
						Port:        4444,
					},
					{
						Name:        "secondFirewallRule",
						Description: "secondFirewallRule description",
						Protocol:    "udp",
						Port:        5555,
					},
				},
			},
		},
	}

	configName := "config.json"

	testCases := map[string]struct {
		from              *Config
		configName        string
		wantResultMutator func(c *Config) //  mutates the Default() config to the expected result.
		wantErr           bool
	}{
		"overwrite fields": {
			from:              &Config{CoordinatorPort: proto.String("1000")},
			configName:        configName,
			wantResultMutator: func(c *Config) { c.CoordinatorPort = proto.String("1000") },
		},
		"overwrite slices": {
			from:              &Config{Provider: someProviderConfig},
			configName:        configName,
			wantResultMutator: func(c *Config) { c.Provider = someProviderConfig },
		},
		"default with empty name": {
			from:              &Config{},
			configName:        "",
			wantResultMutator: func(c *Config) {},
		},
		"err with wrong name": {
			from:              &Config{},
			configName:        "wrongName.json",
			wantResultMutator: func(c *Config) {},
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			require.NoError(fileHandler.WriteJSON(configName, tc.from, file.OptNone))

			result, err := FromFile(fileHandler, tc.configName)

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				wantResult := Default()
				tc.wantResultMutator(wantResult)
				assert.EqualValues(wantResult.CoordinatorPort, result.CoordinatorPort)
				assert.EqualValues(wantResult.AutoscalingNodeGroupsMin, result.AutoscalingNodeGroupsMin)
				assert.EqualValues(wantResult.AutoscalingNodeGroupsMax, result.AutoscalingNodeGroupsMax)
				require.NotNil(wantResult.Provider)
				require.NotNil(wantResult.Provider.GCP, result.Provider.GCP)
				require.NotNil(wantResult.Provider.GCP.FirewallInput, result.Provider.GCP.FirewallInput)
				assert.Equal(len(wantResult.Provider.GCP.FirewallInput.Ingress), len(result.Provider.GCP.FirewallInput.Ingress))
			}
		})
	}
}
