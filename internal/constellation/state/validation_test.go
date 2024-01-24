/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreCreateValidation(t *testing.T) {
	testCases := map[string]struct {
		stateFile     func() *State
		wantErr       bool
		errAssertions func(a *assert.Assertions, err error)
	}{
		"valid": {
			stateFile: func() *State {
				return &State{
					Version: Version1,
				}
			},
		},
		"invalid version": {
			stateFile: func() *State {
				return &State{
					Version: "invalid",
				}
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.version: invalid must be one of [v1]")
			},
		},
		"infrastructure not empty": {
			stateFile: func() *State {
				return &State{
					Version: Version1,
					Infrastructure: Infrastructure{
						ClusterEndpoint: "test",
					},
				}
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.clusterEndpoint: test must be empty")
			},
		},
		"cluster values not empty": {
			stateFile: func() *State {
				return &State{
					Version: Version1,
					ClusterValues: ClusterValues{
						ClusterID: "test",
					},
				}
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.clusterValues.clusterID: test must be empty")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.stateFile().Validate(PreCreate, variant.AzureSEVSNP{})
			if tc.wantErr {
				require.Error(t, err)
				if tc.errAssertions != nil {
					tc.errAssertions(assert.New(t), err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPreInitValidation(t *testing.T) {
	validPreInitState := func() *State {
		s := defaultState()
		s.ClusterValues = ClusterValues{}
		return s
	}

	testCases := map[string]struct {
		stateFile     func() *State
		variant       variant.Variant
		wantErr       bool
		errAssertions func(a *assert.Assertions, err error)
	}{
		"valid": {
			stateFile: validPreInitState,
		},
		"invalid version": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Version = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.version: invalid must be one of [v1]")
			},
		},
		"cluster endpoint invalid": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.ClusterEndpoint = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.clusterEndpoint: invalid must be a valid DNS name")
				a.Contains(err.Error(), "validating State.infrastructure.clusterEndpoint: invalid must be a valid IP address")
			},
		},
		"in-cluster endpoint invalid": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.InClusterEndpoint = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.inClusterEndpoint: invalid must be a valid DNS name")
				a.Contains(err.Error(), "validating State.infrastructure.inClusterEndpoint: invalid must be a valid IP address")
			},
		},
		"node ip cidr invalid": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.IPCidrNode = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.ipCidrNode: invalid must be a valid CIDR")
			},
		},
		"uid empty": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.UID = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.uid: must not be empty")
			},
		},
		"name empty": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.Name = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.name: must not be empty")
			},
		},
		"gcp empty": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.GCP = &GCP{}
				return s
			},
			variant: variant.GCPSEVES{},
			wantErr: true,
		},
		"gcp nil": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.GCP = nil
				return s
			},
			variant: variant.GCPSEVES{},
			wantErr: true,
		},
		"gcp invalid": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.GCP.IPCidrPod = "invalid"
				return s
			},
			wantErr: true,
			variant: variant.GCPSEVES{},
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.gcp.ipCidrPod: invalid must be a valid CIDR")
			},
		},
		"azure empty": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.Azure = &Azure{}
				return s
			},
			variant: variant.AzureSEVSNP{},
			wantErr: true,
		},
		"azure nil": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.Azure = nil
				return s
			},
			variant: variant.AzureSEVSNP{},
			wantErr: true,
		},
		"azure invalid": {
			stateFile: func() *State {
				s := validPreInitState()
				s.Infrastructure.Azure.NetworkSecurityGroupName = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.azure.networkSecurityGroupName: must not be empty")
			},
			variant: variant.AzureSEVSNP{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.stateFile().Validate(PreInit, tc.variant)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errAssertions != nil {
					tc.errAssertions(assert.New(t), err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPostInitValidation(t *testing.T) {
	testCases := map[string]struct {
		stateFile     func() *State
		variant       variant.Variant
		wantErr       bool
		errAssertions func(a *assert.Assertions, err error)
	}{
		"valid": {
			stateFile: defaultGCPState,
			variant:   variant.GCPSEVES{},
		},
		"invalid version": {
			stateFile: func() *State {
				s := defaultState()
				s.Version = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.version: invalid must be one of [v1]")
			},
		},
		"cluster endpoint invalid": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.ClusterEndpoint = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.clusterEndpoint: invalid must be a valid DNS name")
				a.Contains(err.Error(), "validating State.infrastructure.clusterEndpoint: invalid must be a valid IP address")
			},
		},
		"in-cluster endpoint invalid": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.InClusterEndpoint = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.inClusterEndpoint: invalid must be a valid DNS name")
				a.Contains(err.Error(), "validating State.infrastructure.inClusterEndpoint: invalid must be a valid IP address")
			},
		},
		"node ip cidr invalid": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.IPCidrNode = "invalid"
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.ipCidrNode: invalid must be a valid CIDR")
			},
		},
		"uid empty": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.UID = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.uid: must not be empty")
			},
		},
		"name empty": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.Name = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.name: must not be empty")
			},
		},
		"gcp valid": {
			stateFile: func() *State {
				s := defaultGCPState()
				return s
			},
			variant: variant.GCPSEVES{},
		},
		"azure valid": {
			stateFile: func() *State {
				s := defaultAzureState()
				return s
			},
			variant: variant.AzureSEVSNP{},
		},
		"azure SEV needs attestation URL": {
			stateFile: func() *State {
				s := defaultAzureState()
				s.Infrastructure.Azure.AttestationURL = ""
				return s
			},
			variant: variant.AzureSEVSNP{},
			wantErr: true,
		},
		"azure TDX does not need attestation URL": {
			stateFile: func() *State {
				s := defaultAzureState()
				s.Infrastructure.Azure.AttestationURL = ""
				return s
			},
			variant: variant.AzureTDX{},
		},
		"gcp, azure not nil": {
			stateFile: func() *State {
				s := defaultState()
				return s
			},
			variant: variant.GCPSEVES{},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "must be equal to <nil>")
				a.Contains(err.Error(), "must be empty")
			},
		},
		"azure, gcp not nil": {
			stateFile: func() *State {
				s := defaultState()
				return s
			},
			variant: variant.AzureSEVSNP{},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "must be equal to <nil>")
				a.Contains(err.Error(), "must be empty")
			},
		},
		"cluster values invalid": {
			stateFile: func() *State {
				s := defaultState()
				s.ClusterValues.ClusterID = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.clusterValues.clusterID: must not be empty")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.stateFile().Validate(PostInit, tc.variant)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errAssertions != nil {
					tc.errAssertions(assert.New(t), err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
