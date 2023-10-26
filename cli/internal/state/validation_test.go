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
		},
		// "invalid in-cluster endpoint": {
		// 	stateFile: func() *State {
		// 		s := defaultState()
		// 		s.Infrastructure.InClusterEndpoint = "invalid"
		// 		return s
		// 	},
		// 	wantErr: true,
		// 	errAssertions: func(a *assert.Assertions, err error) {
		// 		a.Contains(err.Error(), "validating State.version: invalid must be one of [v1]")
		// 	},
		// },
		// "empty sans": {
		// 	stateFile: func() *State {
		// 		s := defaultState()
		// 		s.Infrastructure.APIServerCertSANs = []string{}
		// 		return s
		// 	},
		// },
		// "not all sans are valid": {
		// 	stateFile: func() *State {
		// 		s := defaultState()
		// 		s.Infrastructure.APIServerCertSANs = []string{"not valid!"}
		// 		return s
		// 	},
		// 	wantErr: true,
		// 	errAssertions: func(a *assert.Assertions, err error) {
		// 		a.Contains(err.Error(), "tt")
		// 	},
		// },
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
