package azure

import (
	"testing"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecrets(t *testing.T) {
	testCases := map[string]struct {
		instance               core.Instance
		cloudServiceAccountURI string
		wantSecrets            resources.Secrets
		wantErr                bool
	}{
		"Secrets works": {
			instance:               core.Instance{ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"},
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			wantSecrets: resources.Secrets{
				&k8s.Secret{
					TypeMeta: meta.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "azureconfig",
						Namespace: "kube-system",
					},
					Data: map[string][]byte{
						"azure.json": []byte(`{"cloud":"AzurePublicCloud","tenantId":"tenant-id","subscriptionId":"subscription-id","resourceGroup":"resource-group","location":"location","useInstanceMetadata":true,"vmType":"standard","aadClientId":"client-id","aadClientSecret":"client-secret"}`),
					},
				},
			},
		},
		"Secrets works for scale sets": {
			instance:               core.Instance{ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id"},
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			wantSecrets: resources.Secrets{
				&k8s.Secret{
					TypeMeta: meta.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "azureconfig",
						Namespace: "kube-system",
					},
					Data: map[string][]byte{
						"azure.json": []byte(`{"cloud":"AzurePublicCloud","tenantId":"tenant-id","subscriptionId":"subscription-id","resourceGroup":"resource-group","location":"location","useInstanceMetadata":true,"vmType":"vmss","aadClientId":"client-id","aadClientSecret":"client-secret"}`),
					},
				},
			},
		},
		"invalid providerID fails": {
			instance: core.Instance{ProviderID: "invalid"},
			wantErr:  true,
		},
		"invalid cloudServiceAccountURI fails": {
			instance:               core.Instance{ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"},
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
	assert.Empty(cloud.ConfigMaps(core.Instance{}))
	assert.NotEmpty(cloud.Volumes())
	assert.NotEmpty(cloud.VolumeMounts())
	assert.Empty(cloud.Env())
	assert.NoError(cloud.PrepareInstance(core.Instance{}, "192.0.2.0"))
	assert.True(cloud.Supported())
}
