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

func TestAutoscalerSecrets(t *testing.T) {
	testCases := map[string]struct {
		instance               core.Instance
		cloudServiceAccountURI string
		expectedSecrets        resources.Secrets
		expectErr              bool
	}{
		"Secrets works": {
			instance:               core.Instance{ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"},
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret",
			expectedSecrets: resources.Secrets{
				&k8s.Secret{
					TypeMeta: meta.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: meta.ObjectMeta{
						Name:      "cluster-autoscaler-azure",
						Namespace: "kube-system",
					},
					Data: map[string][]byte{
						"ClientID":       []byte("client-id"),
						"ClientSecret":   []byte("client-secret"),
						"ResourceGroup":  []byte("resource-group"),
						"SubscriptionID": []byte("subscription-id"),
						"TenantID":       []byte("tenant-id"),
						"VMType":         []byte("vmss"),
					},
				},
			},
		},
		"invalid providerID fails": {
			instance:  core.Instance{ProviderID: "invalid"},
			expectErr: true,
		},
		"invalid cloudServiceAccountURI fails": {
			instance:               core.Instance{ProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name"},
			cloudServiceAccountURI: "invalid",
			expectErr:              true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			autoscaler := Autoscaler{}
			secrets, err := autoscaler.Secrets(tc.instance, tc.cloudServiceAccountURI)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedSecrets, secrets)
		})
	}
}

func TestTrivialAutoscalerFunctions(t *testing.T) {
	assert := assert.New(t)
	autoscaler := Autoscaler{}

	assert.NotEmpty(autoscaler.Name())
	assert.Empty(autoscaler.Volumes())
	assert.Empty(autoscaler.VolumeMounts())
	assert.NotEmpty(autoscaler.Env())
	assert.True(autoscaler.Supported())
}
