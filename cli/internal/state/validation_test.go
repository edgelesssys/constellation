/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package state

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation(t *testing.T) {
	testCases := map[string]struct {
		stateFile     func() *State
		wantErr       bool
		errAssertions func(a *assert.Assertions, err error)
	}{
		"valid": {
			stateFile: defaultState,
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
		"empty infrastructure": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure = Infrastructure{}
				return s
			},
		},
		"invalid cluster endpoint": {
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
		"invalid in-cluster endpoint": {
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
		"empty sans": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.APIServerCertSANs = []string{}
				return s
			},
		},
		"not all sans are valid": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.APIServerCertSANs = []string{"not valid!"}
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.apiServerCertSANs[0]: not valid! must be a valid DNS name")
				a.Contains(err.Error(), "validating State.infrastructure.apiServerCertSANs[0]: not valid! must be a valid IP address")
			},
		},
		"invalid node ip cidr": {
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
		"empty cluster values": {
			stateFile: func() *State {
				s := defaultState()
				s.ClusterValues = ClusterValues{}
				return s
			},
		},
		"only one value filled in cluster values": {
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
		"gcp nil": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.GCP = nil
				return s
			},
		},
		"gcp empty": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.GCP = &GCP{}
				return s
			},
		},
		"only one value filled in gcp": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.GCP.ProjectID = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.gcp.projectID: must not be empty")
			},
		},
		"azure nil": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.Azure = nil
				return s
			},
		},
		"azure empty": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.Azure = &Azure{}
				return s
			},
		},
		"only one value filled in azure": {
			stateFile: func() *State {
				s := defaultState()
				s.Infrastructure.Azure.NetworkSecurityGroupName = ""
				return s
			},
			wantErr: true,
			errAssertions: func(a *assert.Assertions, err error) {
				a.Contains(err.Error(), "validating State.infrastructure.azure.networkSecurityGroupName: must not be empty")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := validation.NewValidator()
			err := v.Validate(tc.stateFile(), validation.ValidateOptions{})
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
			v := validation.NewValidator()
			err := v.Validate(tc.stateFile(), validation.ValidateOptions{
				OverrideConstraints: tc.stateFile().preCreateConstraints,
			})
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := validation.NewValidator()
			err := v.Validate(tc.stateFile(), validation.ValidateOptions{
				OverrideConstraints: tc.stateFile().preInitConstraints,
			})
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
