package azureshared

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestApplicationCredentialsFromURI(t *testing.T) {
	creds := ApplicationCredentials{
		TenantID:          "tenant-id",
		AppClientID:       "client-id",
		ClientSecretValue: "client-secret",
		Location:          "location",
	}
	testCases := map[string]struct {
		cloudServiceAccountURI string
		wantCreds              ApplicationCredentials
		wantErr                bool
	}{
		"getApplicationCredentials works": {
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			wantCreds:              creds,
		},
		"invalid URI fails": {
			cloudServiceAccountURI: "\x00",
			wantErr:                true,
		},
		"incorrect URI scheme fails": {
			cloudServiceAccountURI: "invalid",
			wantErr:                true,
		},
		"incorrect URI host fails": {
			cloudServiceAccountURI: "serviceaccount://incorrect",
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			creds, err := ApplicationCredentialsFromURI(tc.cloudServiceAccountURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantCreds, creds)
		})
	}
}

func TestToCloudServiceAccountURI(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	key := ApplicationCredentials{
		TenantID:          "tenant-id",
		AppClientID:       "client-id",
		ClientSecretValue: "client-secret",
		Location:          "location",
	}

	cloudServiceAccountURI := key.ToCloudServiceAccountURI()
	uri, err := url.Parse(cloudServiceAccountURI)
	require.NoError(err)
	query := uri.Query()
	assert.Equal("serviceaccount", uri.Scheme)
	assert.Equal("azure", uri.Host)
	assert.Equal(url.Values{
		"tenant_id":     []string{"tenant-id"},
		"client_id":     []string{"client-id"},
		"client_secret": []string{"client-secret"},
		"location":      []string{"location"},
	}, query)
}
