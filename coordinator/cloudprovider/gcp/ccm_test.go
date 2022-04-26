package gcp

import (
	"encoding/json"
	"testing"

	"github.com/edgelesssys/constellation/cli/gcp/client"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigMaps(t *testing.T) {
	testCases := map[string]struct {
		instance       core.Instance
		wantConfigMaps resources.ConfigMaps
		wantErr        bool
	}{
		"ConfigMaps works": {
			instance: core.Instance{ProviderID: "gce://project-id/zone/instance-name"},
			wantConfigMaps: resources.ConfigMaps{
				&k8s.ConfigMap{
					TypeMeta: v1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "gceconf",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"gce.conf": `[global]
project-id = project-id
use-metadata-server = false
`,
					},
				},
			},
		},
		"invalid providerID fails": {
			instance: core.Instance{ProviderID: "invalid"},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cloud := CloudControllerManager{}
			configMaps, err := cloud.ConfigMaps(tc.instance)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantConfigMaps, configMaps)
		})
	}
}

func TestSecrets(t *testing.T) {
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
	rawKey, err := json.Marshal(serviceAccountKey)
	require.NoError(t, err)
	testCases := map[string]struct {
		instance               core.Instance
		cloudServiceAccountURI string
		wantSecrets            resources.Secrets
		wantErr                bool
	}{
		"Secrets works": {
			cloudServiceAccountURI: "serviceaccount://gcp?type=type&project_id=project-id&private_key_id=private-key-id&private_key=private-key&client_email=client-email&client_id=client-id&auth_uri=auth-uri&token_uri=token-uri&auth_provider_x509_cert_url=auth-provider-x509-cert-url&client_x509_cert_url=client-x509-cert-url",
			wantSecrets: resources.Secrets{
				&k8s.Secret{
					TypeMeta: v1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: v1.ObjectMeta{
						Name:      "gcekey",
						Namespace: "kube-system",
					},
					Data: map[string][]byte{
						"key.json": rawKey,
					},
				},
			},
		},
		"invalid serviceAccountKey fails": {
			cloudServiceAccountURI: "invalid",
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cloud := CloudControllerManager{}
			secrets, err := cloud.Secrets(tc.instance, tc.cloudServiceAccountURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantSecrets, secrets)
		})
	}
}

func TestTrivialCCMFunctions(t *testing.T) {
	assert := assert.New(t)
	cloud := CloudControllerManager{}

	assert.NotEmpty(cloud.Image())
	assert.NotEmpty(cloud.Path())
	assert.NotEmpty(cloud.Name())
	assert.NotEmpty(cloud.ExtraArgs())
	assert.NotEmpty(cloud.Volumes())
	assert.NotEmpty(cloud.VolumeMounts())
	assert.NotEmpty(cloud.Env())
	assert.NoError(cloud.PrepareInstance(core.Instance{}, "192.0.2.0"))
	assert.True(cloud.Supported())
}
