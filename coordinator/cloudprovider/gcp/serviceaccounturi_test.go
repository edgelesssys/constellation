package gcp

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetServiceAccountKey(t *testing.T) {
	serviceAccountKey := client.ServiceAccountKey{
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
	testCases := map[string]struct {
		cloudServiceAccountURI string
		expectedKey            client.ServiceAccountKey
		expectErr              bool
	}{
		"getServiceAccountKey works": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectedKey:            serviceAccountKey,
		},
		"missing type": {
			cloudServiceAccountURI: "serviceaccount://gcp?project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing project_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing private_key_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing private_key": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing client_email": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing client_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing token_uri": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing auth_provider_x509_cert_url": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&client_x509_cert_url=client-x509-cert-url",
			expectErr:              true,
		},
		"missing client_x509_cert_url": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url",
			expectErr:              true,
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

			key, err := getServiceAccountKey(tc.cloudServiceAccountURI)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedKey, key)
		})
	}
}
