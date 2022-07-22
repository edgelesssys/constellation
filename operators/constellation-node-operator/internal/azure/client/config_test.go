package client

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	testCases := map[string]struct {
		rawConfig          string
		skipWrite          bool
		wantTenantID       string
		wantSubscriptionID string
		wantErr            bool
	}{
		"valid config": {
			rawConfig:          `{"tenantId":"tenantId","subscriptionId":"subscriptionId"}`,
			wantTenantID:       "tenantId",
			wantSubscriptionID: "subscriptionId",
		},
		"invalid config": {
			rawConfig: `{"tenantId":"tenantId","subscriptionId":""}`,
			wantErr:   true,
		},
		"config is empty": {
			wantErr: true,
		},
		"config does not exist": {
			skipWrite: true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.NewMemMapFs()
			if !tc.skipWrite {
				require.NoError(afero.WriteFile(fs, "config.json", []byte(tc.rawConfig), 0o644))
			}
			gotConfig, err := loadConfig(fs, "config.json")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantTenantID, gotConfig.TenantID)
			assert.Equal(tc.wantSubscriptionID, gotConfig.SubscriptionID)
		})
	}
}
