/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcpshared

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountKeyFromURI(t *testing.T) {
	serviceAccountKey := ServiceAccountKey{
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
		wantKey                ServiceAccountKey
		wantErr                bool
	}{
		"successful": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantKey:                serviceAccountKey,
		},
		"missing type": {
			cloudServiceAccountURI: "serviceaccount://gcp?project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing project_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing private_key_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing private_key": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing client_email": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing client_id": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing token_uri": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing auth_provider_x509_cert_url": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&client_x509_cert_url=client-x509-cert-url",
			wantErr:                true,
		},
		"missing client_x509_cert_url": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url",
			wantErr:                true,
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

			key, err := ServiceAccountKeyFromURI(tc.cloudServiceAccountURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantKey, key)
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
	cloudServiceAccountURI := key.ToCloudServiceAccountURI()
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
