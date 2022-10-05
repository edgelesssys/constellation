/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestCreator(t *testing.T) {
	failOnNonAMD64 := (runtime.GOARCH != "amd64") || (runtime.GOOS != "linux")

	wantGCPState := state.ConstellationState{
		CloudProvider:  cloudprovider.GCP.String(),
		LoadBalancerIP: "192.0.2.1",
	}

	wantQEMUState := state.ConstellationState{
		CloudProvider:  cloudprovider.QEMU.String(),
		LoadBalancerIP: "192.0.2.1",
	}

	wantAzureState := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
		AzureControlPlaneInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureWorkerInstances: cloudtypes.Instances{
			"id-0": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-1": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
			"id-2": {PrivateIP: "192.0.2.1", PublicIP: "192.0.2.1"},
		},
		AzureSubnet:               "subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureWorkerScaleSet:       "workers-scale-set",
		AzureControlPlaneScaleSet: "controlplanes-scale-set",
	}

	someErr := errors.New("failed")

	testCases := map[string]struct {
		tfClient          terraformClient
		newTfClientErr    error
		azureclient       azureclient
		newAzureClientErr error
		libvirt           *stubLibvirtRunner
		provider          cloudprovider.Provider
		config            *config.Config
		wantState         state.ConstellationState
		wantErr           bool
		wantRollback      bool // Use only together with stubClients.
	}{
		"gcp": {
			tfClient:  &stubTerraformClient{state: wantGCPState},
			provider:  cloudprovider.GCP,
			config:    config.Default(),
			wantState: wantGCPState,
		},
		"gcp newTerraformClient error": {
			newTfClientErr: someErr,
			provider:       cloudprovider.GCP,
			config:         config.Default(),
			wantErr:        true,
		},
		"gcp create cluster error": {
			tfClient:     &stubTerraformClient{createClusterErr: someErr},
			provider:     cloudprovider.GCP,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"qemu": {
			tfClient:  &stubTerraformClient{state: wantQEMUState},
			libvirt:   &stubLibvirtRunner{},
			provider:  cloudprovider.QEMU,
			config:    config.Default(),
			wantState: wantQEMUState,
			wantErr:   failOnNonAMD64,
		},
		"qemu newTerraformClient error": {
			newTfClientErr: someErr,
			libvirt:        &stubLibvirtRunner{},
			provider:       cloudprovider.QEMU,
			config:         config.Default(),
			wantErr:        true,
		},
		"qemu create cluster error": {
			tfClient:     &stubTerraformClient{createClusterErr: someErr},
			libvirt:      &stubLibvirtRunner{},
			provider:     cloudprovider.QEMU,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: !failOnNonAMD64, // if we run on non-AMD64/linux, we don't get to a point where rollback is needed
		},
		"qemu start libvirt error": {
			tfClient:     &stubTerraformClient{state: wantQEMUState},
			libvirt:      &stubLibvirtRunner{startErr: someErr},
			provider:     cloudprovider.QEMU,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: !failOnNonAMD64,
		},
		"azure": {
			azureclient: &fakeAzureClient{},
			provider:    cloudprovider.Azure,
			config:      config.Default(),
			wantState:   wantAzureState,
		},
		"azure newAzureClient error": {
			newAzureClientErr: someErr,
			provider:          cloudprovider.Azure,
			config:            config.Default(),
			wantErr:           true,
		},
		"azure CreateVirtualNetwork error": {
			azureclient:  &stubAzureClient{createVirtualNetworkErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"azure CreateSecurityGroup error": {
			azureclient:  &stubAzureClient{createSecurityGroupErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"azure CreateInstances error": {
			azureclient:  &stubAzureClient{createInstancesErr: someErr},
			provider:     cloudprovider.Azure,
			config:       config.Default(),
			wantErr:      true,
			wantRollback: true,
		},
		"unknown provider": {
			provider: cloudprovider.Unknown,
			config:   config.Default(),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			creator := &Creator{
				out: &bytes.Buffer{},
				newTerraformClient: func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
				newAzureClient: func(subscriptionID, tenantID, name, location, resourceGroup string) (azureclient, error) {
					return tc.azureclient, tc.newAzureClientErr
				},
				newLibvirtRunner: func() libvirtRunner {
					return tc.libvirt
				},
			}

			state, err := creator.Create(context.Background(), tc.provider, tc.config, "name", "type", 2, 3)

			if tc.wantErr {
				assert.Error(err)
				if tc.wantRollback {
					switch tc.provider {
					case cloudprovider.QEMU:
						cl := tc.tfClient.(*stubTerraformClient)
						assert.True(cl.destroyClusterCalled)
						assert.True(cl.cleanUpWorkspaceCalled)
						assert.True(tc.libvirt.stopCalled)
					case cloudprovider.GCP:
						cl := tc.tfClient.(*stubTerraformClient)
						assert.True(cl.destroyClusterCalled)
						assert.True(cl.cleanUpWorkspaceCalled)
					case cloudprovider.Azure:
						cl := tc.azureclient.(*stubAzureClient)
						assert.True(cl.terminateResourceGroupResourcesCalled)
					}
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantState, state)
			}
		})
	}
}
