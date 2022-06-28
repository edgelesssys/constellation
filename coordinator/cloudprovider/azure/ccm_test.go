package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecrets(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		providerID             string
		metadata               ccmMetadata
		cloudServiceAccountURI string
		wantSecrets            resources.Secrets
		wantErr                bool
	}{
		"Secrets works": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			metadata:               &ccmMetadataStub{loadBalancerName: "load-balancer-name", networkSecurityGroupName: "network-security-group-name"},
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
						"azure.json": []byte(`{"cloud":"AzurePublicCloud","tenantId":"tenant-id","subscriptionId":"subscription-id","resourceGroup":"resource-group","location":"location","securityGroupName":"network-security-group-name","loadBalancerName":"load-balancer-name","loadBalancerSku":"standard","useInstanceMetadata":true,"vmType":"standard","aadClientId":"client-id","aadClientSecret":"client-secret"}`),
					},
				},
			},
		},
		"Secrets works for scale sets": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			metadata:               &ccmMetadataStub{loadBalancerName: "load-balancer-name", networkSecurityGroupName: "network-security-group-name"},
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
						"azure.json": []byte(`{"cloud":"AzurePublicCloud","tenantId":"tenant-id","subscriptionId":"subscription-id","resourceGroup":"resource-group","location":"location","securityGroupName":"network-security-group-name","loadBalancerName":"load-balancer-name","loadBalancerSku":"standard","useInstanceMetadata":true,"vmType":"vmss","aadClientId":"client-id","aadClientSecret":"client-secret"}`),
					},
				},
			},
		},
		"cannot get load balancer Name": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			metadata:               &ccmMetadataStub{getLoadBalancerNameErr: someErr},
			wantErr:                true,
		},
		"cannot get network security group name": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			cloudServiceAccountURI: "serviceaccount://azure?tenant_id=tenant-id&client_id=client-id&client_secret=client-secret&location=location",
			metadata:               &ccmMetadataStub{getNetworkSecurityGroupNameErr: someErr},
			wantErr:                true,
		},
		"invalid providerID fails": {
			providerID: "invalid",
			metadata:   &ccmMetadataStub{},
			wantErr:    true,
		},
		"invalid cloudServiceAccountURI fails": {
			providerID:             "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			metadata:               &ccmMetadataStub{},
			cloudServiceAccountURI: "invalid",
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cloud := NewCloudControllerManager(tc.metadata)
			secrets, err := cloud.Secrets(context.Background(), tc.providerID, tc.cloudServiceAccountURI)
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
	assert.Empty(cloud.ConfigMaps(metadata.InstanceMetadata{}))
	assert.NotEmpty(cloud.Volumes())
	assert.NotEmpty(cloud.VolumeMounts())
	assert.Empty(cloud.Env())
	assert.True(cloud.Supported())
}

type ccmMetadataStub struct {
	networkSecurityGroupName string
	loadBalancerName         string

	getNetworkSecurityGroupNameErr error
	getLoadBalancerNameErr         error
}

func (c *ccmMetadataStub) GetNetworkSecurityGroupName(ctx context.Context) (string, error) {
	return c.networkSecurityGroupName, c.getNetworkSecurityGroupNameErr
}

func (c *ccmMetadataStub) GetLoadBalancerName(ctx context.Context) (string, error) {
	return c.loadBalancerName, c.getLoadBalancerNameErr
}
