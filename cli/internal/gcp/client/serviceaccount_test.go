package client

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/gcpshared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateServiceAccount(t *testing.T) {
	require := require.New(t)
	someErr := errors.New("someErr")
	key := gcpshared.ServiceAccountKey{
		Type:                    "type",
		ProjectID:               "project-id",
		PrivateKeyID:            "private-key-id",
		PrivateKey:              "private-key",
		ClientEmail:             "client-email",
		ClientID:                "client-id",
		AuthURI:                 "auth-uri",
		TokenURI:                "token-uri",
		AuthProviderX509CertURL: "auth-provider-x509-cert-url",
		ClientX509CertURL:       "client-x509-cert-url",
	}
	keyData, err := json.Marshal(key)
	require.NoError(err)

	testCases := map[string]struct {
		iamAPI      iamAPI
		projectsAPI stubProjectsAPI
		input       ServiceAccountInput
		wantErr     bool
	}{
		"successful create": {
			iamAPI: stubIAMAPI{serviceAccountKeyData: keyData},
			input: ServiceAccountInput{
				Roles: []string{"someRole"},
			},
		},
		"successful create with roles": {
			iamAPI: stubIAMAPI{serviceAccountKeyData: keyData},
		},
		"creating account fails": {
			iamAPI:  stubIAMAPI{createErr: someErr},
			wantErr: true,
		},
		"creating account key fails": {
			iamAPI:  stubIAMAPI{createKeyErr: someErr},
			wantErr: true,
		},
		"key data missing": {
			iamAPI:  stubIAMAPI{},
			wantErr: true,
		},
		"key data corrupt": {
			iamAPI:  stubIAMAPI{serviceAccountKeyData: []byte("invalid key data")},
			wantErr: true,
		},
		"retrieving iam policy bindings fails": {
			iamAPI:      stubIAMAPI{},
			projectsAPI: stubProjectsAPI{getPolicyErr: someErr},
			wantErr:     true,
		},
		"setting iam policy bindings fails": {
			iamAPI:      stubIAMAPI{},
			projectsAPI: stubProjectsAPI{setPolicyErr: someErr},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:     "project",
				zone:        "zone",
				name:        "name",
				uid:         "uid",
				iamAPI:      tc.iamAPI,
				projectsAPI: tc.projectsAPI,
			}

			serviceAccountKey, err := client.CreateServiceAccount(ctx, tc.input)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(key.ToCloudServiceAccountURI(), serviceAccountKey)
				assert.Equal("email", client.serviceAccount)
			}
		})
	}
}

func TestTerminateServiceAccount(t *testing.T) {
	testCases := map[string]struct {
		iamAPI  iamAPI
		wantErr bool
	}{
		"delete works": {
			iamAPI: stubIAMAPI{},
		},
		"delete fails": {
			iamAPI: stubIAMAPI{
				deleteServiceAccountErr: errors.New("someErr"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:        "project",
				zone:           "zone",
				name:           "name",
				uid:            "uid",
				serviceAccount: "service-account",
				iamAPI:         tc.iamAPI,
			}

			err := client.TerminateServiceAccount(ctx)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
