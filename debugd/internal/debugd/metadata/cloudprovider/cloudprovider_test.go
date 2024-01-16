/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestRole(t *testing.T) {
	instance1 := metadata.InstanceMetadata{Role: role.ControlPlane}
	instance2 := metadata.InstanceMetadata{Role: role.Worker}

	testCases := map[string]struct {
		meta     *stubMetadata
		wantErr  bool
		wantRole role.Role
	}{
		"control plane": {
			meta:     &stubMetadata{selfRes: instance1},
			wantRole: role.ControlPlane,
		},
		"worker": {
			meta:     &stubMetadata{selfRes: instance2},
			wantRole: role.Worker,
		},
		"self fails": {
			meta:     &stubMetadata{selfErr: errors.New("some err")},
			wantErr:  true,
			wantRole: role.Unknown,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fetcher := Fetcher{tc.meta}

			role, err := fetcher.Role(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantRole, role)
			}
		})
	}
}

func TestDiscoverDebugIPs(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta    stubMetadata
		wantIPs []string
		wantErr bool
	}{
		"disovery works": {
			meta: stubMetadata{
				listRes: []metadata.InstanceMetadata{
					{
						VPCIP: "192.0.2.0",
					},
					{
						VPCIP: "192.0.2.1",
					},
					{
						VPCIP: "192.0.2.2",
					},
				},
			},
			wantIPs: []string{
				"192.0.2.1", "192.0.2.2",
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				listErr: err,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fetcher := Fetcher{
				metaAPI: &tc.meta,
			}
			ips, err := fetcher.DiscoverDebugdIPs(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantIPs, ips)
		})
	}
}

func TestDiscoverLoadBalancerIP(t *testing.T) {
	ip := "192.0.2.1"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		metaAPI providerMetadata
		wantIP  string
		wantErr bool
	}{
		"discovery works": {
			metaAPI: &stubMetadata{getLBHostRes: ip},
			wantIP:  ip,
		},
		"get endpoint fails": {
			metaAPI: &stubMetadata{getLBEndpointErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fetcher := &Fetcher{
				metaAPI: tc.metaAPI,
			}

			ip, err := fetcher.DiscoverLoadBalancerIP(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantIP, ip)
			}
		})
	}
}

type stubMetadata struct {
	listRes                    []metadata.InstanceMetadata
	listErr                    error
	selfRes                    metadata.InstanceMetadata
	selfErr                    error
	getLBHostRes, getLBPortRes string
	getLBEndpointErr           error
	uid                        string
	uidErr                     error
}

func (m *stubMetadata) List(_ context.Context) ([]metadata.InstanceMetadata, error) {
	return m.listRes, m.listErr
}

func (m *stubMetadata) Self(_ context.Context) (metadata.InstanceMetadata, error) {
	return m.selfRes, m.selfErr
}

func (m *stubMetadata) GetLoadBalancerEndpoint(_ context.Context) (string, string, error) {
	return m.getLBHostRes, m.getLBPortRes, m.getLBEndpointErr
}

func (m *stubMetadata) UID(_ context.Context) (string, error) {
	return m.uid, m.uidErr
}
