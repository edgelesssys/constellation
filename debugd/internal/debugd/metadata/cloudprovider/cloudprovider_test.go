package cloudprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/deploy/ssh"
	"github.com/edgelesssys/constellation/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
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

func TestDiscoverLoadbalancerIP(t *testing.T) {
	ip := "192.0.2.1"
	endpoint := ip + ":1234"
	someErr := errors.New("failed")

	testCases := map[string]struct {
		metaAPI providerMetadata
		wantIP  string
		wantErr bool
	}{
		"discovery works": {
			metaAPI: &stubMetadata{getLBEndpointRes: endpoint},
			wantIP:  ip,
		},
		"get endpoint fails": {
			metaAPI: &stubMetadata{getLBEndpointErr: someErr},
			wantErr: true,
		},
		"invalid endpoint": {
			metaAPI: &stubMetadata{getLBEndpointRes: "invalid"},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fetcher := &Fetcher{
				metaAPI: tc.metaAPI,
			}

			ip, err := fetcher.DiscoverLoadbalancerIP(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantIP, ip)
			}
		})
	}
}

func TestFetchSSHKeys(t *testing.T) {
	err := errors.New("some err")

	testCases := map[string]struct {
		meta     stubMetadata
		wantKeys []ssh.UserKey
		wantErr  bool
	}{
		"fetch works": {
			meta: stubMetadata{
				selfRes: metadata.InstanceMetadata{
					Name:       "name",
					ProviderID: "provider-id",
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
			wantKeys: []ssh.UserKey{
				{
					Username:  "bob",
					PublicKey: "ssh-rsa bobskey",
				},
			},
		},
		"retrieve fails": {
			meta: stubMetadata{
				selfErr: err,
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
			keys, err := fetcher.FetchSSHKeys(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantKeys, keys)
		})
	}
}

type stubMetadata struct {
	listRes          []metadata.InstanceMetadata
	listErr          error
	selfRes          metadata.InstanceMetadata
	selfErr          error
	getInstanceRes   metadata.InstanceMetadata
	getInstanceErr   error
	getLBEndpointRes string
	getLBEndpointErr error
	supportedRes     bool
}

func (m *stubMetadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return m.listRes, m.listErr
}

func (m *stubMetadata) Self(ctx context.Context) (metadata.InstanceMetadata, error) {
	return m.selfRes, m.selfErr
}

func (m *stubMetadata) GetInstance(ctx context.Context, providerID string) (metadata.InstanceMetadata, error) {
	return m.getInstanceRes, m.getInstanceErr
}

func (m *stubMetadata) GetLoadBalancerEndpoint(ctx context.Context) (string, error) {
	return m.getLBEndpointRes, m.getLBEndpointErr
}

func (m *stubMetadata) Supported() bool {
	return m.supportedRes
}
