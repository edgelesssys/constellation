package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateServiceAccount(t *testing.T) {
	require := require.New(t)
	someErr := errors.New("someErr")
	key := ServiceAccountKey{
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
		errExpected bool
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
			iamAPI:      stubIAMAPI{createErr: someErr},
			errExpected: true,
		},
		"creating account key fails": {
			iamAPI:      stubIAMAPI{createKeyErr: someErr},
			errExpected: true,
		},
		"key data missing": {
			iamAPI:      stubIAMAPI{},
			errExpected: true,
		},
		"key data corrupt": {
			iamAPI:      stubIAMAPI{serviceAccountKeyData: []byte("invalid key data")},
			errExpected: true,
		},
		"retrieving iam policy bindings fails": {
			iamAPI:      stubIAMAPI{},
			projectsAPI: stubProjectsAPI{getPolicyErr: someErr},
			errExpected: true,
		},
		"setting iam policy bindings fails": {
			iamAPI:      stubIAMAPI{},
			projectsAPI: stubProjectsAPI{setPolicyErr: someErr},
			errExpected: true,
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
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(key.ConvertToCloudServiceAccountURI(), serviceAccountKey)
				assert.Equal("email", client.serviceAccount)
			}
		})
	}
}

func TestTerminateServiceAccount(t *testing.T) {
	testCases := map[string]struct {
		iamAPI      iamAPI
		errExpected bool
	}{
		"delete works": {
			iamAPI: stubIAMAPI{},
		},
		"delete fails": {
			iamAPI: stubIAMAPI{
				deleteServiceAccountErr: errors.New("someErr"),
			},
			errExpected: true,
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
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestConvertToCloudServiceAccountURI(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	key := ServiceAccountKey{
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
	cloudServiceAccountURI := key.ConvertToCloudServiceAccountURI()
	uri, err := url.Parse(cloudServiceAccountURI)
	require.NoError(err)
	query := uri.Query()
	assert.Equal("serviceaccount", uri.Scheme)
	assert.Equal("gcp", uri.Host)
	assert.Equal(url.Values{
		"type":                        []string{"type"},
		"project_id":                  []string{"project-id"},
		"private_key_id":              []string{"private-key-id"},
		"private_key":                 []string{"private-key"},
		"client_email":                []string{"client-email"},
		"client_id":                   []string{"client-id"},
		"auth_uri":                    []string{"auth-uri"},
		"token_uri":                   []string{"token-uri"},
		"auth_provider_x509_cert_url": []string{"auth-provider-x509-cert-url"},
		"client_x509_cert_url":        []string{"client-x509-cert-url"},
	}, query)
}
