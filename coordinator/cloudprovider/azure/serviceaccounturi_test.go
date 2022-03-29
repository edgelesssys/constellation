package azure

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/azure/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetApplicationCredentials(t *testing.T) {
	creds := client.ApplicationCredentials{
		TenantID:     "tenant-id",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Location:     "location",
	}
	testCases := map[string]struct {
		cloudServiceAccountURI string
		expectedCreds          client.ApplicationCredentials
		expectErr              bool
	}{
		"getApplicationCredentials works": {
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			expectedCreds:          creds,
		},
		"invalid URI fails": {
			cloudServiceAccountURI: "\x00",
			expectErr:              true,
		},
		"incorrect URI scheme fails": {
			cloudServiceAccountURI: "invalid",
			expectErr:              true,
		},
		"incorrect URI host fails": {
			cloudServiceAccountURI: "serviceaccount://incorrect",
			expectErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			creds, err := getApplicationCredentials(tc.cloudServiceAccountURI)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedCreds, creds)
		})
	}
}
